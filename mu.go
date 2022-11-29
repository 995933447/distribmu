package distribmu

import (
	"context"
	"time"
)

type Mutex interface {
	// 获取过期时间
	GetExpireTime() time.Time
	// 进行分布式锁,失败返回false
	Lock(ctx context.Context) (bool, err error)
	// 进行分布式锁,并尝试在timeout时间期间内直到获得锁,失败返回false
	LockWait(ctx context.Context, timeout time.Duration) (bool, error)
	// 在timeout时间期间内等待锁释放
	WaitKeyRelease(ctx context.Context, timeout time.Duration) error
	// 释放锁
	Unlock(ctx context.Context) error
	// 获取分布式锁后执行，并尝试在timeout时间期间内直到获得锁
	DoWithMustDone(ctx context.Context, timeout time.Duration, logic func() error) error
	// 获取分布式锁后执行，并尝试在timeout时间期间内最多执行max次直到获得锁
	DoWithMaxRetry(ctx context.Context, max int, timeout time.Duration, logic func() error) error
	// 刷新过期时间，刷新时间为当前时间+ttl
	RefreshTTL(ctx context.Context) error
}