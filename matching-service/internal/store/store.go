package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

type Store struct {
	pool *pgxpool.Pool
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func NewEventID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *Store) CreateVacancy(ctx context.Context, input Vacancy) (Vacancy, error) {
	sql, args, err := psql.Insert("matching.vacancies").
		Columns("owner_id", "pvz_id", "status", "title", "description", "shift_start_at", "shift_end_at", "pvz_name_snapshot", "pvz_address_snapshot", "pvz_city_snapshot", "pay_amount", "slots_total").
		Values(input.OwnerID, input.PvzID, vacancyStatusToDB(commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED), input.Title, input.Description, input.ShiftStartAt, input.ShiftEndAt, input.PvzNameSnapshot, input.PvzAddressSnapshot, input.PvzCitySnapshot, input.PayAmount, input.SlotsTotal).
		Suffix("RETURNING " + vacancyColumns()).
		ToSql()
	if err != nil {
		return Vacancy{}, err
	}
	return scanVacancy(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) GetVacancy(ctx context.Context, id int64) (Vacancy, error) {
	sql, args, err := psql.Select(vacancyColumns()).From("matching.vacancies").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return Vacancy{}, err
	}
	return scanVacancy(s.pool.QueryRow(ctx, sql, args...))
}

type ListVacancyFilter struct {
	ActorUserID    int64
	Role           commonv1.UserRole
	PvzID          *int64
	ShiftStartFrom *time.Time
	ShiftStartTo   *time.Time
	Status         commonv1.VacancyStatus
	Limit          uint64
	Offset         uint64
}

func (s *Store) ListVacancies(ctx context.Context, filter ListVacancyFilter) ([]Vacancy, error) {
	builder := psql.Select(vacancyColumns()).From("matching.vacancies")
	switch filter.Role {
	case commonv1.UserRole_USER_ROLE_OWNER:
		builder = builder.Where(sq.Eq{"owner_id": filter.ActorUserID})
	case commonv1.UserRole_USER_ROLE_WORKER:
		builder = builder.Where(sq.Eq{"status": vacancyStatusToDB(commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED)})
	case commonv1.UserRole_USER_ROLE_MODERATOR:
	default:
		builder = builder.Where("1 = 0")
	}
	if filter.PvzID != nil {
		builder = builder.Where(sq.Eq{"pvz_id": *filter.PvzID})
	}
	if filter.Status != commonv1.VacancyStatus_VACANCY_STATUS_UNSPECIFIED && filter.Role != commonv1.UserRole_USER_ROLE_WORKER {
		builder = builder.Where(sq.Eq{"status": vacancyStatusToDB(filter.Status)})
	}
	if filter.ShiftStartFrom != nil {
		builder = builder.Where(sq.GtOrEq{"shift_start_at": *filter.ShiftStartFrom})
	}
	if filter.ShiftStartTo != nil {
		builder = builder.Where(sq.LtOrEq{"shift_start_at": *filter.ShiftStartTo})
	}
	sql, args, err := builder.OrderBy("shift_start_at ASC", "id ASC").Limit(filter.Limit).Offset(filter.Offset).ToSql()
	if err != nil {
		return nil, err
	}
	return list(ctx, s.pool, sql, scanVacancy, args...)
}

func (s *Store) CreateResponseWithOutbox(ctx context.Context, response Response, buildOutbox func(Response) (OutboxEventInput, error)) (Response, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Response{}, err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("matching.responses").
		Columns("vacancy_id", "worker_id", "status", "message").
		Values(response.VacancyID, response.WorkerID, responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_PENDING), response.Message).
		Suffix("RETURNING " + responseColumns()).
		ToSql()
	if err != nil {
		return Response{}, err
	}
	created, err := scanResponse(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		if isUniqueViolation(err) {
			return Response{}, ErrAlreadyExists
		}
		return Response{}, err
	}
	outbox, err := buildOutbox(created)
	if err != nil {
		return Response{}, err
	}
	if err := enqueueOutboxEvent(ctx, tx, outbox); err != nil {
		return Response{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Response{}, err
	}
	return created, nil
}

func (s *Store) GetResponse(ctx context.Context, id int64) (Response, error) {
	sql, args, err := psql.Select(responseColumns()).From("matching.responses").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return Response{}, err
	}
	return scanResponse(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) ListResponsesForVacancy(ctx context.Context, vacancyID int64, limit uint64, offset uint64) ([]Response, error) {
	sql, args, err := psql.Select(responseColumns()).From("matching.responses").
		Where(sq.Eq{"vacancy_id": vacancyID}).OrderBy("id ASC").Limit(limit).Offset(offset).ToSql()
	if err != nil {
		return nil, err
	}
	return list(ctx, s.pool, sql, scanResponse, args...)
}

