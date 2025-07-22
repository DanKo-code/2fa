package second_factor

import "context"

type Repository interface {
	SetSecondFactorCode(ctx context.Context, email, code string) error
	GetSecondFactorCode(ctx context.Context, email string) (string, error)
}
