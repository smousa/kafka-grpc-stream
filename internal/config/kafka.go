package config

import (
	"time"

	"github.com/spf13/viper"
	"github.com/twmb/franz-go/pkg/kgo"
)

func NewKafkaClient() (*kgo.Client, error) {
	var opts []kgo.Opt

	opts = append(opts, kgo.ClientID(viper.GetString("kafka.clientId")))
	opts = append(opts, kgo.SeedBrokers(viper.GetStringSlice("kafka.seedBrokers")...))
	opts = append(opts, kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
		viper.GetString("worker.topic"): {
			viper.GetInt32("worker.partition"): kgo.NewOffset().AfterMilli(time.Now().UnixMilli()),
		},
	}))

	if viper.IsSet("kafka.brokerMaxReadBytes") {
		opts = append(opts, kgo.BrokerMaxReadBytes(viper.GetInt32("kafka.brokerMaxReadBytes")))
	}

	if viper.IsSet("kafka.fetchMaxBytes") {
		opts = append(opts, kgo.FetchMaxBytes(viper.GetInt32("kafka.fetchMaxBytes")))
	}

	if viper.IsSet("kafka.fetchMaxPartitionBytes") {
		opts = append(opts, kgo.FetchMaxPartitionBytes(viper.GetInt32("kafka.fetchMaxPartitionBytes")))
	}

	if viper.IsSet("kafka.fetchMaxWait") {
		opts = append(opts, kgo.FetchMaxWait(viper.GetDuration("kafka.fetchMaxWait")))
	}

	if viper.IsSet("kafka.fetchMinBytes") {
		opts = append(opts, kgo.FetchMinBytes(viper.GetInt32("kafka.fetchMinBytes")))
	}

	if viper.IsSet("kafka.maxConcurrentFetches") {
		opts = append(opts, kgo.MaxConcurrentFetches(viper.GetInt("kafka.maxConcurrentFetches")))
	}

	//nolint:wrapcheck
	return kgo.NewClient(opts...)
}