func (s *Store) ListResponsesByWorker(ctx context.Context, workerID int64, limit uint64, offset uint64) ([]Response, error) {
	sql, args, err := psql.Select(responseColumns()).From("matching.responses").
		Where(sq.Eq{"worker_id": workerID}).OrderBy("created_at DESC", "id DESC").Limit(limit).Offset(offset).ToSql()
	if err != nil {
		return nil, err
	}
	return list(ctx, s.pool, sql, scanResponse, args...)
}

func (s *Store) WithdrawResponse(ctx context.Context, responseID int64) (Response, error) {
	sql, args, err := psql.Update("matching.responses").
		Set("status", responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_WITHDRAWN)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": responseID, "status": responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_PENDING)}).
		Suffix("RETURNING " + responseColumns()).
		ToSql()
	if err != nil {
		return Response{}, err
	}
	return scanResponse(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) AssignmentForResponse(ctx context.Context, responseID int64) (Assignment, error) {
	sql, args, err := psql.Select(assignmentColumns()).From("matching.assignments").Where(sq.Eq{"response_id": responseID}).ToSql()
	if err != nil {
		return Assignment{}, err
	}
	return scanAssignment(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) AcceptResponse(ctx context.Context, input AssignmentSnapshotInput, admin *AdminEventInput) (Assignment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Assignment{}, err
	}
	defer tx.Rollback(ctx)

	updateResponseSQL, updateResponseArgs, err := psql.Update("matching.responses").
		Set("status", responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_ACCEPTED)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": input.ResponseID, "status": responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_PENDING)}).
		Suffix("RETURNING " + responseColumns()).
		ToSql()
	if err != nil {
		return Assignment{}, err
	}
	accepted, err := scanResponse(tx.QueryRow(ctx, updateResponseSQL, updateResponseArgs...))
	if err != nil {
		return Assignment{}, err
	}

	rejectSQL, rejectArgs, err := psql.Update("matching.responses").
		Set("status", responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_REJECTED)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"vacancy_id": input.Vacancy.ID, "status": responseStatusToDB(commonv1.ResponseStatus_RESPONSE_STATUS_PENDING)}).
		Where(sq.NotEq{"id": input.ResponseID}).
		ToSql()
	if err != nil {
		return Assignment{}, err
	}
	if _, err := tx.Exec(ctx, rejectSQL, rejectArgs...); err != nil {
		return Assignment{}, err
	}

	vacancySQL, vacancyArgs, err := psql.Update("matching.vacancies").
		Set("status", vacancyStatusToDB(commonv1.VacancyStatus_VACANCY_STATUS_ASSIGNED)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": input.Vacancy.ID, "status": vacancyStatusToDB(commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED)}).
		ToSql()
	if err != nil {
		return Assignment{}, err
	}
	if tag, err := tx.Exec(ctx, vacancySQL, vacancyArgs...); err != nil {
		return Assignment{}, err
	} else if tag.RowsAffected() == 0 {
		return Assignment{}, ErrNotFound
	}

	sql, args, err := psql.Insert("matching.assignments").
		Columns("vacancy_id", "response_id", "owner_id", "worker_id", "status", "pvz_name_snapshot", "pvz_address_snapshot", "pvz_city_snapshot", "shift_start_at_snapshot", "pay_amount_snapshot", "shift_end_at_snapshot", "owner_name_snapshot", "worker_name_snapshot").
		Values(input.Vacancy.ID, accepted.ID, input.Vacancy.OwnerID, input.WorkerID, assignmentStatusToDB(commonv1.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE), input.PvzNameSnapshot, input.PvzAddressSnapshot, input.PvzCitySnapshot, input.Vacancy.ShiftStartAt, input.Vacancy.PayAmount, input.Vacancy.ShiftEndAt, input.OwnerNameSnapshot, input.WorkerNameSnapshot).
		Suffix("RETURNING " + assignmentColumns()).
		ToSql()
	if err != nil {
		return Assignment{}, err
	}
	assignment, err := scanAssignment(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return Assignment{}, err
	}
	if admin != nil {
		if err := insertAdminEvent(ctx, tx, *admin); err != nil {
			return Assignment{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Assignment{}, err
	}
	return assignment, nil
}

func (s *Store) GetAssignment(ctx context.Context, id int64) (Assignment, error) {
	sql, args, err := psql.Select(assignmentColumns()).From("matching.assignments").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return Assignment{}, err
	}
	return scanAssignment(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) ListAssignmentsForUser(ctx context.Context, userID int64, limit uint64, offset uint64) ([]Assignment, error) {
	sql, args, err := psql.Select(assignmentColumns()).From("matching.assignments").
		Where(sq.Or{sq.Eq{"owner_id": userID}, sq.Eq{"worker_id": userID}}).
		OrderBy("created_at DESC", "id DESC").
		Limit(limit).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, err
	}
	return list(ctx, s.pool, sql, scanAssignment, args...)
}

func (s *Store) AssignmentForVacancy(ctx context.Context, vacancyID int64) (Assignment, error) {
	sql, args, err := psql.Select(assignmentColumns()).From("matching.assignments").Where(sq.Eq{"vacancy_id": vacancyID}).ToSql()
	if err != nil {
		return Assignment{}, err
	}
	return scanAssignment(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) FinalizeAssignment(ctx context.Context, id int64, status commonv1.AssignmentStatus, cancelledBy *int64, reason *string, admin *AdminEventInput) (Assignment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Assignment{}, err
	}
	defer tx.Rollback(ctx)

	builder := psql.Update("matching.assignments").
		Set("status", assignmentStatusToDB(status)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "status": assignmentStatusToDB(commonv1.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE)}).
		Suffix("RETURNING " + assignmentColumns())
	if status == commonv1.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED {
		builder = builder.Set("completed_at", sq.Expr("NOW()"))
	} else {
		builder = builder.Set("cancelled_at", sq.Expr("NOW()")).Set("cancelled_by_user_id", cancelledBy).Set("cancel_reason", reason)
	}
	sql, args, err := builder.ToSql()
	if err != nil {
		return Assignment{}, err
	}
	assignment, err := scanAssignment(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return Assignment{}, err
	}
	vacancyStatus := commonv1.VacancyStatus_VACANCY_STATUS_COMPLETED
	if status == commonv1.AssignmentStatus_ASSIGNMENT_STATUS_CANCELLED {
		vacancyStatus = commonv1.VacancyStatus_VACANCY_STATUS_CANCELLED
	}
	vacancySQL, vacancyArgs, err := psql.Update("matching.vacancies").
		Set("status", vacancyStatusToDB(vacancyStatus)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": assignment.VacancyID}).
		ToSql()
	if err != nil {
		return Assignment{}, err
	}
	if _, err := tx.Exec(ctx, vacancySQL, vacancyArgs...); err != nil {
		return Assignment{}, err
	}
	if admin != nil {
		if err := insertAdminEvent(ctx, tx, *admin); err != nil {
			return Assignment{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Assignment{}, err
	}
	return assignment, nil
}

func (s *Store) CreateReviewAndRecalculateRating(ctx context.Context, review Review) (Review, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Review{}, err
	}
	defer tx.Rollback(ctx)
	created, err := insertReview(ctx, tx, review)
	if err != nil {
		return Review{}, err
	}
	if err := recalculateRating(ctx, tx, created.TargetUserID); err != nil {
		return Review{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Review{}, err
	}
	return created, nil
}

func (s *Store) UpdateReviewAndRecalculateRating(ctx context.Context, id int64, rating *int32, body *string) (Review, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Review{}, err
	}
	defer tx.Rollback(ctx)
	builder := psql.Update("matching.reviews").Set("updated_at", sq.Expr("NOW()")).Where(sq.Eq{"id": id}).Suffix("RETURNING " + reviewColumns())
	if rating != nil {
		builder = builder.Set("rating", *rating)
	}
	if body != nil {
		builder = builder.Set("body", body)
	}
	sql, args, err := builder.ToSql()
	if err != nil {
		return Review{}, err
	}
	updated, err := scanReview(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return Review{}, err
	}
	if err := recalculateRating(ctx, tx, updated.TargetUserID); err != nil {
		return Review{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Review{}, err
	}
	return updated, nil
}

func (s *Store) GetReview(ctx context.Context, id int64) (Review, error) {
	sql, args, err := psql.Select(reviewColumns()).From("matching.reviews").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return Review{}, err
	}
	return scanReview(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) ListReviews(ctx context.Context, targetUserID *int64, limit uint64, offset uint64) ([]Review, error) {
	builder := psql.Select(reviewColumns()).From("matching.reviews")
	if targetUserID != nil {
		builder = builder.Where(sq.Eq{"target_user_id": *targetUserID})
	}
	sql, args, err := builder.OrderBy("created_at DESC", "id DESC").Limit(limit).Offset(offset).ToSql()
	if err != nil {
		return nil, err
	}
	return list(ctx, s.pool, sql, scanReview, args...)
}

func (s *Store) ListReviewReplies(ctx context.Context, reviewID int64, limit uint64, offset uint64) ([]ReviewReply, error) {
	sql, args, err := psql.Select(reviewReplyColumns()).
		From("matching.review_replies").
		Where(sq.Eq{"review_id": reviewID}).
		OrderBy("created_at ASC", "id ASC").
		Limit(limit).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, err
	}
	return list(ctx, s.pool, sql, scanReviewReply, args...)
}

func (s *Store) GetUserRating(ctx context.Context, userID int64) (UserRating, error) {
	sql, args, err := psql.Select("user_id", "avg_rating", "reviews_count", "updated_at").
		From("matching.user_ratings").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return UserRating{}, err
	}
	var rating UserRating
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&rating.UserID, &rating.AvgRating, &rating.ReviewsCount, &rating.UpdatedAt)
	if err != nil {
		return UserRating{}, rowErr(err)
	}
	return rating, nil
}

