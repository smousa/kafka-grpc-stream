package main

import (
	"bytes"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/smousa/kafka-grpc-stream/internal/config"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/worker"
	"github.com/spf13/viper"
)

//nolint:gochecknoglobals
var (
	Version string
	Date    string
	Commit  string
)

func main() {
	// Set up signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Set up logging
	config.SetupLogging()

	// Get worker info
	hostId, err := os.ReadFile("/etc/hostname")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not retrieve worker id")
	}

	workerInfo := worker.Worker{
		HostId:    string(bytes.TrimSpace(hostId)),
		HostAddr:  viper.GetString("listen.advertiseUrl"),
		Topic:     viper.GetString("worker.topic"),
		Partition: viper.GetString("worker.partition"),
	}

	logger := log.With().
		Str("version", Version+"-"+Commit+" "+Date).
		Str("host_id", workerInfo.HostId).
		Str("topic", workerInfo.Topic).
		Str("partition", workerInfo.Partition).
		Logger()
	ctx = logger.WithContext(ctx)

	// Initialize etcd client
	etcdClient, err := config.NewEtcdClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize etcd client connection")
	}
	defer etcdClient.Close()

	// Initialize kafka client
	kafkaClient, err := config.NewKafkaClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize kafka client connection")
	}
	defer kafkaClient.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Set up the registry
	registry := worker.NewEtcdRegistry(worker.WithEtcdClient(etcdClient))

	logger.Info().Msg("Registering worker")

	wg.Add(1)

	go func() {
		defer wg.Done()

		//nolint:godox
		// FIXME: If the worker leaves the cluster, then stop the consumer too
		// Will need to change the behavior so that the consumer will only exit
		// if there are no active publishers, but this is good enough for now.
		defer cancel()

		err := registry.Register(ctx, workerInfo, viper.GetInt64("worker.leaseExpirySeconds"))
		if err != nil {
			logger.Error().Err(err).Msg("Worker left the cluster")
		} else {
			logger.Info().Msg("Worker left the cluster")
		}
	}()

	// Set up the consumer
	subscriber := subscribe.NewKafkaSubscriber(kafkaClient, registry)

	logger.Info().Msg("Starting consumer")

	wg.Add(1)

	go func() {
		defer wg.Done()

		//nolint:godox
		// FIXME: If the consumer exits, then have the worker leave the cluster
		// too.
		// Will need to change the behavior so that the consumer will only exit
		// if there are no active publishers, but this is good enough for now.
		defer cancel()

		subscriber.Subscribe(ctx)
		logger.Info().Msg("Stopping consumer")
	}()
}
