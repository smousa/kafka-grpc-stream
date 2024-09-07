package service

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/smousa/kafka-grpc-stream/internal/metrics"
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
	// Reset the session count
	metrics.ActiveSessions.Set(0)

	return &Service{registry: registry}
}

func (s *Service) Subscribe(stream grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]) error {
	// Track the number of active sessions
	metrics.ActiveSessions.Inc()
	defer metrics.ActiveSessions.Dec()

	pub := subscribe.NewFilterPublisher(subscribe.NewStreamPublisher(stream))

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
func (s *Service) handle(stream grpc.BidiStreamingServer[protobuf.SubscribeRequest, protobuf.Message]) (subscribe.FilterFunc, error) {
	// wait for the request
	req, err := stream.Recv()
	if err != nil {
		if status.Code(err) == codes.Canceled {
			return nil, context.Canceled
		}

		return nil, status.Error(codes.Unknown, errors.Wrap(err, "unexpected client error").Error())
	}

	// validate the request
	//nolint:gomnd
	var filters = make([]subscribe.FilterFunc, 0, 3)

	if offset := req.GetMinOffset(); offset > 0 {
		offsetFilter := subscribe.FilterMinOffset(offset)
		filters = append(filters, offsetFilter)
	}

	if age := req.GetMaxAge(); age != "" {
		ageFilter, err := subscribe.FilterMaxAge(age)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		filters = append(filters, ageFilter)
	}

	keyFilter, err := subscribe.FilterKeys(req.GetKeys())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	filters = append(filters, keyFilter)

	return subscribe.FilterAnd(filters...), nil
}
