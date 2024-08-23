package config

import (
	"context"
	"net"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewGrpcServer(ctx context.Context) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", viper.GetString("listen.url"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not create listener")
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainStreamInterceptor(
		func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			wss := middleware.WrapServerStream(ss)
			wss.WrappedContext = ctx

			return handler(srv, wss)
		},
	))

	if viper.IsSet("server.tls.enabled") && viper.GetBool("server.tls.enabled") {
		creds, err := credentials.NewServerTLSFromFile(
			viper.GetString("server.tls.certFile"),
			viper.GetString("server.tls.keyFile"),
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not load server credentials")
		}

		opts = append(opts, grpc.Creds(creds))
	}

	if viper.IsSet("server.maxConcurrentStreams") {
		opts = append(opts, grpc.MaxConcurrentStreams(
			viper.GetUint32("server.maxConcurrentStreams"),
		))
	}

	return grpc.NewServer(opts...), lis, nil
}
