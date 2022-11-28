package test

import (
	"context"
	"github.com/995933447/distribmu/factory"
	"github.com/etcd-io/etcd/client"
	"testing"
	"time"
)

func TestEtcdMuLock(t *testing.T) {
	etcdCli, err := client.New(client.Config{
		Endpoints:               []string{"http://localhost:4001"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: 3 * time.Second,
	})
	if err != nil {
		t.Error(err)
		return
	}
	newMuConf := factory.NewMuConf(
		factory.MuTypeEtcd,
		"abc",
		time.Second * 20,
		factory.NewEtcdMuDriverConf(etcdCli, "1"),
		)
	mu := factory.MustNewMu(newMuConf)
	if existed, err := mu.Lock(context.Background()); err != nil {
		t.Error(err)
	} else if !existed {
		t.Log(existed)
	}
}