func (s *Store) CreateReviewReply(ctx context.Context, reply ReviewReply) (ReviewReply, error) {
	sql, args, err := psql.Insert("matching.review_replies").
		Columns("review_id", "author_user_id", "target_user_id", "body").
		Values(reply.ReviewID, reply.AuthorUserID, reply.TargetUserID, reply.Body).
		Suffix("RETURNING " + reviewReplyColumns()).ToSql()
	if err != nil {
		return ReviewReply{}, err
	}
	return scanReviewReply(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) GetReviewReply(ctx context.Context, id int64) (ReviewReply, error) {
	sql, args, err := psql.Select(reviewReplyColumns()).From("matching.review_replies").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return ReviewReply{}, err
	}
	return scanReviewReply(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) UpdateReviewReply(ctx context.Context, id int64, body string) (ReviewReply, error) {
	sql, args, err := psql.Update("matching.review_replies").Set("body", body).Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).Suffix("RETURNING " + reviewReplyColumns()).ToSql()
	if err != nil {
		return ReviewReply{}, err
	}
	return scanReviewReply(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) HasActiveVacancyForPvz(ctx context.Context, pvzID int64) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM matching.vacancies
		WHERE pvz_id = $1 AND status IN ('published', 'assigned')
	)`, pvzID).Scan(&exists)
	return exists, err
}

func (s *Store) ClaimPendingOutboxEvents(ctx context.Context, limit int, claimTimeout time.Duration) ([]OutboxEvent, error) {
	if limit < 1 {
		limit = 1
	}
	if claimTimeout <= 0 {
		claimTimeout = 5 * time.Minute
	}
	rows, err := s.pool.Query(ctx, `
