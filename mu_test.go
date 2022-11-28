package distribmu

import (
	"context"
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
	ctx := context.Background()
	mu := MustNewMu(NewMuConf(MuTypeEtcd, "111", time.Second * 20, NewEtcdDriverConf(etcdCli, "abc")))
	if existed, err := mu.Lock(ctx); err != nil {
		t.Error(err)
	} else if !existed {
		t.Log(existed)
	}
}
