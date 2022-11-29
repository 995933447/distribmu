# distribmu
简单易用的分布式锁，支持api如下
````
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
````
助手函数
```
package distribmu

import (
	"context"
	"errors"
	"time"
)

// DoWithMustDone 分布式锁上后做一些事情，会一直重试至获取到了锁为止，所以要谨慎使用
func DoWithMustDone(ctx context.Context, mu Mutex, timeout time.Duration, logic func() error) error {...}

// DoWithMaxRetry 分布式锁上后做一些事情，如果锁成功，则执行func，否则尝试至最大重试次数
func DoWithMaxRetry(ctx context.Context, mu Mutex, max int, timeout time.Duration, logic func() error) error {...}
```
使用示例：
```
package test

import (
	"context"
	"github.com/995933447/distribmu/factory"
	"github.com/etcd-io/etcd/client"
	"testing"
	"time"
)

func TestEtcdMuLock(t *testing.T) {
	t.Log("start")
	etcdCli, err := client.New(client.Config{
		Endpoints:               []string{"http://127.0.0.1:2379"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: 3 * time.Second,
	})
	if err != nil {
		t.Error(err)
		return
	}
	newMuConf := factory.NewMuConf(
		factory.MuTypeEtcd,
		"/abcd",
		time.Second * 10,
		factory.NewEtcdMuDriverConf(etcdCli, "123"),
		)
	mu := factory.MustNewMu(newMuConf)
	success, err := mu.Lock(context.Background())
	if err != nil {
		t.Error(err)
	}
	t.Logf("bool:%v", success)
	err = mu.RefreshTTL(context.Background())
	if err != nil {
		t.Error(err)
	}
	success, err = mu.LockWait(context.Background(), time.Second * 10)
	if err != nil {
		t.Error(err)
	}
	t.Logf("bool:%v", success)
}

```
