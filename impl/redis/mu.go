package redis

import (
	"context"
	"github.com/995933447/distribmu"
	"github.com/995933447/redisgroup"
	"time"
)

type Mutex struct {
	id         string
	key		   string
	ttl		   time.Duration
	redisGroup *redisgroup.Group
	expireTime time.Time
	precisionMillSec int
}

// 获取过期时间
func (m *Mutex) GetExpireTime() time.Time {
	return m.expireTime	
}

// 进行分布式锁,失败返回false
func (m *Mutex) Lock(ctx context.Context) (bool, error) {
	locked, err := m.redisGroup.SetNX(ctx, m.key, []byte(m.id), m.ttl)
	if err != nil {
		return false, err
	}

	if !locked {
		return false, nil
	}

	m.expireTime = time.Now().Add(m.ttl)
	return true, nil
}

// 进行分布式锁,并尝试在timeout时间期间内直到获得锁,失败返回false
func (m *Mutex) LockWait(ctx context.Context, timeout time.Duration) (bool, error) {
	tryAt := time.Now()
	for {
		locked, err := m.Lock(ctx)
		if err != nil {
			return false, err
		}

		if locked {
			return true, nil
		}

		if time.Now().Sub(tryAt) >= timeout {
			break
		}

		time.Sleep(time.Millisecond * time.Duration(m.precisionMillSec))

		//if time.Now().Sub(tryAt) >= timeout {
		//	break
		//}
	}
	return false, nil
}

// 在timeout时间期间内等待锁释放
func (m *Mutex) WaitKeyRelease(ctx context.Context, timeout time.Duration) error {
	tryAt := time.Now()
	for {
		// 不使用ttl，key可能会被人为删除
		existed, err := m.redisGroup.Exists(ctx, m.key)
		if err != nil {
			return err
		}

		if !existed {
			return nil
		}

		if time.Now().Sub(tryAt) >= timeout {
			break
		}

		time.Sleep(time.Millisecond * time.Duration(m.precisionMillSec))

		//if time.Now().Sub(tryAt) >= timeout {
		//	break
		//}
	}
	return distribmu.ErrWaitTimeout
}

// 释放锁
func (m *Mutex) Unlock(ctx context.Context, force bool) error {
	if force || m.expireTime.After(time.Now()) {
		err := m.redisGroup.Del(ctx, m.key)
		if err != nil {
			return err
		}
	}
	return nil
}

// 刷新过期时间，刷新时间为当前时间+ttl
func (m *Mutex) RefreshTTL(ctx context.Context) error {
	idBytes, err := m.redisGroup.Get(ctx, m.key)
	if err != nil {
		return err
	}

	if string(idBytes) != m.id {
		return distribmu.ErrLockLost
	}

	err = m.redisGroup.Expire(ctx, m.key, m.ttl)
	if err != nil {
		return err
	}

	idBytes, err = m.redisGroup.Get(ctx, m.key)
	if err != nil {
		return err
	}

	if string(idBytes) != m.id {
		return distribmu.ErrLockLost
	}

	m.expireTime = time.Now().Add(m.ttl)

	return nil
}

func New(redisGroup *redisgroup.Group, key, id string, ttl time.Duration, precisionMillSec int) *Mutex {
	return &Mutex{
		id:     id,
		key: 	key,
		redisGroup:redisGroup,
		ttl: ttl,
		precisionMillSec: precisionMillSec,
	}
}