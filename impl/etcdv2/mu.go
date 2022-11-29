package etcdv2

import (
	"context"
	"github.com/995933447/distribmu"
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

func (m *Mutex) Lock(ctx context.Context) (bool, error) {
	setOptions := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       m.ttl,
	}
	_, err := m.etcdKeyApi.Set(ctx, m.key, m.id, setOptions)
	if err != nil {
		e, ok := err.(client.Error)
		if !ok {
			// not etcd client error
			return false, err
		}
		if e.Code != client.ErrorCodeNodeExist {
			return false, err
		}
		return false, nil
	}

	m.expireTime = time.Now().Add(m.ttl)

	return true, nil
}

// LockWait is a distributed lock implementation.
// if bool is ok, it means locked
func (m *Mutex) LockWait(ctx context.Context, timeout time.Duration) (bool, error) {
	locked, err := m.Lock(ctx)
	if err != nil {
		return false, err
	}

	if locked {
		return true, nil
	}

	err = m.WaitKeyRelease(ctx, timeout)
	if err != nil {
		if err == distribmu.ErrWaitTimeout {
			return true, nil
		}
		return false, err
	}

	locked, err = m.Lock(ctx)
	if err != nil {
		return false, err
	}

	if locked {
		return true, nil
	}

	return false, nil
}

func (m *Mutex) WaitKeyRelease(ctx context.Context, timeout time.Duration) (error) {
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

func (m *Mutex) Unlock(ctx context.Context, force bool) error {
	if force || m.expireTime.After(time.Now()) {
		_, err := m.etcdKeyApi.Delete(ctx, m.key, nil)
		if err != nil {
			e, ok := err.(client.Error)
			if !ok || e.Code != client.ErrorCodeKeyNotFound {
				return err
			}
		}
	}
	return nil
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
	if err != nil {
		return err
	}

	m.expireTime = time.Now().Add(m.ttl)

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
