package service

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/smousa/kafka-grpc-stream/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BidiStream interface {
	grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]
}

type Service struct {
	protobuf.UnimplementedKafkaStreamerServer
	broadcast *subscribe.Broadcast
}

func New(broadcast *subscribe.Broadcast) *Service {
	return &Service{broadcast: broadcast}
}

//nolint:wrapcheck
func (s *Service) Subscribe(stream grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]) error {
	// Generate the session id
	id := uuid.Must(uuid.NewV7()).String()
	pub := subscribe.NewFilterPublisher(NewStreamPublisher(id, stream))

	s.broadcast.RegisterSession(id, pub)
	defer s.broadcast.UnregisterSession(id)

	// Wait for the client to close the connection
	for {
		req, err := stream.Recv()
		if err != nil {
			if status.Code(err) == codes.Canceled {
				break
			}

			return status.Error(codes.Unknown, errors.Wrap(err, "unexpected client error").Error())
		}

		// Try to compile key filter
		keyPattern, err := glob.Compile(fmt.Sprintf("{%s}", strings.Join(req.GetKeys(), ",")))
		if err != nil {
			return status.Error(codes.InvalidArgument, errors.Wrap(err, "could not compile criteria for keys").Error())
		}

		// Update the key filter
		pub.SetFilter(func(m *subscribe.Message) bool {
			return keyPattern.Match(m.Key)
		})
	}

	return nil
}
