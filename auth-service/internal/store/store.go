package store

import (
	"context"
	"crypto/sha256"
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

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	pool *pgxpool.Pool
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type User struct {
	ID              int64
	Email           string
	EmailVerifiedAt *time.Time
	PasswordHash    string
	Role            commonv1.UserRole
	FullName        string
	Phone           *string
	AvatarURL       *string
	Status          string
	BlockedUntil    *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
		Columns("email", "password_hash", "role", "full_name", "phone").
		Values(user.Email, user.PasswordHash, roleToDB(user.Role), user.FullName, user.Phone).
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

func (s *Store) UpdateProfile(ctx context.Context, userID int64, fullName *string, phone *string, avatarURL *string) (User, error) {
	current, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if fullName != nil {
		current.FullName = *fullName
	}
	if phone != nil {
		current.Phone = phone
	}
	if avatarURL != nil {
		current.AvatarURL = avatarURL
	}

	sql, args, err := psql.Update("auth.users").
		Set("full_name", current.FullName).
		Set("phone", current.Phone).
		Set("avatar_url", current.AvatarURL).
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

func scanUser(row pgx.Row) (User, error) {
	var user User
	var role string
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.PasswordHash,
		&role,
		&user.FullName,
		&user.Phone,
		&user.AvatarURL,
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
	return "id, email, email_verified_at, password_hash, role, full_name, phone, avatar_url, status, blocked_until, created_at, updated_at"
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
		"u.full_name",
		"u.phone",
		"u.avatar_url",
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

func NormalizeEmail(email string) string {
	return normalizeEmail(email)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
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
