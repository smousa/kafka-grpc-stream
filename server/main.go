package main

import (
	"bytes"
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/smousa/kafka-grpc-stream/internal/config"
	"github.com/smousa/kafka-grpc-stream/internal/worker"
	"github.com/spf13/viper"
)

var (
	Version string
	Date    string
	Commit  string
)

func main() {
	// Set up signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Set up logging
	config.SetupLogging()

	// Get worker info
	id, err := os.ReadFile("/etc/hostname")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not retrieve worker id")
	}
	w := worker.Worker{
		HostId:    string(bytes.TrimSpace(id)),
		HostAddr:  viper.GetString("listen.advertiseUrl"),
		Topic:     viper.GetString("worker.topic"),
		Partition: viper.GetString("worker.partition"),
	}

	logger := log.With().
		Str("version", Version+"-"+Commit+" "+Date).
		Str("host_id", w.HostId).
		Str("topic", w.Topic).
		Str("partition", w.Partition).
		Logger()
	ctx = logger.WithContext(ctx)

	// Initialize etcd client
	etcdClient, err := config.NewEtcdClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize etcd client connection")
	}
	defer etcdClient.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Set up the registry
	logger.Info().Msg("Registering worker")
	registry := worker.NewEtcdRegistry(worker.WithEtcdClient(etcdClient))
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := registry.Register(ctx, w, viper.GetInt64("worker.leaseExpirySeconds"))
		if err != nil {
			logger.Error().Err(err).Msg("Worker left the cluster")
		} else {
			logger.Info().Msg("Worker left the cluster")
		}
	}()
}