UPDATE matching.outbox_events
SET status = 'processing',
    available_at = NOW() + ($2 * INTERVAL '1 second')
WHERE id IN (
	SELECT id FROM matching.outbox_events
	WHERE status IN ('pending', 'processing') AND available_at <= NOW()
	ORDER BY id
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
RETURNING id, event_id, topic, event_type, event_key, payload, attempts`, limit, int64(claimTimeout.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]OutboxEvent, 0, limit)
	for rows.Next() {
		var event OutboxEvent
		if err := rows.Scan(&event.ID, &event.EventID, &event.Topic, &event.EventType, &event.EventKey, &event.Payload, &event.Attempts); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) MarkOutboxEventPublished(ctx context.Context, id int64) error {
	sql, args, err := psql.Update("matching.outbox_events").Set("status", "published").Set("published_at", sq.Expr("NOW()")).Set("last_error", nil).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *Store) MarkOutboxEventFailed(ctx context.Context, id int64, cause error, retryBackoff time.Duration) error {
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	sql, args, err := psql.Update("matching.outbox_events").Set("status", "pending").Set("attempts", sq.Expr("attempts + 1")).Set("available_at", time.Now().Add(retryBackoff)).Set("last_error", message).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *Store) CreateAdminEvent(ctx context.Context, input AdminEventInput) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := insertAdminEvent(ctx, tx, input); err != nil {
		return err
	}
	err = tx.Commit(ctx)
	return err
}

func insertAdminEvent(ctx context.Context, tx pgx.Tx, input AdminEventInput) error {
	sql, args, err := psql.Insert("matching.admin_events").
		Columns("actor_user_id", "target_type", "target_id", "action", "previous_state", "new_state", "reason", "comment").
		Values(input.ActorUserID, input.TargetType, input.TargetID, input.Action, input.PreviousState, input.NewState, input.Reason, input.Comment).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}

func insertReview(ctx context.Context, tx pgx.Tx, review Review) (Review, error) {
	sql, args, err := psql.Insert("matching.reviews").
		Columns("assignment_id", "author_user_id", "target_user_id", "target_role", "rating", "body").
		Values(review.AssignmentID, review.AuthorUserID, review.TargetUserID, reviewTargetRoleToDB(review.TargetRole), review.Rating, review.Body).
		Suffix("RETURNING " + reviewColumns()).ToSql()
	if err != nil {
		return Review{}, err
	}
	created, err := scanReview(tx.QueryRow(ctx, sql, args...))
	if err != nil && isUniqueViolation(err) {
		return Review{}, ErrAlreadyExists
	}
	return created, err
}

func recalculateRating(ctx context.Context, tx pgx.Tx, userID int64) error {
	_, err := tx.Exec(ctx, `
INSERT INTO matching.user_ratings (user_id, avg_rating, reviews_count, updated_at)
SELECT $1, COALESCE(AVG(rating), 0), COUNT(*), NOW()
FROM matching.reviews
WHERE target_user_id = $1
ON CONFLICT (user_id) DO UPDATE
SET avg_rating = EXCLUDED.avg_rating,
    reviews_count = EXCLUDED.reviews_count,
    updated_at = NOW()`, userID)
	return err
}

func enqueueOutboxEvent(ctx context.Context, tx pgx.Tx, input OutboxEventInput) error {
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		var err error
		eventID, err = NewEventID()
		if err != nil {
			return err
		}
	}
	sql, args, err := psql.Insert("matching.outbox_events").Columns("event_id", "topic", "event_type", "event_key", "payload").Values(eventID, input.Topic, input.EventType, input.EventKey, input.Payload).ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func list[T any](ctx context.Context, q queryer, sql string, scan func(pgx.Row) (T, error), args ...any) ([]T, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]T, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func scanVacancy(row pgx.Row) (Vacancy, error) {
	var item Vacancy
	var status string
	err := row.Scan(&item.ID, &item.OwnerID, &item.PvzID, &status, &item.Title, &item.Description, &item.ShiftStartAt, &item.ShiftEndAt, &item.PvzNameSnapshot, &item.PvzAddressSnapshot, &item.PvzCitySnapshot, &item.PayAmount, &item.SlotsTotal, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Vacancy{}, rowErr(err)
	}
	item.Status = vacancyStatusFromDB(status)
	return item, nil
}

func scanResponse(row pgx.Row) (Response, error) {
	var item Response
	var status string
	err := row.Scan(&item.ID, &item.VacancyID, &item.WorkerID, &status, &item.Message, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Response{}, rowErr(err)
	}
	item.Status = responseStatusFromDB(status)
	return item, nil
}

func scanAssignment(row pgx.Row) (Assignment, error) {
	var item Assignment
	var status string
	err := row.Scan(&item.ID, &item.VacancyID, &item.ResponseID, &item.OwnerID, &item.WorkerID, &status, &item.AcceptedAt, &item.PvzNameSnapshot, &item.PvzAddressSnapshot, &item.PvzCitySnapshot, &item.ShiftStartAtSnapshot, &item.PayAmountSnapshot, &item.ShiftEndAtSnapshot, &item.OwnerNameSnapshot, &item.WorkerNameSnapshot, &item.CompletedAt, &item.CancelledAt, &item.CancelledByUserID, &item.CancelReason, &item.UpdatedAt, &item.CreatedAt)
	if err != nil {
		return Assignment{}, rowErr(err)
	}
	item.Status = assignmentStatusFromDB(status)
	return item, nil
}

func scanReview(row pgx.Row) (Review, error) {
	var item Review
	var targetRole string
	err := row.Scan(&item.ID, &item.AssignmentID, &item.AuthorUserID, &item.TargetUserID, &targetRole, &item.Rating, &item.Body, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Review{}, rowErr(err)
	}
	item.TargetRole = reviewTargetRoleFromDB(targetRole)
	return item, nil
}

func scanReviewReply(row pgx.Row) (ReviewReply, error) {
	var item ReviewReply
	err := row.Scan(&item.ID, &item.ReviewID, &item.AuthorUserID, &item.TargetUserID, &item.Body, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return ReviewReply{}, rowErr(err)
	}
	return item, nil
}

func rowErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func vacancyColumns() string {
	return "id, owner_id, pvz_id, status, title, description, shift_start_at, shift_end_at, pvz_name_snapshot, pvz_address_snapshot, pvz_city_snapshot, pay_amount, slots_total, created_at, updated_at"
}

func responseColumns() string {
	return "id, vacancy_id, worker_id, status, message, created_at, updated_at"
}

func assignmentColumns() string {
	return "id, vacancy_id, response_id, owner_id, worker_id, status, accepted_at, pvz_name_snapshot, pvz_address_snapshot, pvz_city_snapshot, shift_start_at_snapshot, pay_amount_snapshot, shift_end_at_snapshot, owner_name_snapshot, worker_name_snapshot, completed_at, cancelled_at, cancelled_by_user_id, cancel_reason, updated_at, created_at"
}

func reviewColumns() string {
	return "id, assignment_id, author_user_id, target_user_id, target_role, rating, body, created_at, updated_at"
}

func reviewReplyColumns() string {
	return "id, review_id, author_user_id, target_user_id, body, created_at, updated_at"
}

func vacancyStatusToDB(status commonv1.VacancyStatus) string {
	switch status {
	case commonv1.VacancyStatus_VACANCY_STATUS_ASSIGNED:
		return "assigned"
	case commonv1.VacancyStatus_VACANCY_STATUS_COMPLETED:
		return "completed"
	case commonv1.VacancyStatus_VACANCY_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "published"
	}
}

func vacancyStatusFromDB(status string) commonv1.VacancyStatus {
	switch status {
	case "assigned":
		return commonv1.VacancyStatus_VACANCY_STATUS_ASSIGNED
	case "completed":
		return commonv1.VacancyStatus_VACANCY_STATUS_COMPLETED
	case "cancelled":
		return commonv1.VacancyStatus_VACANCY_STATUS_CANCELLED
	default:
		return commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED
	}
}

func responseStatusToDB(status commonv1.ResponseStatus) string {
	switch status {
	case commonv1.ResponseStatus_RESPONSE_STATUS_ACCEPTED:
		return "accepted"
	case commonv1.ResponseStatus_RESPONSE_STATUS_REJECTED:
		return "rejected"
	case commonv1.ResponseStatus_RESPONSE_STATUS_WITHDRAWN:
		return "withdrawn"
	default:
		return "pending"
	}
}

func responseStatusFromDB(status string) commonv1.ResponseStatus {
	switch status {
	case "accepted":
		return commonv1.ResponseStatus_RESPONSE_STATUS_ACCEPTED
	case "rejected":
		return commonv1.ResponseStatus_RESPONSE_STATUS_REJECTED
	case "withdrawn":
		return commonv1.ResponseStatus_RESPONSE_STATUS_WITHDRAWN
	default:
		return commonv1.ResponseStatus_RESPONSE_STATUS_PENDING
	}
}

func assignmentStatusToDB(status commonv1.AssignmentStatus) string {
	switch status {
	case commonv1.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED:
		return "completed"
	case commonv1.AssignmentStatus_ASSIGNMENT_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "active"
	}
}

func assignmentStatusFromDB(status string) commonv1.AssignmentStatus {
	switch status {
	case "completed":
		return commonv1.AssignmentStatus_ASSIGNMENT_STATUS_COMPLETED
	case "cancelled":
		return commonv1.AssignmentStatus_ASSIGNMENT_STATUS_CANCELLED
	default:
		return commonv1.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE
	}
}

func reviewTargetRoleToDB(role commonv1.ReviewTargetRole) string {
	if role == commonv1.ReviewTargetRole_REVIEW_TARGET_ROLE_OWNER {
		return "owner"
	}
	return "worker"
}

func reviewTargetRoleFromDB(role string) commonv1.ReviewTargetRole {
	if role == "owner" {
		return commonv1.ReviewTargetRole_REVIEW_TARGET_ROLE_OWNER
	}
	return commonv1.ReviewTargetRole_REVIEW_TARGET_ROLE_WORKER
}
