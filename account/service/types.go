package service

import (
	"context"
	"geektime/webook/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, cr domain.Credit) error
}
