package distribmu

import (
	"995933447/distribmu/impl/etcd"
	"github.com/etcd-io/etcd/client"
	"time"
)

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
		newEtcdMu(conf.key, conf.ttl, conf.muDriverConf.(*EtcdDriverConf))
	}

	panic(any("no support mutex type"))
}

func newEtcdMu(key string, ttl time.Duration, driverConf *EtcdDriverConf) *etcd.Mutex {
	return etcd.New(driverConf.etcdCli, key, driverConf.id, ttl)
}
