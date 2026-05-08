package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/tmcrypted/berusmenu/inventory-service/internal/dto"
	"github.com/tmcrypted/berusmenu/inventory-service/internal/store"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

const (
	defaultPageSize = uint64(20)
	maxPageSize     = uint64(100)
)

type PvzRepository interface {
	CreatePvzWithEvent(ctx context.Context, pvz store.Pvz, actorUserID int64) (store.Pvz, error)
	GetPvz(ctx context.Context, id int64) (store.Pvz, error)
	ListOwnerPvz(ctx context.Context, ownerID int64, limit uint64, offset uint64) ([]store.Pvz, error)
	ListPendingPvz(ctx context.Context, limit uint64, offset uint64) ([]store.Pvz, error)
	UpdatePvzWithEvent(ctx context.Context, id int64, patch store.PvzPatch, actorUserID int64) (store.Pvz, error)
	ModeratePvzWithEvent(ctx context.Context, id int64, status commonv1.PvzStatus, actorUserID int64, reason *string, action commonv1.PvzModerationAction, comment *string) (store.Pvz, error)
	DeletePendingAndBlockOwnerPvz(ctx context.Context, ownerID int64, actorUserID int64, comment *string) (int64, int64, error)
}

type ActiveVacancyChecker interface {
	HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error)
}

type Service struct {
	pvz             PvzRepository
	activeVacancies ActiveVacancyChecker
}

func New(pvz PvzRepository, activeVacancies ActiveVacancyChecker) *Service {
	return &Service{pvz: pvz, activeVacancies: activeVacancies}
}

func (s *Service) CreatePvz(ctx context.Context, input dto.CreatePvzInput) (store.Pvz, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_OWNER); err != nil {
		return store.Pvz{}, err
	}
	city, err := required(input.City, "city")
	if err != nil {
		return store.Pvz{}, err
	}
	address, err := required(input.Address, "address")
	if err != nil {
		return store.Pvz{}, err
	}
	companyName, err := required(input.CompanyName, "company_name")
	if err != nil {
		return store.Pvz{}, err
	}
	contactPhone, err := required(input.ContactPhone, "contact_phone")
	if err != nil {
		return store.Pvz{}, err
	}

	pvz, err := s.pvz.CreatePvzWithEvent(ctx, store.Pvz{
		OwnerID:         input.Actor.UserID,
		City:            city,
		Address:         address,
		CompanyName:     companyName,
		ContactPhone:    contactPhone,
		ContactTelegram: trimOptional(input.ContactTelegram),
		Description:     trimOptional(input.Description),
		Lat:             input.Lat,
		Lng:             input.Lng,
	}, input.Actor.UserID)
	if err != nil {
		return store.Pvz{}, serviceError(ErrInternal, "failed to create pvz")
	}
	return pvz, nil
}

func (s *Service) GetPvz(ctx context.Context, input dto.GetPvzInput) (store.Pvz, error) {
	pvz, err := s.pvz.GetPvz(ctx, input.ID)
	if err != nil {
		return store.Pvz{}, notFoundOrInternal(err, "pvz not found", "failed to get pvz")
	}
	if err := canReadPvz(input.Actor, pvz); err != nil {
		return store.Pvz{}, err
	}
	return pvz, nil
}

func (s *Service) ListOwnerPvz(ctx context.Context, input dto.ListOwnerPvzInput) ([]store.Pvz, *string, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_OWNER); err != nil {
		return nil, nil, err
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.pvz.ListOwnerPvz(ctx, input.Actor.UserID, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list pvz")
	}
	return trimPage(items, page)
}

