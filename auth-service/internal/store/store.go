package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	pool *pgxpool.Pool
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type User struct {
	ID               int64
	Email            string
	EmailVerifiedAt  *time.Time
	PasswordHash     string
	Role             commonv1.UserRole
	LastName         string
	FirstName        string
	MiddleName       *string
	BirthDate        *time.Time
	Phone            *string
	AvatarURL        *string
	TelegramUsername *string
	MaxUsername      *string
	Status           string
	BlockedUntil     *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type OutboxEvent struct {
	ID        int64
	EventID   string
	Topic     string
	EventType string
	EventKey  string
	Payload   []byte
	Attempts  int
}

type OutboxEventInput struct {
	EventID   string
	Topic     string
	EventType string
	EventKey  string
	Payload   []byte
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateUserWithSession(ctx context.Context, user User, refreshToken string, refreshExpiresAt time.Time) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("auth.users").
		Columns("email", "password_hash", "role", "last_name", "first_name", "middle_name", "birth_date", "phone").
		Values(user.Email, user.PasswordHash, roleToDB(user.Role), user.LastName, user.FirstName, user.MiddleName, user.BirthDate, user.Phone).
		Suffix("RETURNING " + userColumns()).
		ToSql()
	if err != nil {
		return User{}, err
	}

	row := tx.QueryRow(ctx, sql, args...)

	created, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrAlreadyExists
		}
		return User{}, err
	}

	if err := insertRefreshSession(ctx, tx, created.ID, refreshToken, refreshExpiresAt); err != nil {
		return User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}

	return created, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, error) {
	sql, args, err := psql.Select(userColumns()).
		From("auth.users").
		Where(sq.Eq{"email": normalizeEmail(email)}).
		ToSql()
	if err != nil {
		return User{}, err
	}
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanUser(row)
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (User, error) {
	sql, args, err := psql.Select(userColumns()).
		From("auth.users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return User{}, err
	}
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanUser(row)
}

func (s *Store) CreateRefreshSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	sql, args, err := refreshSessionInsert(userID, refreshToken, expiresAt).ToSql()
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *Store) RefreshSessionUser(ctx context.Context, refreshToken string) (User, error) {
	sql, args, err := activeRefreshSessionUserSelect(refreshToken, false).ToSql()
	if err != nil {
		return User{}, err
	}
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanUser(row)
}

