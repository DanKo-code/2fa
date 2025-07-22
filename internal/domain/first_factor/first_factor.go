package first_factor

import "context"

type Repository interface {
	SetFirstFactorCode(ctx context.Context, email, code string) error
	GetFirstFactorCode(ctx context.Context, email string) (string, error)
	DeleteFirstFactorCode(ctx context.Context, email string) error
}
