package distribmu

import (
	"995933447/distribmu/impl/etcd"
	"context"
	"github.com/etcd-io/etcd/client"
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

type MuType int

const (
	MuTypeNil MuType = iota
	MuTypeEtcd
)

type EtcdDriverConf struct {
	etcdCli client.Client
	id string
}

func NewEtcdDriverConf(etcdCli client.Client, id string) *EtcdDriverConf {
	return &EtcdDriverConf{
		etcdCli: etcdCli,
		id: id,
	}
}

type MuConf struct {
	muType MuType
	key string
	ttl time.Duration
	muDriverConf any
}

func NewMuConf(muType MuType, key string, ttl time.Duration, muDriverConf any) *MuConf {
	switch muType {
	case MuTypeEtcd:
		_ = muDriverConf.(*EtcdDriverConf)
	}
	return &MuConf{
		muType: muType,
		key: key,
		ttl: ttl,
		muDriverConf: muDriverConf,
	}
}

func MustNewMu(conf *MuConf) Mutex {
	switch conf.muType {
	case MuTypeEtcd:
		createEtcdMu(conf.key, conf.ttl, conf.muDriverConf.(*EtcdDriverConf))
	}

	panic(any("no support mutex type"))
}

func createEtcdMu(key string, ttl time.Duration, driverConf *EtcdDriverConf) *etcd.Mutex {
	return etcd.New(driverConf.etcdCli, key, driverConf.id, ttl)
}