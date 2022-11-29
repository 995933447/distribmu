package util


import (
	"github.com/995933447/distribmu"
	"github.com/995933447/distribmu/impl/etcdv2"
	"github.com/995933447/distribmu/impl/redis"
	"github.com/995933447/redisgroup"
	"github.com/etcd-io/etcd/client"
	"time"
)

type MuType int

const (
	MuTypeNil MuType = iota
	MuTypeEtcd
	MuTypeRedis
)

type EtcdMuDriverConf struct {
	etcdCli client.Client
}

type RedisGroupMuDriverConf struct {
	redisGroup *redisgroup.Group
	precisionMillSec int
}

func NewEtcdMuDriverConf(etcdCli client.Client) *EtcdMuDriverConf {
	return &EtcdMuDriverConf{
		etcdCli: etcdCli,
	}
}


func NewRedisMuDriverConf(redisGroup *redisgroup.Group, precisionMillSec int) *RedisGroupMuDriverConf {
	return &RedisGroupMuDriverConf{
		redisGroup: redisGroup,
		precisionMillSec: precisionMillSec,
	}
}

type MuConf struct {
	muType MuType
	key string
	ttl time.Duration
	id string
	muDriverConf any
}

func NewMuConf(muType MuType, key string, ttl time.Duration, id string, muDriverConf any) *MuConf {
	switch muType {
	case MuTypeEtcd:
		_ = muDriverConf.(*EtcdMuDriverConf)
	case MuTypeRedis:
		_= muDriverConf.(*RedisGroupMuDriverConf)
	}
	return &MuConf{
		muType: muType,
		key: key,
		ttl: ttl,
		id: id,
		muDriverConf: muDriverConf,
	}
}

func MustNewMu(conf *MuConf) distribmu.Mutex {
	switch conf.muType {
	case MuTypeEtcd:
		return newEtcdMu(conf.key, conf.ttl, conf.id, conf.muDriverConf.(*EtcdMuDriverConf))
	case MuTypeRedis:
		return newRedisMu(conf.key, conf.ttl, conf.id, conf.muDriverConf.(*RedisGroupMuDriverConf))
	}

	panic(any("no support mutex type"))
}

func newEtcdMu(key string, ttl time.Duration, id string, driverConf *EtcdMuDriverConf) *etcdv2.Mutex {
	return etcdv2.New(driverConf.etcdCli, key, id, ttl)
}

func newRedisMu(key string, ttl time.Duration, id string, driverConf *RedisGroupMuDriverConf) *redis.Mutex {
	return redis.New(driverConf.redisGroup, key, id, ttl, driverConf.precisionMillSec)
}