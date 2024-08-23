package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/smousa/kafka-grpc-stream/internal/config"
	"github.com/spf13/viper"
	"github.com/twmb/franz-go/pkg/kgo"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite", Label("e2e"))
}

var (
	serverExe   string
	etcdClient  *clientv3.Client
	kafkaClient *kgo.Client
)

var _ = SynchronizedBeforeSuite(
	func() {
		gofakeit.Seed(GinkgoRandomSeed())

		viper.SetDefault("kafka.seedBrokers", "localhost:9092")
		viper.SetDefault("etcd.endpoints", "localhost:2379")

		var err error

		etcdClient, err = config.NewEtcdClient()
		Ω(err).ShouldNot(HaveOccurred())

		kafkaClient, err = kgo.NewClient(
			kgo.AllowAutoTopicCreation(),
			kgo.SeedBrokers(viper.GetStringSlice("kafka.seedBrokers")...),
		)
		Ω(err).ShouldNot(HaveOccurred())

		serverExe, err = Build(
			"github.com/smousa/kafka-grpc-stream/server",
			"-buildvcs=false",
		)
		Ω(err).ShouldNot(HaveOccurred())
	},
	func() {
	},
)

var _ = SynchronizedAfterSuite(
	func() {
	},
	func() {
		if etcdClient != nil {
			etcdClient.Close()
		}
		if kafkaClient != nil {
			kafkaClient.Close()
		}
		CleanupBuildArtifacts()
	},
)
