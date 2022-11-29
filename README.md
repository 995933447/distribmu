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
	getResp, err := client.NewKeysAPI(etcdCli).Get(context.TODO(), "abcd", nil)
	if err != nil {
		t.Error(err)
	}
	t.Logf("node val:%v", getResp.Node.Value)
	t.Logf("node val ttl:%v", getResp.Node.TTL)
}

```
