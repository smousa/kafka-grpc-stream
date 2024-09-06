package config

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/smousa/kafka-grpc-stream/internal/metrics"
	"github.com/spf13/viper"
)

func NewMetricsServer() *http.Server {
	// register metrics
	metrics.Register()

	// set up server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:        viper.GetString("metrics.listen.url"),
		Handler:     mux,
		ReadTimeout: viper.GetDuration("metrics.readTimeout"),
	}
}
