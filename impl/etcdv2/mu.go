package etcdv2

import (
	"github.com/995933447/distribmu"
	"context"
	"errors"
	"github.com/etcd-io/etcd/client"
	"time"
)

type Mutex struct {
	id         string
	key		   string
	ttl		   time.Duration
	etcdCli    client.Client
	etcdKeyApi client.KeysAPI
	expireTime time.Time
}

var _ distribmu.Mutex = (*Mutex)(nil)

func (m *Mutex) GetExpireTime() time.Time {
	return m.expireTime
}

func (m *Mutex) Lock(ctx context.Context) (existed bool, err error) {
	existed = false
	err = nil
	setOptions := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       m.ttl,
	}
	_, err = m.etcdKeyApi.Set(ctx, m.key, m.id, setOptions)
	if err == nil {
		m.expireTime = time.Now().Add(m.ttl)
		return
	}
	e, ok := err.(client.Error)
	if !ok {
		// not etcd client error
		return
	}
	if e.Code != client.ErrorCodeNodeExist {
		return
	}
	// node has existed
	existed = true
	err = nil
	return
}

// LockWait is a distributed lock implementation.
// if bool is ok, it means lock is not acquired.
func (m *Mutex) LockWait(ctx context.Context, timeout time.Duration) (bool, error) {
	existed, err := m.Lock(ctx)
	if err != nil {
		return existed, err
	}
	if !existed {
		return false, nil
	}
	err = m.WaitKeyRelease(ctx, timeout)
	if err != nil {
		if err == distribmu.ErrWaitTimeout {
			return true, nil
		}
		return false, err
	}
	existed, err = m.Lock(ctx)
	if err != nil {
		return existed, err
	}
	if !existed {
		return false, nil
	}
	// 存在了
	return true, nil
}

func (m *Mutex) WaitKeyRelease(ctx context.Context, timeout time.Duration) error {
	resp, err := m.etcdKeyApi.Get(ctx, m.key, nil)
	if err != nil {
		if e, ok := err.(client.Error); ok {
			if e.Code == client.ErrorCodeKeyNotFound {
				return nil
			}
		}
		return err
	}
	watcherOptions := &client.WatcherOptions{
		AfterIndex: resp.Index,
		Recursive:  false,
	}
	watcher := m.etcdKeyApi.Watcher(m.key, watcherOptions)
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	resp, err = watcher.Next(ctx)
	if err != nil {
		if e, ok := err.(client.Error); ok {
			if e.Code == client.ErrorCodeKeyNotFound {
				return nil
			}
		}
		if err == context.DeadlineExceeded {
			return distribmu.ErrWaitTimeout
		}
		return err
	}
	if resp != nil && (resp.Action == "delete" || resp.Action == "expire") {
		return nil
	}
	return distribmu.ErrWaitTimeout
}

func (m *Mutex) Unlock(ctx context.Context) error {
	_, err := m.etcdKeyApi.Delete(ctx, m.key, nil)
	if err == nil {
		return nil
	}
	e, ok := err.(client.Error)
	if ok && e.Code == client.ErrorCodeKeyNotFound {
		return nil
	}
	return nil
}

// DoWithMustDone 分布式锁上后做一些事情，会一直重试至获取到了锁为止，所以要谨慎使用
func (m *Mutex) DoWithMustDone(ctx context.Context, timeout time.Duration, logic func() error) error {
	for {
		lost, err := m.LockWait(ctx, timeout)
		if err != nil {
			return err
		}

		// 没有获取到，重试
		if lost {
			continue
		}

		if err = logic(); err != nil {
			return err
		}

		if err := m.Unlock(ctx); err != nil {
			return err
		}

		return nil
	}
}

// DoWithMaxRetry 分布式锁上后做一些事情，如果锁成功，则执行func，否则尝试至最大重试次数
func (m *Mutex) DoWithMaxRetry(ctx context.Context, max int, timeout time.Duration, logic func() error) error {
	for i := 0; i < max; i++ {
		lost, err := m.LockWait(ctx, timeout)
		if err != nil {
			return err
		}
		// 没有获取到，重试
		if lost {
			continue
		}
		if err = logic(); err != nil {
			return err
		}
		if err := m.Unlock(ctx); err != nil {
		}
		return nil
	}

	return errors.New("over max retry")
}

func (m *Mutex) RefreshTTL(ctx context.Context) error {
	isLostLock := func() (bool, error) {
		resp, err := m.etcdKeyApi.Get(ctx, m.key, nil)
		if err != nil {
			return false, err
		}

		if resp.Node.Value != m.id {
			//fmt.printf("RefreshTTL my id != etcd id, my id is:%s etcd id is:%s\n", m.id, resp.Node.Value)
			return true, nil
		}

		return false, nil
	}

	if lost, err := isLostLock(); err != nil {
		return err
	} else if lost {
		return distribmu.ErrLockLost
	}

	setOptions := &client.SetOptions{
		TTL:     m.ttl,
		Refresh: true,
	}

	_, err := m.etcdKeyApi.Set(ctx, m.key, "", setOptions)
	if err == nil {
		m.expireTime = time.Now().Add(m.ttl)
		return nil
	}

	if lost, err := isLostLock(); err != nil {
		return err
	} else if lost {
		return distribmu.ErrLockLost
	}

	return err
}

func New(etcdCli client.Client, key, id string, ttl time.Duration) *Mutex {
	return &Mutex{
		id:     id,
		key: 	key,
		etcdCli:etcdCli,
		etcdKeyApi: client.NewKeysAPI(etcdCli),
		ttl: ttl,
	}
}
