package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	apperrors "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
)

const (
	FirstFactorFailedAttemptsPref = "fffa"
)

type FirstFactorFailedAttempt struct {
	db     *redis.Client
	config config.Config
}

func NewFirstFactorFailedAttempt(db *redis.Client, config config.Config) *FirstFactorFailedAttempt {
	return &FirstFactorFailedAttempt{
		db: db, config: config,
	}
}

func (fffa FirstFactorFailedAttempt) GetAndAddFirstFactorFailedAttempts(ctx context.Context, email string) (int64, error) {
	fac, err := fffa.db.Incr(ctx, fmt.Sprintf("%s:%s", FirstFactorFailedAttemptsPref, email)).Result()
	if err != nil {
		return 0, err
	}

	return fac, nil
}

func (fffa FirstFactorFailedAttempt) DeleteFirstFactorFailedAttempts(ctx context.Context, email string) error {
	deleted, err := fffa.db.Del(ctx, fmt.Sprintf("%s:%s", FirstFactorFailedAttemptsPref, email)).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
