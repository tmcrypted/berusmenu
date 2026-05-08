package store

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Pvz struct {
	ID              int64
	OwnerID         int64
	Status          commonv1.PvzStatus
	City            string
	Address         string
	CompanyName     string
	ContactPhone    string
	ContactTelegram *string
	Description     *string
	Lat             *float64
	Lng             *float64
	RejectionReason *string
	ModeratedBy     *int64
	ModeratedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PvzPatch struct {
	City            *string
	Address         *string
	CompanyName     *string
	ContactPhone    *string
	ContactTelegram *string
	Description     *string
	Lat             *float64
	Lng             *float64
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreatePvzWithEvent(ctx context.Context, pvz Pvz, actorUserID int64) (Pvz, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Pvz{}, err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("inventory.pvz").
		Columns("owner_id", "status", "city", "address", "company_name", "contact_phone", "contact_telegram", "description", "lat", "lng").
		Values(pvz.OwnerID, statusToDB(commonv1.PvzStatus_PVZ_STATUS_PENDING), pvz.City, pvz.Address, pvz.CompanyName, pvz.ContactPhone, pvz.ContactTelegram, pvz.Description, pvz.Lat, pvz.Lng).
		Suffix("RETURNING " + pvzColumns()).
		ToSql()
	if err != nil {
		return Pvz{}, err
	}
	created, err := scanPvz(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return Pvz{}, err
	}

	if err := insertModerationEvent(ctx, tx, created.ID, commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_SUBMITTED, actorUserID, nil); err != nil {
		return Pvz{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Pvz{}, err
	}
	return created, nil
}

func (s *Store) GetPvz(ctx context.Context, id int64) (Pvz, error) {
	sql, args, err := psql.Select(pvzColumns()).
		From("inventory.pvz").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return Pvz{}, err
	}
	return scanPvz(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) ListOwnerPvz(ctx context.Context, ownerID int64, limit uint64, offset uint64) ([]Pvz, error) {
	sql, args, err := psql.Select(pvzColumns()).
		From("inventory.pvz").
		Where(sq.Eq{"owner_id": ownerID}).
		OrderBy("id ASC").
		Limit(limit).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, err
	}
	return s.listPvz(ctx, sql, args...)
}

func (s *Store) ListPendingPvz(ctx context.Context, limit uint64, offset uint64) ([]Pvz, error) {
	sql, args, err := psql.Select(pvzColumns()).
		From("inventory.pvz").
		Where(sq.Eq{"status": statusToDB(commonv1.PvzStatus_PVZ_STATUS_PENDING)}).
		OrderBy("id ASC").
		Limit(limit).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, err
	}
	return s.listPvz(ctx, sql, args...)
}

func (s *Store) UpdatePvzWithEvent(ctx context.Context, id int64, patch PvzPatch, actorUserID int64) (Pvz, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Pvz{}, err
	}
	defer tx.Rollback(ctx)

	builder := psql.Update("inventory.pvz").
		Set("status", statusToDB(commonv1.PvzStatus_PVZ_STATUS_PENDING)).
		Set("rejection_reason", nil).
		Set("moderated_by", nil).
		Set("moderated_at", nil).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + pvzColumns())

	if patch.City != nil {
		builder = builder.Set("city", *patch.City)
	}
	if patch.Address != nil {
		builder = builder.Set("address", *patch.Address)
	}
	if patch.CompanyName != nil {
		builder = builder.Set("company_name", *patch.CompanyName)
	}
	if patch.ContactPhone != nil {
		builder = builder.Set("contact_phone", *patch.ContactPhone)
	}
	if patch.ContactTelegram != nil {
		builder = builder.Set("contact_telegram", patch.ContactTelegram)
	}
	if patch.Description != nil {
		builder = builder.Set("description", patch.Description)
	}
	if patch.Lat != nil {
		builder = builder.Set("lat", patch.Lat)
	}
	if patch.Lng != nil {
		builder = builder.Set("lng", patch.Lng)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return Pvz{}, err
	}
	updated, err := scanPvz(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return Pvz{}, err
	}
	if err := insertModerationEvent(ctx, tx, id, commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_SUBMITTED, actorUserID, nil); err != nil {
		return Pvz{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Pvz{}, err
	}
	return updated, nil
}

