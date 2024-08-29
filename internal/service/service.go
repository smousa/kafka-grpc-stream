package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

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

type SessionRegistry interface {
	RegisterSession(ctx context.Context, pub subscribe.Publisher)
}

type Service struct {
	protobuf.UnimplementedKafkaStreamerServer
	registry SessionRegistry
}

func New(registry SessionRegistry) *Service {
	return &Service{registry: registry}
}

func (s *Service) Subscribe(stream grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]) error {
	// Set up the publisher
	id := uuid.Must(uuid.NewV7()).String()
	pub := subscribe.NewFilterPublisher(NewStreamPublisher(id, stream))

	// Wait for the initial subscription
	filterFunc, err := s.handle(stream)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return err
	}

	pub.SetFilter(filterFunc)

	// Register the session
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	wg.Add(1)

	go func() {
		defer wg.Done()
		s.registry.RegisterSession(ctx, pub)
	}()

	// Wait for the client to close the connection
	for {
		filterFunc, err = s.handle(stream)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		}

		pub.SetFilter(filterFunc)
	}
}

//nolint:wrapcheck
func (s *Service) handle(stream grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]) (func(*subscribe.Message) bool, error) {
	// wait for the request
	req, err := stream.Recv()
	if err != nil {
		if status.Code(err) == codes.Canceled {
			return nil, context.Canceled
		}

		return nil, status.Error(codes.Unknown, errors.Wrap(err, "unexpected client error").Error())
	}

	// validate the request

	// validate the key globs
	keyPat, err := glob.Compile(fmt.Sprintf("{%s}", strings.Join(req.GetKeys(), ",")))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "invalid key glob").Error())
	}

	return func(msg *subscribe.Message) bool {
		return keyPat.Match(msg.Key)
	}, nil
}
