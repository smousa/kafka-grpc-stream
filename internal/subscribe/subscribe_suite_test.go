package subscribe_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/smousa/kafka-grpc-stream/internal/config"
	"github.com/spf13/viper"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestSubscribe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Subscribe Suite")
}

var (
	kafkaClient *kgo.Client
	topic       string
)

var _ = SynchronizedBeforeSuite(
	func() {
		gofakeit.Seed(GinkgoRandomSeed())
		if Label("integration").MatchesLabelFilter(GinkgoLabelFilter()) {
			viper.SetDefault("kafka.seedBrokers", "localhost:9092")

			var err error
			kafkaClient, err = kgo.NewClient(
				kgo.AllowAutoTopicCreation(),
				kgo.SeedBrokers(viper.GetStringSlice("kafka.seedBrokers")...),
				kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelDebug, nil)),
			)
			Ω(err).ShouldNot(HaveOccurred())
		}
	},
	func() {
		topic = "test.subscribe." + gofakeit.LetterN(12)
		if Label("integration").MatchesLabelFilter(GinkgoLabelFilter()) {
			kafkaClient.AddConsumePartitions(map[string]map[int32]kgo.Offset{
				topic: map[int32]kgo.Offset{
					0: kgo.NewOffset().AtEnd(),
				},
			})
		}
	},
)

var _ = SynchronizedAfterSuite(
	func() {
		if kafkaClient != nil {
			kafkaClient.RemoveConsumePartitions(map[string][]int32{
				topic: []int32{0},
			})
		}
	},
	func() {
		if kafkaClient != nil {
			kafkaClient.Close()
		}
	},
)
