package service

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/smousa/kafka-grpc-stream/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SenderStream interface {
	grpc.ServerStreamingServer[protobuf.Message]
}

type StreamPublisher struct {
	sessionId string
	stream    SenderStream
}

func NewStreamPublisher(sessionId string, stream SenderStream) *StreamPublisher {
	return &StreamPublisher{sessionId, stream}
}

func (s *StreamPublisher) Publish(ctx context.Context, message *subscribe.Message) {
	headers := make([]*protobuf.Header, len(message.Headers))
	for i, h := range message.Headers {
		headers[i] = &protobuf.Header{
			Key:   h.Key,
			Value: h.Value,
		}
	}

	err := s.stream.Send(&protobuf.Message{
		Key:       message.Key,
		Value:     message.Value,
		Headers:   headers,
		Timestamp: message.Timestamp.UnixMilli(),
		Topic:     message.Topic,
		Partition: message.Partition,
		Offset:    message.Offset,
	})

	if err != nil && status.Code(err) != codes.Canceled {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("session_id", s.sessionId).
			Int64("offset", message.Offset).
			Msg("Error publishing message")
	}
}