func (s *Service) UpdatePvz(ctx context.Context, input dto.UpdatePvzInput) (store.Pvz, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_OWNER); err != nil {
		return store.Pvz{}, err
	}
	current, err := s.pvz.GetPvz(ctx, input.ID)
	if err != nil {
		return store.Pvz{}, notFoundOrInternal(err, "pvz not found", "failed to get pvz")
	}
	if current.OwnerID != input.Actor.UserID {
		return store.Pvz{}, serviceError(ErrPermissionDenied, "cannot update another owner's pvz")
	}
	patch, err := updatePatch(input)
	if err != nil {
		return store.Pvz{}, err
	}

	if s.activeVacancies == nil {
		return store.Pvz{}, serviceError(ErrUnavailable, "active vacancy checker is not configured")
	}
	hasActive, err := s.activeVacancies.HasActiveVacancyForPvz(ctx, input.ID)
	if err != nil {
		return store.Pvz{}, serviceError(ErrUnavailable, "failed to check active vacancies for pvz")
	}
	if hasActive {
		return store.Pvz{}, serviceError(ErrFailedPrecondition, "pvz has active vacancies")
	}

	updated, err := s.pvz.UpdatePvzWithEvent(ctx, input.ID, patch, input.Actor.UserID)
	if err != nil {
		return store.Pvz{}, notFoundOrInternal(err, "pvz not found", "failed to update pvz")
	}
	return updated, nil
}

func (s *Service) CheckPvzForVacancy(ctx context.Context, input dto.CheckPvzForVacancyInput) (store.Pvz, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_OWNER); err != nil {
		return store.Pvz{}, err
	}
	pvz, err := s.pvz.GetPvz(ctx, input.PvzID)
	if err != nil {
		return store.Pvz{}, notFoundOrInternal(err, "pvz not found", "failed to get pvz")
	}
	if pvz.OwnerID != input.Actor.UserID {
		return store.Pvz{}, serviceError(ErrPermissionDenied, "cannot create vacancy for another owner's pvz")
	}
	if pvz.Status != commonv1.PvzStatus_PVZ_STATUS_APPROVED {
		return store.Pvz{}, serviceError(ErrFailedPrecondition, "pvz is not approved")
	}
	return pvz, nil
}

func (s *Service) ListPendingPvz(ctx context.Context, input dto.ListPendingPvzInput) ([]store.Pvz, *string, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_MODERATOR); err != nil {
		return nil, nil, err
	}
	page, err := pageFrom(input.PageSize, input.PageToken)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.pvz.ListPendingPvz(ctx, page.Limit+1, page.Offset)
	if err != nil {
		return nil, nil, serviceError(ErrInternal, "failed to list pending pvz")
	}
	return trimPage(items, page)
}

func (s *Service) ApprovePvz(ctx context.Context, input dto.ModeratePvzInput) (store.Pvz, error) {
	return s.moderate(ctx, input, commonv1.PvzStatus_PVZ_STATUS_APPROVED, nil, commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_APPROVED)
}

func (s *Service) RejectPvz(ctx context.Context, input dto.ModeratePvzInput) (store.Pvz, error) {
	reason, err := required(deref(input.Reason), "reason")
	if err != nil {
		return store.Pvz{}, err
	}
	return s.moderate(ctx, input, commonv1.PvzStatus_PVZ_STATUS_REJECTED, &reason, commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_REJECTED)
}

func (s *Service) BlockPvz(ctx context.Context, input dto.ModeratePvzInput) (store.Pvz, error) {
	return s.moderate(ctx, input, commonv1.PvzStatus_PVZ_STATUS_BLOCKED, nil, commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_BLOCKED)
}

func (s *Service) DeleteOwnerPvzRequests(ctx context.Context, input dto.DeleteOwnerPvzRequestsInput) (int64, int64, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_MODERATOR); err != nil {
		return 0, 0, err
	}
	if input.OwnerID <= 0 {
		return 0, 0, serviceError(ErrInvalidArgument, "owner_id is required")
	}
	deleted, blocked, err := s.pvz.DeletePendingAndBlockOwnerPvz(ctx, input.OwnerID, input.Actor.UserID, trimOptional(input.Comment))
	if err != nil {
		return 0, 0, serviceError(ErrInternal, "failed to delete owner pvz requests")
	}
	return deleted, blocked, nil
}

