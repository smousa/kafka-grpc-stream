package config

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcdClient() (*clientv3.Client, error) {
	cfg := clientv3.Config{
		Endpoints:             viper.GetStringSlice("etcd.endpoints"),
		AutoSyncInterval:      viper.GetDuration("etcd.autoSyncInterval"),
		DialTimeout:           viper.GetDuration("etcd.dialTimeout"),
		DialKeepAliveTime:     viper.GetDuration("etcd.dialKeepAliveTime"),
		DialKeepAliveTimeout:  viper.GetDuration("etcd.dialKeepAliveTimeout"),
		MaxCallSendMsgSize:    viper.GetInt("etcd.maxCallSendMsgSize"),
		MaxCallRecvMsgSize:    viper.GetInt("etcd.maxCallRecvMsgSize"),
		Username:              viper.GetString("etcd.username"),
		Password:              viper.GetString("etcd.password"),
		RejectOldCluster:      viper.GetBool("etcd.rejectOldCluster"),
		PermitWithoutStream:   viper.GetBool("etcd.permitWithoutStream"),
		MaxUnaryRetries:       viper.GetUint("etcd.maxUnaryRetries"),
		BackoffWaitBetween:    viper.GetDuration("etcd.backoffWaitBetween"),
		BackoffJitterFraction: viper.GetFloat64("etcd.backoffJitterFraction"),
	}

	//nolint:wrapcheck
	return clientv3.New(cfg)
}
