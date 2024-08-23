package config

import (
	"github.com/spf13/viper"
)

const (
	DefaultWorkerLeaseExpirySeconds = 30
)

func loadEtcdEnv(v *viper.Viper) {
	v.MustBindEnv("etcd.endpoints", "ETCD_ENDPOINTS")
	v.MustBindEnv("etcd.autoSyncInterval", "ETCD_AUTO_SYNC_INTERVAL")
	v.MustBindEnv("etcd.dialTimeout", "ETCD_DIAL_TIMEOUT")
	v.MustBindEnv("etcd.dialKeepAliveTime", "ETCD_DIAL_KEEP_ALIVE_TIME")
	v.MustBindEnv("etcd.dialKeepAliveTimeout", "ETCD_DIAL_KEEP_ALIVE_TIMEOUT")
	v.MustBindEnv("etcd.maxCallSendMsgSize", "ETCD_MAX_CALL_SEND_MSG_SIZE")
	v.MustBindEnv("etcd.maxCallRecvMsgSize", "ETCD_MAX_CALL_RECV_MSG_SIZE")
	v.MustBindEnv("etcd.username", "ETCD_USERNAME")
	v.MustBindEnv("etcd.password", "ETCD_PASSWORD")
	v.MustBindEnv("etcd.rejectOldCluster", "ETCD_REJECT_OLD_CLUSTER")
	v.MustBindEnv("etcd.permitWithoutStream", "ETCD_PERMIT_WITHOUT_STREAM")
	v.MustBindEnv("etcd.maxUnaryRetries", "ETCD_MAX_UNARY_RETRIES")
	v.MustBindEnv("etcd.backoffWaitBetween", "ETCD_BACKOFF_WAIT_BETWEEN")
	v.MustBindEnv("etcd.backoffJitterFraction", "ETCD_BACKOFF_JITTER_FRACTION")
}

func loadKafkaEnv(v *viper.Viper) {
	v.MustBindEnv("kafka.brokerMaxReadBytes", "KAFKA_BROKER_MAX_READ_BYTES")
	v.MustBindEnv("kafka.clientId", "KAFKA_CLIENT_ID")
	v.MustBindEnv("kafka.seedBrokers", "KAFKA_SEED_BROKERS")
	v.MustBindEnv("kafka.fetchMaxBytes", "KAFKA_FETCH_MAX_BYTES")
	v.MustBindEnv("kafka.fetchMaxPartitionBytes", "KAFKA_FETCH_MAX_PARTITION_BYTES")
	v.MustBindEnv("kafka.fetchMaxWait", "KAFKA_FETCH_MAX_WAIT")
	v.MustBindEnv("kafka.fetchMinBytes", "KAFKA_FETCH_MIN_BYTES")
	v.MustBindEnv("kafka.maxConcurrentFetches", "KAFKA_MAX_CONCURRENT_FETCHES")

	v.SetDefault("kafka.clientId", "kafka-grpc-stream")
}

func loadEnv(v *viper.Viper) {
	v.MustBindEnv("log.level", "LOG_LEVEL")
	v.MustBindEnv("log.timeFieldFormat", "LOG_TIME_FIELD_FORMAT")

	v.MustBindEnv("listen.url", "LISTEN_URL")
	v.MustBindEnv("listen.advertiseUrl", "LISTEN_ADVERTISE_URL")

	v.MustBindEnv("server.tls.enabled", "SERVER_TLS_ENABLED")
	v.MustBindEnv("server.tls.certFile", "SERVER_TLS_CERT_FILE")
	v.MustBindEnv("server.tls.keyFile", "SERVER_TLS_KEY_FILE")
	v.MustBindEnv("server.maxConcurrentStreams", "SERVER_MAX_CONCURRENT_STREAMS")

	v.MustBindEnv("worker.topic", "WORKER_TOPIC")
	v.MustBindEnv("worker.partition", "WORKER_PARTITION")
	v.MustBindEnv("worker.leaseExpirySeconds", "WORKER_LEASE_EXPIRY_SECONDS")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.timeFieldFormat", "unix")
	v.SetDefault("worker.leaseExpirySeconds", DefaultWorkerLeaseExpirySeconds)
}

//nolint:gochecknoinits
func init() {
	v := viper.GetViper()
	loadEtcdEnv(v)
	loadKafkaEnv(v)
	loadEnv(v)
}
