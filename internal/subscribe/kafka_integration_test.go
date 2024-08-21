package subscribe_test

import (
	"context"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/twmb/franz-go/pkg/kgo"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("KafkaSubscriber", Label("integration"), func() {
	var (
		T   = GinkgoT()
		pub *mocks.MockPublisher
		sub *subscribe.KafkaSubscriber
	)

	BeforeEach(func() {
		pub = mocks.NewMockPublisher(T)
		sub = subscribe.NewKafkaSubscriber(kafkaClient, pub)
	})

	It("should return if the context is canceled", func(sCtx SpecContext) {
		ctx, cancel := context.WithTimeout(sCtx, time.Second)
		defer cancel()
		sub.Subscribe(ctx)
	}, SpecTimeout(5*time.Second))

	It("should publish a record", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			sub.Subscribe(ctx)
		}()

		car := gofakeit.Car()
		record := &kgo.Record{
			Key:   []byte(car.Brand),
			Value: []byte(car.Model),
			Headers: []kgo.RecordHeader{
				{
					Key:   "type",
					Value: []byte(car.Type),
				},
				{
					Key:   "fuel",
					Value: []byte(car.Fuel),
				},
				{
					Key:   "transmission",
					Value: []byte(car.Transmission),
				},
			},
			Topic:   topic,
			Context: sCtx,
		}

		pub.On("Publish", ctx, mock.AnythingOfType("*subscribe.Message")).Once().Run(func(args mock.Arguments) {
			defer cancel()
			defer GinkgoRecover()
			Ω(args.Get(1)).Should(PointTo(MatchAllFields(Fields{
				"Key":   BeEquivalentTo(record.Key),
				"Value": BeEquivalentTo(record.Value),
				"Headers": Equal([]subscribe.Header{
					{
						Key:   "type",
						Value: car.Type,
					},
					{
						Key:   "fuel",
						Value: car.Fuel,
					},
					{
						Key:   "transmission",
						Value: car.Transmission,
					},
				}),
				"Timestamp": BeTemporally("~", time.Now(), time.Second),
				"Topic":     Equal(topic),
				"Partition": BeEquivalentTo(0),
				"Offset":    BeNumerically(">=", 0),
			})))
		})

		Ω(kafkaClient.ProduceSync(ctx, record).FirstErr()).ShouldNot(HaveOccurred())
	}, SpecTimeout(5*time.Second))

})
