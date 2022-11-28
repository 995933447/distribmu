package distribmu

import (
	"context"
	"time"
)

type Mutex interface {
	GetExpireTime() time.Time
	Lock(ctx context.Context) (existed bool, err error)
	LockWait(ctx context.Context, timeout time.Duration) (bool, error)
	WaitKeyRelease(ctx context.Context, timeout time.Duration) error
	Unlock(ctx context.Context) error
	DoWithMustDone(ctx context.Context, timeout time.Duration, logic func() error) error
	DoWithMaxRetry(ctx context.Context, max int, timeout time.Duration, logic func() error) error
	RefreshTTL(ctx context.Context) error
}
