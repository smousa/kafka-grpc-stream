package cmd

import (
	"context"
	"io"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smousa/kafka-grpc-stream/protobuf"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	serverAddr    string
	serverTlsCert string
	keys          []string

	conn  *grpc.ClientConn
	ksCli protobuf.KafkaStreamerClient
)

type TieredLevelWriter struct {
	io.Writer
	OutputWriter io.Writer
}

func (lw *TieredLevelWriter) WriteLevel(l zerolog.Level, data []byte) (n int, err error) {
	var writer io.Writer = lw
	if l == zerolog.NoLevel {
		writer = lw.OutputWriter
	}

	//nolint:wrapcheck
	return writer.Write(data)
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:           "cli",
	Short:         "A command line client for the kafka-grpc-stream server",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// configure the client
		var dialOps []grpc.DialOption

		if serverTlsCert != "" {
			creds, err := credentials.NewClientTLSFromFile(serverTlsCert, "")
			if err != nil {
				return errors.Wrap(err, "could not load credentials certificate")
			}

			dialOps = append(dialOps, grpc.WithTransportCredentials(creds))
		} else {
			dialOps = append(dialOps, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		var err error
		conn, err = grpc.NewClient(serverAddr, dialOps...)
		if err != nil {
			return errors.Wrap(err, "could not create connection to server")
		}

		ksCli = protobuf.NewKafkaStreamerClient(conn)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		conn.Close()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log := zerolog.Ctx(cmd.Context())

		// call subscribe
		stream, err := ksCli.Subscribe(cmd.Context())
		if err != nil {
			if status.Code(err) == codes.Canceled {
				return nil
			}

			return errors.Wrap(err, "could not subscribe to server")
		}

		err = stream.Send(&protobuf.SubscribeRequest{Keys: keys})
		if err != nil {
			if status.Code(err) == codes.Canceled {
				return nil
			}

			return errors.Wrap(err, "could not send request")
		}

		for {
			msg, err := stream.Recv()
			if err != nil {
				if status.Code(err) == codes.Canceled {
					return nil
				}

				return errors.Wrap(err, "could not receive response")
			}

			// print the output
			headers := zerolog.Arr()
			for _, h := range msg.GetHeaders() {
				headers.Dict(zerolog.Dict().
					Str("key", h.GetKey()).
					Str("value", h.GetValue()),
				)
			}
			log.Log().
				Str("key", msg.GetKey()).
				Bytes("value", msg.GetValue()).
				Array("headers", headers).
				Time("timestamp", time.UnixMilli(msg.GetTimestamp())).
				Str("topic", msg.GetTopic()).
				Int32("partition", msg.GetPartition()).
				Int64("offset", msg.GetOffset()).
				Msg("")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// set up logger
	zerolog.SetGlobalLevel(zerolog.NoLevel)
	rootLogger := zerolog.New(&TieredLevelWriter{
		Writer: &zerolog.ConsoleWriter{
			Out:        rootCmd.OutOrStderr(),
			TimeFormat: time.RFC3339,
		},
		OutputWriter: rootCmd.OutOrStdout(),
	})
	ctx = rootLogger.WithContext(ctx)

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		rootLogger.Fatal().
			Timestamp().
			Err(err).
			Msg("Command exited with error")
	}
}

//nolint:gochecknoinits
func init() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server-address", "localhost:9000", "Address of the server to connect.")
	rootCmd.PersistentFlags().StringVar(&serverTlsCert, "server-tls-cert", "", "Filename of the tls certificate to use")
	rootCmd.Flags().StringSliceVar(&keys, "keys", []string{"*"}, "Keys to filter (can use glob expressions).")
}
