package sms

import "context"

type Service interface {
	Send(ctx context.Context, args []string, numbers ...string) error
}
