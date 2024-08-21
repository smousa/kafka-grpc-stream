package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

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
	serverCmd   *exec.Cmd
	etcdClient  *clientv3.Client
	kafkaClient *kgo.Client
	topic       string
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

		serverBuildPath, err := gexec.Build(
			"github.com/smousa/kafka-grpc-stream/server",
			"-buildvcs=false",
		)
		Ω(err).ShouldNot(HaveOccurred())
		serverCmd = exec.Command(serverBuildPath)
	},
	func() {
		topic = "e2e." + gofakeit.LetterN(12)
		serverCmd.Env = append(os.Environ(),
			"LISTEN_URL=localhost:9000",
			"LISTEN_ADVERTISE_URL=localhost:9000",
			"WORKER_TOPIC="+topic,
			"WORKER_PARTITION=0",
		)
	},
)

var _ = SynchronizedAfterSuite(
	func() {
		etcdClient.Delete(context.TODO(), "/topics/"+topic, clientv3.WithPrefix())
	},
	func() {
		if etcdClient != nil {
			etcdClient.Close()
		}
		if kafkaClient != nil {
			kafkaClient.Close()
		}
		gexec.CleanupBuildArtifacts()
	},
)
