package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/smousa/kafka-grpc-stream/internal/config"
	"github.com/smousa/kafka-grpc-stream/internal/service"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/worker"
	"github.com/smousa/kafka-grpc-stream/protobuf"
	"github.com/spf13/viper"
)

var (
	Version string
	Date    string
	Commit  string
)

func main() {
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

	// Set up logging
	config.SetupLogging()

	logger := log.With().
		Str("version", Version+"-"+Commit+" "+Date).
		Str("host_id", workerInfo.HostId).
		Str("topic", workerInfo.Topic).
		Str("partition", workerInfo.Partition).
		Logger()

	// Set up root context
	rootCtx := logger.WithContext(context.Background())

	rootCtx, rootCancel := context.WithCancel(rootCtx)
	defer rootCancel()

	// Set up signal handling
	ctx, cancel := signal.NotifyContext(rootCtx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize the grpc server
	srv, lis, err := config.NewGrpcServer(rootCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize grpc server")
	}

	// Initialize kafka client
	kafkaClient, err := config.NewKafkaClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize kafka client connection")
	}
	defer kafkaClient.Close()

	// Initialize etcd client
	etcdClient, err := config.NewEtcdClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize etcd client connection")
	}
	defer etcdClient.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Set up monitoring
	if viper.GetBool("metrics.enabled") {
		metricsServer := config.NewMetricsServer()
		defer metricsServer.Close()

		logger.Info().Msg("Starting metrics server")

		wg.Add(1)

		go func() {
			defer wg.Done()

			var err error
			if viper.GetBool("metrics.tls.enabled") {
				err = metricsServer.ListenAndServeTLS(
					viper.GetString("metrics.tls.certFile"),
					viper.GetString("metrics.tls.keyFile"),
				)
			} else {
				err = metricsServer.ListenAndServe()
			}

			if errors.Is(err, http.ErrServerClosed) {
				err = nil
			}

			logger.Err(err).Msg("Metrics server stopped")
		}()
	}

	// Set up the registry
	registry := worker.NewEtcdRegistry(
		worker.WithEtcdClient(etcdClient),
		worker.WithWorkerTTL(viper.GetInt64("worker.leaseExpirySeconds")),
		worker.WithKeyTTL(viper.GetInt64("key.leaseExpirySeconds")),
	)

	logger.Info().Msg("Registering worker")

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer cancel()

		err := registry.Register(ctx, workerInfo)
		logger.Err(err).Msg("Worker left the cluster")
	}()

	// Set up the publisher
	broadcast := subscribe.NewBufferedBroadcast(viper.GetInt("server.broadcast.bufferSize"))

	// Set up the server
	s := service.New(broadcast)
	protobuf.RegisterKafkaStreamerServer(srv, s)

	logger.Info().Str("address", lis.Addr().String()).Msg("Starting server")

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer rootCancel()

		err := srv.Serve(lis)
		logger.Err(err).Msg("Server stopped")
	}()

	// Set up the consumer
	subscriber := subscribe.NewKafkaSubscriber(kafkaClient, broadcast)

	logger.Info().Msg("Starting consumer")

	wg.Add(1)

	go func() {
		defer wg.Done()
		subscriber.Subscribe(rootCtx)
		logger.Info().Msg("Stopping consumer")
	}()

	// Watch the signal handler
	<-ctx.Done()
	srv.GracefulStop()
}
