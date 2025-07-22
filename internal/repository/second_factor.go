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
	SecondFactorPref = "sf"
)

type SecondFactor struct {
	db     *redis.Client
	config config.Config
}

func NewSecondFactor(db *redis.Client, config config.Config) *SecondFactor {
	return &SecondFactor{
		db: db, config: config,
	}
}

func (sf SecondFactor) SetSecondFactorCode(ctx context.Context, email, firstFactor string) error {
	operation := "SetSecondFactorCode"

	err := sf.db.Set(ctx, fmt.Sprintf("%s:%s", SecondFactorPref, email), firstFactor, sf.config.RedisSecondFactorCodeTTLDays).Err()
	if err != nil {
		return apperrors.Wrap(operation, err)
	}

	return nil
}

func (sf SecondFactor) GetSecondFactorCode(ctx context.Context, email string) (string, error) {
	operation := "GetSecondFactorCode"

	code, err := sf.db.Get(ctx, fmt.Sprintf("%s:%s", SecondFactorPref, email)).Result()
	if errors.Is(err, redis.Nil) {
		return "", apperrors.Wrap(operation, apperrors.ErrNotFound)
	} else if err != nil {
		return "", apperrors.Wrap(operation, err)
	}

	return code, nil
}