func (s *Service) moderate(ctx context.Context, input dto.ModeratePvzInput, status commonv1.PvzStatus, reason *string, action commonv1.PvzModerationAction) (store.Pvz, error) {
	if err := requireRole(input.Actor, commonv1.UserRole_USER_ROLE_MODERATOR); err != nil {
		return store.Pvz{}, err
	}
	updated, err := s.pvz.ModeratePvzWithEvent(ctx, input.PvzID, status, input.Actor.UserID, reason, action, trimOptional(input.Comment))
	if err != nil {
		return store.Pvz{}, notFoundOrInternal(err, "pvz not found", "failed to moderate pvz")
	}
	return updated, nil
}

func requireRole(actor dto.Actor, role commonv1.UserRole) error {
	if actor.UserID <= 0 {
		return serviceError(ErrUnauthenticated, "missing authenticated user")
	}
	if actor.Role != role {
		return serviceError(ErrPermissionDenied, "insufficient role")
	}
	return nil
}

func canReadPvz(actor dto.Actor, pvz store.Pvz) error {
	if actor.UserID <= 0 {
		return serviceError(ErrUnauthenticated, "missing authenticated user")
	}
	if actor.Role == commonv1.UserRole_USER_ROLE_MODERATOR {
		return nil
	}
	if actor.Role == commonv1.UserRole_USER_ROLE_OWNER && actor.UserID == pvz.OwnerID {
		return nil
	}
	return serviceError(ErrPermissionDenied, "cannot access pvz")
}

func updatePatch(input dto.UpdatePvzInput) (store.PvzPatch, error) {
	if input.City == nil && input.Address == nil && input.CompanyName == nil && input.ContactPhone == nil &&
		input.ContactTelegram == nil && input.Description == nil && input.Lat == nil && input.Lng == nil {
		return store.PvzPatch{}, serviceError(ErrInvalidArgument, "at least one field is required")
	}

	patch := store.PvzPatch{
		ContactTelegram: trimOptional(input.ContactTelegram),
		Description:     trimOptional(input.Description),
		Lat:             input.Lat,
		Lng:             input.Lng,
	}
	var err error
	if input.City != nil {
		value, valueErr := required(*input.City, "city")
		if valueErr != nil {
			err = valueErr
		}
		patch.City = &value
	}
	if err == nil && input.Address != nil {
		value, valueErr := required(*input.Address, "address")
		if valueErr != nil {
			err = valueErr
		}
		patch.Address = &value
	}
	if err == nil && input.CompanyName != nil {
		value, valueErr := required(*input.CompanyName, "company_name")
		if valueErr != nil {
			err = valueErr
		}
		patch.CompanyName = &value
	}
	if err == nil && input.ContactPhone != nil {
		value, valueErr := required(*input.ContactPhone, "contact_phone")
		if valueErr != nil {
			err = valueErr
		}
		patch.ContactPhone = &value
	}
	if err != nil {
		return store.PvzPatch{}, err
	}
	return patch, nil
}

func required(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", serviceError(ErrInvalidArgument, field+" is required")
	}
	return trimmed, nil
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func pageFrom(pageSize int32, token *string) (dto.Page, error) {
	limit := defaultPageSize
	if pageSize > 0 {
		limit = uint64(pageSize)
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	var offset uint64
	if token != nil && strings.TrimSpace(*token) != "" {
		parsed, err := strconv.ParseUint(strings.TrimSpace(*token), 10, 64)
		if err != nil {
			return dto.Page{}, serviceError(ErrInvalidArgument, "invalid page token")
		}
		offset = parsed
	}
	return dto.Page{Limit: limit, Offset: offset}, nil
}

func trimPage(items []store.Pvz, page dto.Page) ([]store.Pvz, *string, error) {
	if uint64(len(items)) <= page.Limit {
		return items, nil, nil
	}
	trimmed := items[:page.Limit]
	next := strconv.FormatUint(page.Offset+page.Limit, 10)
	return trimmed, &next, nil
}

func notFoundOrInternal(err error, notFoundMessage string, internalMessage string) error {
	if errors.Is(err, store.ErrNotFound) {
		return serviceError(ErrNotFound, notFoundMessage)
	}
	return serviceError(ErrInternal, internalMessage)
}
