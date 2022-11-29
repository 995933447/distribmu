package distribmu

import (
	"context"
	"errors"
	"time"
)

type Mutex interface {
	// 获取过期时间
	GetExpireTime() time.Time
	// 进行分布式锁,失败返回false
	Lock(ctx context.Context) (bool, error)
	// 进行分布式锁,并尝试在timeout时间期间内直到获得锁,失败返回false
	LockWait(ctx context.Context, timeout time.Duration) (bool, error)
	// 在timeout时间期间内等待锁释放
	WaitKeyRelease(ctx context.Context, timeout time.Duration) error
	// 释放锁
	Unlock(ctx context.Context, force bool) error
	// 刷新过期时间，刷新时间为当前时间+ttl
	RefreshTTL(ctx context.Context) error
}


// DoWithMustDone 分布式锁上后做一些事情，会一直重试至获取到了锁为止，所以要谨慎使用
func DoWithMustDone(ctx context.Context, mu Mutex, timeout time.Duration, logic func() error) error {
	for {
		locked, err := mu.LockWait(ctx, timeout)
		if err != nil {
			return err
		}

		// 没有获取到，重试
		if !locked {
			continue
		}

		if err = logic(); err != nil {
			return err
		}

		if err := mu.Unlock(ctx, false); err != nil {
			return err
		}
		return nil
	}
}

// DoWithMaxRetry 分布式锁上后做一些事情，如果锁成功，则执行func，否则尝试至最大重试次数
func DoWithMaxRetry(ctx context.Context, mu Mutex, max int, timeout time.Duration, logic func() error) error {
	for i := 0; i < max; i++ {
		locked, err := mu.LockWait(ctx, timeout)
		if err != nil {
			return err
		}
		// 没有获取到，重试
		if !locked {
			continue
		}
		if err = logic(); err != nil {
			return err
		}
		if err := mu.Unlock(ctx, false); err != nil {
		}
		return nil
	}
	return errors.New("over max retry")
}
