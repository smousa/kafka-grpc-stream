package service

import (
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

func (s *Service) Subscribe(stream grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]) error {
	// Generate the session id
	id := uuid.Must(uuid.NewV7()).String()
	s.broadcast.RegisterSession(id, NewStreamPublisher(id, stream))
	defer s.broadcast.UnregisterSession(id)

	// Wait for the client to close the connection
	for {
		_, err := stream.Recv()
		if err != nil {
			if status.Code(err) == codes.Canceled {
				break
			}

			return errors.Wrap(err, "unexpected client error")
		}
	}

	return nil
}
