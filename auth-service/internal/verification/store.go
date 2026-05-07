package verification

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tmcrypted/berusmenu/auth-service/internal/config"
)

var (
	ErrCooldown = errors.New("verification resend cooldown is active")
	ErrLockout  = errors.New("verification attempts lockout is active")
	ErrNoCode   = errors.New("verification code is not active")
	ErrBadCode  = errors.New("verification code is invalid")
)

type Store struct {
	redis *redis.Client
	cfg   config.Config
}

type CodeState struct {
	Code              string
	CodeExpiresAt     time.Time
	ResendAvailableAt time.Time
}

func New(redis *redis.Client, cfg config.Config) *Store {
	return &Store{redis: redis, cfg: cfg}
}

func (s *Store) CreateOrReuseCode(ctx context.Context, userID int64) (CodeState, error) {
	code, err := s.redis.Get(ctx, codeKey(userID)).Result()
	if errors.Is(err, redis.Nil) {
		code, err = randomCode()
		if err != nil {
			return CodeState{}, err
		}
		if err := s.redis.Set(ctx, codeKey(userID), code, s.cfg.CodeTTL).Err(); err != nil {
			return CodeState{}, err
		}
	} else if err != nil {
		return CodeState{}, err
	}

	codeTTL, err := s.redis.TTL(ctx, codeKey(userID)).Result()
	if err != nil {
		return CodeState{}, err
	}
	if codeTTL <= 0 {
		codeTTL = s.cfg.CodeTTL
	}
	if err := s.redis.Set(ctx, cooldownKey(userID), "1", s.cfg.ResendCooldown).Err(); err != nil {
		return CodeState{}, err
	}

	return CodeState{
		Code:              code,
		CodeExpiresAt:     time.Now().Add(codeTTL),
		ResendAvailableAt: time.Now().Add(s.cfg.ResendCooldown),
	}, nil
}

func (s *Store) ResendCode(ctx context.Context, userID int64) (CodeState, error) {
	if ttl, err := s.redis.TTL(ctx, cooldownKey(userID)).Result(); err == nil && ttl > 0 {
		return CodeState{
			CodeExpiresAt:     time.Now().Add(s.cfg.CodeTTL),
			ResendAvailableAt: time.Now().Add(ttl),
		}, ErrCooldown
	}
	if locked, err := s.isLocked(ctx, userID); err != nil {
		return CodeState{}, err
	} else if locked {
		return CodeState{}, ErrLockout
	}

	state, err := s.CreateOrReuseCode(ctx, userID)
	if err != nil {
		return CodeState{}, err
	}
	state.ResendAvailableAt = time.Now().Add(s.cfg.ResendCooldown)
	return state, nil
}

func (s *Store) VerifyCode(ctx context.Context, userID int64, submitted string) error {
	if locked, err := s.isLocked(ctx, userID); err != nil {
		return err
	} else if locked {
		return ErrLockout
	}

	code, err := s.redis.Get(ctx, codeKey(userID)).Result()
	if errors.Is(err, redis.Nil) {
		return ErrNoCode
	}
	if err != nil {
		return err
	}
	if code != submitted {
		return s.recordBadAttempt(ctx, userID)
	}

	return s.redis.Del(ctx, codeKey(userID), attemptsKey(userID), cooldownKey(userID)).Err()
}

func (s *Store) recordBadAttempt(ctx context.Context, userID int64) error {
	key := attemptsKey(userID)
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err := s.redis.Expire(ctx, key, s.cfg.CodeTTL).Err(); err != nil {
			return err
		}
	}
	if int(count) >= s.cfg.MaxVerifyAttempts {
		if err := s.redis.Set(ctx, key, strconv.FormatInt(count, 10), s.cfg.LockoutTTL).Err(); err != nil {
			return err
		}
		return ErrLockout
	}
	return ErrBadCode
}

func (s *Store) isLocked(ctx context.Context, userID int64) (bool, error) {
	countText, err := s.redis.Get(ctx, attemptsKey(userID)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	count, err := strconv.Atoi(countText)
	if err != nil {
		return false, nil
	}
	return count >= s.cfg.MaxVerifyAttempts, nil
}

func randomCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func codeKey(userID int64) string {
	return fmt.Sprintf("email_verify:%d", userID)
}

func cooldownKey(userID int64) string {
	return fmt.Sprintf("email_verify_resend:%d", userID)
}

func attemptsKey(userID int64) string {
	return fmt.Sprintf("email_verify_attempts:%d", userID)
}
