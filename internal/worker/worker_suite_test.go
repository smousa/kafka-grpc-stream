package worker_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/smousa/kafka-grpc-stream/internal/config"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestWorker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Worker Suite", Label("integration"))
}

var (
	rootCtx    context.Context
	rootCancel context.CancelFunc
	etcdClient *clientv3.Client
)

var _ = BeforeSuite(func() {
	viper.SetDefault("etcd.endpoints", "localhost:2379")

	var err error
	rootCtx, rootCancel = context.WithCancel(context.TODO())
	etcdClient, err = config.NewEtcdClient()
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	rootCancel()
	if etcdClient != nil {
		etcdClient.Close()
	}
})
