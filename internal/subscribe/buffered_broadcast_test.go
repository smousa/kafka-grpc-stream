package subscribe_test

import (
	"context"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
)

var _ = Describe("BufferedBroadcast", func() {
	var (
		T  = GinkgoT()
		bb *subscribe.BufferedBroadcast
	)

	BeforeEach(func() {
		bb = subscribe.NewBufferedBroadcast(5)
	})

	It("should retain messages that are ready to be consumed", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		pub := mocks.NewMockPublisher(T)

		msg := &subscribe.Message{
			Key:       gofakeit.ProductName(),
			Value:     []byte(gofakeit.ProductFeature()),
			Timestamp: time.Now(),
			Topic:     "products",
			Partition: 1,
			Offset:    2,
		}
		bb.Publish(sCtx, msg)

		pub.On("Publish", ctx, msg).
			Once().
			Run(func(args mock.Arguments) {
				cancel()
			})
		bb.RegisterSession(ctx, pub)
		bb.Publish(sCtx, msg)
	}, SpecTimeout(5*time.Second))

	It("should wait for a message to be ready", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		msg := &subscribe.Message{
			Key:       gofakeit.ProductName(),
			Value:     []byte(gofakeit.ProductFeature()),
			Timestamp: time.Now(),
			Topic:     "products",
			Partition: 1,
			Offset:    2,
		}

		pub := mocks.NewMockPublisher(T)
		pub.On("Publish", ctx, msg).
			Once().
			Run(func(args mock.Arguments) {
				cancel()
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			bb.RegisterSession(ctx, pub)
		}()

		time.Sleep(500 * time.Millisecond)
		bb.Publish(sCtx, msg)
	}, SpecTimeout(5*time.Second))

	It("should skip messages that it cannot consume in time", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		msgs := []*subscribe.Message{
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    2,
			},
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    3,
			},
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    4,
			},
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    5,
			},
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    6,
			},
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    7,
			},
			{
				Key:       gofakeit.ProductName(),
				Value:     []byte(gofakeit.ProductFeature()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    8,
			},
		}

		bb.Publish(sCtx, msgs[0])
		bb.Publish(sCtx, msgs[1])
		bb.Publish(sCtx, msgs[2])
		bb.Publish(sCtx, msgs[3])
		bb.Publish(sCtx, msgs[4])

		wait, done := make(chan struct{}), make(chan struct{})
		pub := mocks.NewMockPublisher(T)
		pub.On("Publish", ctx, msgs[0]).Once().
			Run(func(args mock.Arguments) {
				<-wait
			})
		pub.On("Publish", ctx, msgs[2]).Once()
		pub.On("Publish", ctx, msgs[3]).Once()
		pub.On("Publish", ctx, msgs[4]).Once()
		pub.On("Publish", ctx, msgs[5]).Once()
		pub.On("Publish", ctx, msgs[6]).Once().
			Run(func(args mock.Arguments) {
				close(done)
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			bb.RegisterSession(ctx, pub)
		}()

		time.Sleep(500 * time.Millisecond)
		bb.Publish(sCtx, msgs[5])
		bb.Publish(sCtx, msgs[6])
		close(wait)
		Eventually(done).Should(BeClosed())
		cancel()

	}, SpecTimeout(5*time.Second))
})
