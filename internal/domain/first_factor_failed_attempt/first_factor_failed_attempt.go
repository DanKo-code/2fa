package first_factor_failed_attempt

import "context"

type Repository interface {
	GetAndAddFirstFactorFailedAttempts(ctx context.Context, email string) (int64, error)
	DeleteFirstFactorFailedAttempts(ctx context.Context, email string) error
}
