package test

import (
	"context"
	"github.com/995933447/distribmu"
	"github.com/995933447/distribmu/util"
	logger "github.com/995933447/log-go"
	"github.com/995933447/log-go/impls/loggerwriters"
	"github.com/995933447/redisgroup"
	"github.com/etcd-io/etcd/client"
	"testing"
	"time"
)

func TestRedisGroupMuLock(t *testing.T) {
	redisGroup := redisgroup.NewGroup([]*redisgroup.Node{
		redisgroup.NewNode("127.0.0.1", 6379, ""),
	}, logger.NewLogger(MustNewFileLoggerWriter()))

	muCfg := util.NewMuConf(
		util.MuTypeRedis,
		"abc",
		time.Second * 10,
		"123",
		util.NewRedisMuDriverConf(redisGroup, 500),
		)
	mu := util.MustNewMu(muCfg)
	success, err := mu.Lock(context.Background())
	if err != nil {
		t.Error(err)
	}
	t.Logf("bool:%v", success)
	time.Sleep(time.Second * 2)
	err = mu.RefreshTTL(context.Background())
	if err != nil {
		t.Error(err)
	}
	success, err = mu.LockWait(context.Background(), time.Second * 10)
	if err != nil {
		t.Error(err)
	}
	t.Logf("bool:%v", success)
	//err = mu.WaitKeyRelease(context.Background(), time.Second * 10)
	//if err != nil {
	//	t.Error(err)
	//}
	err = distribmu.DoWithMaxRetry(context.Background(), mu, 3, time.Second, func() error {
		t.Log("hello")
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	err = distribmu.DoWithMustDone(context.Background(), mu, time.Second, func() error {
		t.Log("hello")
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func CheckTimeToOpenNewFileHandlerForFileLogger() loggerwriters.CheckTimeToOpenNewFileFunc {
	return func(lastOpenFileTime *time.Time, isNeverOpenFile bool) (string, bool) {
		if isNeverOpenFile {
			return time.Now().Format("2006010215.log"), true
		}

		if lastOpenFileTime.Hour() != time.Now().Hour() {
			return time.Now().Format("2006010215.log"), true
		}

		lastOpenYear, lastOpenMonth, lastOpenDay := lastOpenFileTime.Date()
		nowYear, nowMonth, nowDay := time.Now().Date()
		if lastOpenDay != nowDay || lastOpenMonth != nowMonth || lastOpenYear != nowYear {
			return time.Now().Format("2006010215.log"), true
		}

		return "", false
	}
}

func MustNewFileLoggerWriter() logger.LoggerWriter {
	loggerWriter := loggerwriters.NewFileLoggerWriter(
		"./",
		1023,
		10,
		CheckTimeToOpenNewFileHandlerForFileLogger(),
		100000,
	)

	go func() {
		if err := loggerWriter.Loop(); err != nil {
			panic(any(err))
		}
	}()

	return loggerWriter
}

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
	newMuConf := util.NewMuConf(
		util.MuTypeEtcd,
		"/abcd",
		time.Second * 10,
		"123",
		util.NewEtcdMuDriverConf(etcdCli),
		)
	mu := util.MustNewMu(newMuConf)
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
}
