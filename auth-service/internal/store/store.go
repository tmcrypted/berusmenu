package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	commonv1 "gitlab.com/berusmenu/proto/gen/go/common/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	pool *pgxpool.Pool
}

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

	row := tx.QueryRow(ctx, `
		INSERT INTO auth.users (email, password_hash, role, full_name, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, email_verified_at, password_hash, role, full_name, phone, avatar_url, status, created_at, updated_at
	`, user.Email, user.PasswordHash, roleToDB(user.Role), user.FullName, user.Phone)

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
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, email_verified_at, password_hash, role, full_name, phone, avatar_url, status, created_at, updated_at
		FROM auth.users
		WHERE email = $1
	`, normalizeEmail(email))
	return scanUser(row)
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, email_verified_at, password_hash, role, full_name, phone, avatar_url, status, created_at, updated_at
		FROM auth.users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func (s *Store) CreateRefreshSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO auth.refresh_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, HashToken(refreshToken), expiresAt)
	return err
}

func (s *Store) RefreshSessionUser(ctx context.Context, refreshToken string) (User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.email_verified_at, u.password_hash, u.role, u.full_name, u.phone, u.avatar_url, u.status, u.created_at, u.updated_at
		FROM auth.refresh_sessions rs
		JOIN auth.users u ON u.id = rs.user_id
		WHERE rs.token_hash = $1
		  AND rs.revoked_at IS NULL
		  AND rs.expires_at > NOW()
	`, HashToken(refreshToken))
	return scanUser(row)
}

func (s *Store) RevokeRefreshSession(ctx context.Context, refreshToken string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE auth.refresh_sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, HashToken(refreshToken))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkEmailVerified(ctx context.Context, userID int64) (User, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE auth.users
		SET email_verified_at = COALESCE(email_verified_at, NOW()), updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, email_verified_at, password_hash, role, full_name, phone, avatar_url, status, created_at, updated_at
	`, userID)
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

	row := s.pool.QueryRow(ctx, `
		UPDATE auth.users
		SET full_name = $2, phone = $3, avatar_url = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, email_verified_at, password_hash, role, full_name, phone, avatar_url, status, created_at, updated_at
	`, userID, current.FullName, current.Phone, current.AvatarURL)
	return scanUser(row)
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
	_, err := tx.Exec(ctx, `
		INSERT INTO auth.refresh_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, HashToken(refreshToken), expiresAt)
	return err
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