func (s *Store) RotateRefreshSession(ctx context.Context, oldRefreshToken string, newRefreshToken string, newExpiresAt time.Time) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	sql, args, err := activeRefreshSessionUserSelect(oldRefreshToken, true).ToSql()
	if err != nil {
		return User{}, err
	}
	user, err := scanUser(tx.QueryRow(ctx, sql, args...))
	if err != nil {
		return User{}, err
	}

	updateSQL, updateArgs, err := psql.Update("auth.refresh_sessions").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"token_hash": HashToken(oldRefreshToken)}).
		Where(sq.Eq{"revoked_at": nil}).
		ToSql()
	if err != nil {
		return User{}, err
	}
	tag, err := tx.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		return User{}, err
	}
	if tag.RowsAffected() == 0 {
		return User{}, ErrNotFound
	}

	insertSQL, insertArgs, err := refreshSessionInsert(user.ID, newRefreshToken, newExpiresAt).ToSql()
	if err != nil {
		return User{}, err
	}
	if _, err := tx.Exec(ctx, insertSQL, insertArgs...); err != nil {
		return User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Store) RevokeRefreshSession(ctx context.Context, refreshToken string) error {
	sql, args, err := psql.Update("auth.refresh_sessions").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"token_hash": HashToken(refreshToken)}).
		Where(sq.Eq{"revoked_at": nil}).
		ToSql()
	if err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkEmailVerified(ctx context.Context, userID int64) (User, error) {
	sql, args, err := psql.Update("auth.users").
		Set("email_verified_at", sq.Expr("COALESCE(email_verified_at, NOW())")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING " + userColumns()).
		ToSql()
	if err != nil {
		return User{}, err
	}
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanUser(row)
}

func (s *Store) UpdateProfile(ctx context.Context, userID int64, lastName *string, firstName *string, middleName *string, birthDate *time.Time, phone *string, avatarURL *string, telegramUsername *string, maxUsername *string) (User, error) {
	current, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if lastName != nil {
		current.LastName = *lastName
	}
	if firstName != nil {
		current.FirstName = *firstName
	}
	if middleName != nil {
		current.MiddleName = middleName
	}
	if birthDate != nil {
		current.BirthDate = birthDate
	}
	if phone != nil {
		current.Phone = phone
	}
	if avatarURL != nil {
		current.AvatarURL = avatarURL
	}
	if telegramUsername != nil {
		current.TelegramUsername = blankToNil(telegramUsername)
	}
	if maxUsername != nil {
		current.MaxUsername = blankToNil(maxUsername)
	}

	sql, args, err := psql.Update("auth.users").
		Set("last_name", current.LastName).
		Set("first_name", current.FirstName).
		Set("middle_name", current.MiddleName).
		Set("birth_date", current.BirthDate).
		Set("phone", current.Phone).
		Set("avatar_url", current.AvatarURL).
		Set("telegram_username", current.TelegramUsername).
		Set("max_username", current.MaxUsername).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING " + userColumns()).
		ToSql()
	if err != nil {
		return User{}, err
	}
	row := s.pool.QueryRow(ctx, sql, args...)
	return scanUser(row)
}

func (s *Store) BlockUser(ctx context.Context, userID int64, blockedUntil time.Time) (User, error) {
	sql, args, err := psql.Update("auth.users").
		Set("status", "blocked").
		Set("blocked_until", blockedUntil).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING " + userColumns()).
		ToSql()
	if err != nil {
		return User{}, err
	}
	return scanUser(s.pool.QueryRow(ctx, sql, args...))
}

func (s *Store) EnqueueOutboxEvent(ctx context.Context, input OutboxEventInput) error {
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		var err error
		eventID, err = NewEventID()
		if err != nil {
			return err
		}
	}
	sql, args, err := psql.Insert("auth.outbox_events").
		Columns("event_id", "topic", "event_type", "event_key", "payload").
		Values(eventID, input.Topic, input.EventType, input.EventKey, input.Payload).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *Store) ClaimPendingOutboxEvents(ctx context.Context, limit int, claimTimeout time.Duration) ([]OutboxEvent, error) {
	if limit < 1 {
		limit = 1
	}
	if claimTimeout <= 0 {
		claimTimeout = 5 * time.Minute
	}
	sql := `
UPDATE auth.outbox_events
SET status = 'processing',
    available_at = NOW() + ($2 * INTERVAL '1 second')
WHERE id IN (
	SELECT id
	FROM auth.outbox_events
	WHERE status IN ('pending', 'processing') AND available_at <= NOW()
	ORDER BY id
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
RETURNING id, event_id, topic, event_type, event_key, payload, attempts`

	rows, err := s.pool.Query(ctx, sql, limit, int64(claimTimeout.Seconds()))
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
	sql, args, err := psql.Update("auth.outbox_events").
		Set("status", "published").
		Set("published_at", sq.Expr("NOW()")).
		Set("last_error", nil).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *Store) MarkOutboxEventFailed(ctx context.Context, id int64, cause error, retryBackoff time.Duration) error {
	if retryBackoff <= 0 {
		retryBackoff = time.Second
	}
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	sql, args, err := psql.Update("auth.outbox_events").
		Set("status", "pending").
		Set("attempts", sq.Expr("attempts + 1")).
		Set("available_at", time.Now().Add(retryBackoff)).
		Set("last_error", message).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func scanUser(row pgx.Row) (User, error) {
	var user User
	var role string
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.PasswordHash,
		&role,
		&user.LastName,
		&user.FirstName,
		&user.MiddleName,
		&user.BirthDate,
		&user.Phone,
		&user.AvatarURL,
		&user.TelegramUsername,
		&user.MaxUsername,
		&user.Status,
		&user.BlockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	user.Role = roleFromDB(role)
	return user, nil
}

func insertRefreshSession(ctx context.Context, tx pgx.Tx, userID int64, refreshToken string, expiresAt time.Time) error {
	sql, args, err := refreshSessionInsert(userID, refreshToken, expiresAt).ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	return err
}

func userColumns() string {
	return "id, email, email_verified_at, password_hash, role, last_name, first_name, middle_name, birth_date, phone, avatar_url, telegram_username, max_username, status, blocked_until, created_at, updated_at"
}

func refreshSessionInsert(userID int64, refreshToken string, expiresAt time.Time) sq.InsertBuilder {
	return psql.Insert("auth.refresh_sessions").
		Columns("user_id", "token_hash", "expires_at").
		Values(userID, HashToken(refreshToken), expiresAt)
}

func activeRefreshSessionUserSelect(refreshToken string, forUpdate bool) sq.SelectBuilder {
	builder := psql.Select(
		"u.id",
		"u.email",
		"u.email_verified_at",
		"u.password_hash",
		"u.role",
		"u.last_name",
		"u.first_name",
		"u.middle_name",
		"u.birth_date",
		"u.phone",
		"u.avatar_url",
		"u.telegram_username",
		"u.max_username",
		"u.status",
		"u.blocked_until",
		"u.created_at",
		"u.updated_at",
	).
		From("auth.refresh_sessions rs").
		Join("auth.users u ON u.id = rs.user_id").
		Where(sq.Eq{"rs.token_hash": HashToken(refreshToken)}).
		Where(sq.Eq{"rs.revoked_at": nil}).
		Where(sq.Expr("rs.expires_at > NOW()"))
	if forUpdate {
		builder = builder.Suffix("FOR UPDATE")
	}
	return builder
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func NewEventID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func NormalizeEmail(email string) string {
	return normalizeEmail(email)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func blankToNil(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func roleToDB(role commonv1.UserRole) string {
	switch role {
	case commonv1.UserRole_USER_ROLE_OWNER:
		return "owner"
	case commonv1.UserRole_USER_ROLE_MODERATOR:
		return "moderator"
	default:
		return "worker"
	}
}

func roleFromDB(role string) commonv1.UserRole {
	switch role {
	case "owner":
		return commonv1.UserRole_USER_ROLE_OWNER
	case "moderator":
		return commonv1.UserRole_USER_ROLE_MODERATOR
	default:
		return commonv1.UserRole_USER_ROLE_WORKER
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
