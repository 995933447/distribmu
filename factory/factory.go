package factory


import (
	"github.com/995933447/distribmu"
	"github.com/995933447/distribmu/impl/etcdv2"
	"github.com/etcd-io/etcd/client"
	"time"
)

type MuType int

const (
	MuTypeNil MuType = iota
	MuTypeEtcd
)

type EtcdMuDriverConf struct {
	etcdCli client.Client
	id string
}

func NewEtcdMuDriverConf(etcdCli client.Client, id string) *EtcdMuDriverConf {
	return &EtcdMuDriverConf{
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
		_ = muDriverConf.(*EtcdMuDriverConf)
	}
	return &MuConf{
		muType: muType,
		key: key,
		ttl: ttl,
		muDriverConf: muDriverConf,
	}
}

func MustNewMu(conf *MuConf) distribmu.Mutex {
	switch conf.muType {
	case MuTypeEtcd:
		return newEtcdMu(conf.key, conf.ttl, conf.muDriverConf.(*EtcdMuDriverConf))
	}

	panic(any("no support mutex type"))
}

func newEtcdMu(key string, ttl time.Duration, driverConf *EtcdMuDriverConf) *etcdv2.Mutex {
	return etcdv2.New(driverConf.etcdCli, key, driverConf.id, ttl)
}

