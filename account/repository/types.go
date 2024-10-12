package repository

import (
	"context"
	"geektime/webook/account/domain"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, c domain.Credit) error
}
