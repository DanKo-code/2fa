package repository

import (
	"context"

	"github.com/redis/go-redis/v9"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
)

func Connect(ctx context.Context, config config.Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     ":" + config.RedisPort,
		Password: config.RedisPass,
	})

	err := db.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return db, nil
}
