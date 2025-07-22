package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	apperrors "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
)

const (
	FirstFactorPref = "ff"
)

type FirstFactor struct {
	db     *redis.Client
	config config.Config
}

func NewFirstFactor(db *redis.Client, config config.Config) *FirstFactor {
	return &FirstFactor{
		db: db, config: config,
	}
}

func (ff FirstFactor) SetFirstFactorCode(ctx context.Context, email, code string) error {
	operation := "SetFirstFactorCode"

	err := ff.db.Set(ctx, fmt.Sprintf("%s:%s", FirstFactorPref, email), code, ff.config.RedisFirstFactorCodeTTLMinutes).Err()
	if err != nil {
		return apperrors.Wrap(operation, err)
	}

	return nil
}

func (ff FirstFactor) GetFirstFactorCode(ctx context.Context, email string) (string, error) {
	operation := "GetFirstFactorCode"

	code, err := ff.db.Get(ctx, fmt.Sprintf("%s:%s", FirstFactorPref, email)).Result()
	if errors.Is(err, redis.Nil) {
		return "", apperrors.Wrap(operation, apperrors.ErrNotFound)
	} else if err != nil {
		return "", apperrors.Wrap(operation, err)
	}

	return code, nil
}

func (ff FirstFactor) DeleteFirstFactorCode(ctx context.Context, email string) error {
	operation := "DeleteFirstFactorCode"

	deleted, err := ff.db.Del(ctx, fmt.Sprintf("%s:%s", FirstFactorPref, email)).Result()
	if err != nil {
		return apperrors.Wrap(operation, err)
	}
	if deleted == 0 {
		return apperrors.Wrap(operation, apperrors.ErrNotFound)
	}

	return nil
}
