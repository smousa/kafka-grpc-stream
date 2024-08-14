package config

import (
	"github.com/spf13/viper"
)

const (
	DefaultWorkerLeaseExpirySeconds = 30
)

func loadLogEnv(v *viper.Viper) {
	v.MustBindEnv("log.level", "LOG_LEVEL")
	v.MustBindEnv("log.timeFieldFormat", "LOG_TIME_FIELD_FORMAT")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.timeFieldFormat", "unix")
}

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

func loadEnv(v *viper.Viper) {
	v.MustBindEnv("listen.url", "LISTEN_URL")
	v.MustBindEnv("listen.advertiseUrl", "LISTEN_ADVERTISE_URL")
	v.MustBindEnv("worker.topic", "WORKER_TOPIC")
	v.MustBindEnv("worker.partition", "WORKER_PARTITION")
	v.MustBindEnv("worker.leaseExpirySeconds", "WORKER_LEASE_EXPIRY_SECONDS")

	v.SetDefault("worker.leaseExpirySeconds", DefaultWorkerLeaseExpirySeconds)
}

//nolint:gochecknoinits
func init() {
	v := viper.GetViper()
	loadLogEnv(v)
	loadEtcdEnv(v)
	loadEnv(v)
}