func (s *Store) ModeratePvzWithEvent(ctx context.Context, id int64, status commonv1.PvzStatus, actorUserID int64, reason *string, action commonv1.PvzModerationAction, comment *string) (Pvz, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Pvz{}, err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Update("inventory.pvz").
		Set("status", statusToDB(status)).
		Set("rejection_reason", reason).
		Set("moderated_by", actorUserID).
		Set("moderated_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + pvzColumns()).
		ToSql()
	if err != nil {
		return Pvz{}, err
	}
	updated, err := scanPvz(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return Pvz{}, err
	}
	if err := insertModerationEvent(ctx, tx, id, action, actorUserID, comment); err != nil {
		return Pvz{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Pvz{}, err
	}
	return updated, nil
}

func (s *Store) DeletePendingAndBlockOwnerPvz(ctx context.Context, ownerID int64, actorUserID int64, comment *string) (int64, int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	deleteSQL, deleteArgs, err := psql.Delete("inventory.pvz").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.Eq{"status": statusToDB(commonv1.PvzStatus_PVZ_STATUS_PENDING)}).
		ToSql()
	if err != nil {
		return 0, 0, err
	}
	deleteTag, err := tx.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		return 0, 0, err
	}

	updateSQL, updateArgs, err := psql.Update("inventory.pvz").
		Set("status", statusToDB(commonv1.PvzStatus_PVZ_STATUS_BLOCKED)).
		Set("rejection_reason", nil).
		Set("moderated_by", actorUserID).
		Set("moderated_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.NotEq{"status": []string{
			statusToDB(commonv1.PvzStatus_PVZ_STATUS_PENDING),
			statusToDB(commonv1.PvzStatus_PVZ_STATUS_BLOCKED),
		}}).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, 0, err
	}
	rows, err := tx.Query(ctx, updateSQL, updateArgs...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	blocked := int64(0)
	for rows.Next() {
		var pvzID int64
		if err := rows.Scan(&pvzID); err != nil {
			return 0, 0, err
		}
		if err := insertModerationEvent(ctx, tx, pvzID, commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_BLOCKED, actorUserID, comment); err != nil {
			return 0, 0, err
		}
		blocked++
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return deleteTag.RowsAffected(), blocked, nil
}

func (s *Store) listPvz(ctx context.Context, sql string, args ...any) ([]Pvz, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Pvz, 0)
	for rows.Next() {
		pvz, err := scanPvz(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, pvz)
	}
	return items, rows.Err()
}

func insertModerationEvent(ctx context.Context, tx pgx.Tx, pvzID int64, action commonv1.PvzModerationAction, actorUserID int64, comment *string) error {
	sql, args, err := psql.Insert("inventory.pvz_moderation_events").
		Columns("pvz_id", "action", "actor_user_id", "comment").
		Values(pvzID, actionToDB(action), actorUserID, comment).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}

func scanPvz(row pgx.Row) (Pvz, error) {
	var pvz Pvz
	var status string
	err := row.Scan(
		&pvz.ID,
		&pvz.OwnerID,
		&status,
		&pvz.City,
		&pvz.Address,
		&pvz.CompanyName,
		&pvz.ContactPhone,
		&pvz.ContactTelegram,
		&pvz.Description,
		&pvz.Lat,
		&pvz.Lng,
		&pvz.RejectionReason,
		&pvz.ModeratedBy,
		&pvz.ModeratedAt,
		&pvz.CreatedAt,
		&pvz.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Pvz{}, ErrNotFound
		}
		return Pvz{}, err
	}
	pvz.Status = statusFromDB(status)
	return pvz, nil
}

func pvzColumns() string {
	return "id, owner_id, status, city, address, company_name, contact_phone, contact_telegram, description, lat, lng, rejection_reason, moderated_by, moderated_at, created_at, updated_at"
}

func statusToDB(status commonv1.PvzStatus) string {
	switch status {
	case commonv1.PvzStatus_PVZ_STATUS_APPROVED:
		return "approved"
	case commonv1.PvzStatus_PVZ_STATUS_REJECTED:
		return "rejected"
	case commonv1.PvzStatus_PVZ_STATUS_BLOCKED:
		return "blocked"
	default:
		return "pending"
	}
}

func statusFromDB(status string) commonv1.PvzStatus {
	switch status {
	case "approved":
		return commonv1.PvzStatus_PVZ_STATUS_APPROVED
	case "rejected":
		return commonv1.PvzStatus_PVZ_STATUS_REJECTED
	case "blocked":
		return commonv1.PvzStatus_PVZ_STATUS_BLOCKED
	default:
		return commonv1.PvzStatus_PVZ_STATUS_PENDING
	}
}

func actionToDB(action commonv1.PvzModerationAction) string {
	switch action {
	case commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_APPROVED:
		return "approved"
	case commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_REJECTED:
		return "rejected"
	case commonv1.PvzModerationAction_PVZ_MODERATION_ACTION_BLOCKED:
		return "blocked"
	default:
		return "submitted"
	}
}
