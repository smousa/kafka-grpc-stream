package subscribe_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/twmb/franz-go/pkg/kgo"
)

var _ = Describe("KafkaSubscriber", func() {
	var (
		T   = GinkgoT()
		cli *mocks.MockKafkaClient
		pub *mocks.MockPublisher
		sub *subscribe.KafkaSubscriber
	)

	BeforeEach(func() {
		cli = mocks.NewMockKafkaClient(T)
		pub = mocks.NewMockPublisher(T)
		sub = subscribe.NewKafkaSubscriber(cli, pub)
	})

	It("should return if the context is canceled", func(sCtx SpecContext) {
		ctx, cancel := context.WithTimeout(sCtx, time.Second)
		defer cancel()
		cli.On("PollFetches", ctx).Return(nil).WaitUntil(time.After(1250 * time.Millisecond))
		sub.Subscribe(ctx)
	}, SpecTimeout(5*time.Second))

	It("should publish a record", func(sCtx SpecContext) {
		expected := &subscribe.Message{
			Key:   "foo",
			Value: []byte("bar"),
			Headers: []subscribe.Header{
				{
					Key:   "aaa",
					Value: "bbb",
				},
			},
			Timestamp: time.Now(),
			Topic:     "my.topic",
			Partition: 3,
			Offset:    4,
		}

		ctx, cancel := context.WithCancel(sCtx)

		var mu sync.Mutex
		mu.Lock()
		cli.On("PollFetches", ctx).Return(kgo.Fetches{
			{
				Topics: []kgo.FetchTopic{
					{
						Topic: expected.Topic,
						Partitions: []kgo.FetchPartition{
							{
								Partition: expected.Partition,
								Records: []*kgo.Record{
									{
										Key:   []byte(expected.Key),
										Value: expected.Value,
										Headers: []kgo.RecordHeader{
											{
												Key:   expected.Headers[0].Key,
												Value: []byte(expected.Headers[0].Value),
											},
										},
										Timestamp: expected.Timestamp,
										Topic:     expected.Topic,
										Partition: expected.Partition,
										Offset:    expected.Offset,
										Context:   ctx,
									},
								},
							},
						},
					},
				},
			},
		}).Run(func(args mock.Arguments) {
			mu.Unlock()
		}).Once()
		cli.On("PollFetches", ctx).Return(nil).Run(func(args mock.Arguments) {
			<-ctx.Done()
		}).Once()
		pub.On("Publish", ctx, expected).Return().Once()

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			sub.Subscribe(ctx)
		}()

		mu.Lock()
		cancel()
	})
})
