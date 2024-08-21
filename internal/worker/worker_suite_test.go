package worker_test

import (
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

var etcdClient *clientv3.Client

var _ = BeforeSuite(func() {
	if Label("integration").MatchesLabelFilter(GinkgoLabelFilter()) {
		viper.SetDefault("etcd.endpoints", "localhost:2379")

		var err error
		etcdClient, err = config.NewEtcdClient()
		Ω(err).ShouldNot(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	if etcdClient != nil {
		etcdClient.Close()
	}
})
