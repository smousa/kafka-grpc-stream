package subscribe

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/smousa/kafka-grpc-stream/internal/metrics"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaClient interface {
	PollFetches(ctx context.Context) kgo.Fetches
}

type KafkaSubscriber struct {
	cli KafkaClient
	pub Publisher
}

func NewKafkaSubscriber(cli KafkaClient, pub Publisher) *KafkaSubscriber {
	return &KafkaSubscriber{cli, pub}
}

func (k *KafkaSubscriber) Subscribe(ctx context.Context) {
	log := zerolog.Ctx(ctx)

	for {
		fetches := k.cli.PollFetches(ctx)

		if ctx.Err() != nil {
			return
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			log.Error().
				Err(err).
				Str("topic", topic).
				Int32("partition", partition).
				Msg("Unretryable error when fetching records")
		})

		fetches.EachRecord(func(record *kgo.Record) {
			metrics.KafkaMaxOffset.Set(float64(record.Offset))

			headers := make([]Header, len(record.Headers))
			for i, h := range record.Headers {
				headers[i] = Header{h.Key, string(h.Value)}
			}

			k.pub.Publish(ctx, &Message{
				Key:       string(record.Key),
				Value:     record.Value,
				Headers:   headers,
				Timestamp: record.Timestamp,
				Topic:     record.Topic,
				Partition: record.Partition,
				Offset:    record.Offset,
			})
		})
	}
}
