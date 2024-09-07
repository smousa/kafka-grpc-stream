package service_test

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/service"
	"github.com/smousa/kafka-grpc-stream/internal/service"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/smousa/kafka-grpc-stream/protobuf"
)

var _ = Describe("Service", func() {
	var (
		T         = GinkgoT()
		stream    *mocks.MockBidiStream
		broadcast *mocks.MockSessionRegistry
		svc       *service.Service
	)

	BeforeEach(func() {
		stream = mocks.NewMockBidiStream(T)
		broadcast = mocks.NewMockSessionRegistry(T)
		svc = service.New(broadcast)
	})

	It("should exit if the client closes the connection", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		stream.On("Context").Return(ctx).Maybe()
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()

		cancel()
	}, SpecTimeout(5*time.Second))

	It("should return an error if the receiver raises an error", func(ctx SpecContext) {
		stream.On("Context").Return(ctx).Maybe()
		stream.On("Recv").
			Return(nil, status.Error(codes.Unknown, "error")).
			Once()

		Ω(svc.Subscribe(stream)).ShouldNot(Succeed())
	})

	It("should return an error if the receiver eventually raises an error", func(ctx SpecContext) {
		broadcast.On("RegisterSession", mock.Anything, mock.Anything).Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())
				<-ctx.Done()
			})

		stream.On("Context").Return(ctx).Maybe()
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"*"},
			}, nil).
			Once()
		stream.On("Recv").
			Return(nil, status.Error(codes.Unknown, "error")).
			Once()

		Ω(svc.Subscribe(stream)).ShouldNot(Succeed())
	}, SpecTimeout(5*time.Second))

	It("should deliver the message", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})

		msg := &subscribe.Message{
			Key:   "foo",
			Value: []byte("bar"),
			Headers: []subscribe.Header{
				{
					Key:   "hi",
					Value: "bye",
				},
			},
			Timestamp: time.Now(),
			Topic:     "my.topic",
			Partition: 1,
			Offset:    20,
		}

		broadcast.On("RegisterSession", mock.Anything, mock.Anything).Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())

				pub, ok := args.Get(1).(subscribe.Publisher)
				Ω(ok).Should(BeTrue())

				Eventually(recvWait).Should(BeClosed())
				pub.Publish(ctx, msg)
				close(sendWait)
				<-ctx.Done()
			})
		stream.On("Context").Return(ctx)
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"*"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(nil).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(msg.Key),
					"Value": BeEquivalentTo(msg.Value),
					"Headers": MatchAllElementsWithIndex(IndexIdentity, Elements{
						"0": PointTo(MatchFields(IgnoreExtras, Fields{
							"Key":   Equal("hi"),
							"Value": Equal("bye"),
						})),
					}),
					"Timestamp": Equal(msg.Timestamp.UnixMilli()),
					"Topic":     Equal(msg.Topic),
					"Partition": Equal(msg.Partition),
					"Offset":    Equal(msg.Offset),
				})))
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()

		Eventually(sendWait).Should(BeClosed())
		cancel()
	}, SpecTimeout(5*time.Second))

	It("should not panic if it cannot deliver the message", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})

		msg := &subscribe.Message{
			Key:   "foo",
			Value: []byte("bar"),
			Headers: []subscribe.Header{
				{
					Key:   "hi",
					Value: "bye",
				},
			},
			Timestamp: time.Now(),
			Topic:     "my.topic",
			Partition: 1,
			Offset:    20,
		}

		broadcast.On("RegisterSession", mock.Anything, mock.Anything).Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())

				pub, ok := args.Get(1).(subscribe.Publisher)
				Ω(ok).Should(BeTrue())

				Eventually(recvWait).Should(BeClosed())
				pub.Publish(ctx, msg)
				close(sendWait)
				<-ctx.Done()
			})
		stream.On("Context").Return(ctx)
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"*"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(msg.Key),
					"Value": BeEquivalentTo(msg.Value),
					"Headers": MatchAllElementsWithIndex(IndexIdentity, Elements{
						"0": PointTo(MatchFields(IgnoreExtras, Fields{
							"Key":   Equal("hi"),
							"Value": Equal("bye"),
						})),
					}),
					"Timestamp": Equal(msg.Timestamp.UnixMilli()),
					"Topic":     Equal(msg.Topic),
					"Partition": Equal(msg.Partition),
					"Offset":    Equal(msg.Offset),
				})))
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()

		Eventually(sendWait).Should(BeClosed())
		cancel()
	}, SpecTimeout(5*time.Second))

	It("should return an error if the key filter cannot be parsed", func(ctx SpecContext) {
		stream.On("Context").Return(ctx).Maybe()
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"[]"},
			}, nil).
			Once()
		Ω(svc.Subscribe(stream)).ShouldNot(Succeed())
	}, SpecTimeout(5*time.Second))

	It("should return an error if the age filter cannot be parsed", func(ctx SpecContext) {
		stream.On("Context").Return(ctx).Maybe()
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys:   []string{"*"},
				MaxAge: "foo",
			}, nil).
			Once()
		Ω(svc.Subscribe(stream)).ShouldNot(Succeed())
	})

	It("should return records that have matching keys", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})

		msgs := []*subscribe.Message{
			{
				Key:   "foo",
				Value: []byte("bar"),
				Headers: []subscribe.Header{
					{
						Key:   "hi",
						Value: "bye",
					},
				},
				Timestamp: time.Now(),
				Topic:     "my.topic",
				Partition: 1,
				Offset:    20,
			}, {
				Key:   "fuh",
				Value: []byte("bar"),
				Headers: []subscribe.Header{
					{
						Key:   "hi",
						Value: "bye",
					},
				},
				Timestamp: time.Now(),
				Topic:     "my.topic",
				Partition: 1,
				Offset:    20,
			},
		}

		broadcast.On("RegisterSession", mock.Anything, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())

				pub, ok := args.Get(1).(subscribe.Publisher)
				Ω(ok).Should(BeTrue())

				Eventually(recvWait).Should(BeClosed())
				for _, msg := range msgs {
					pub.Publish(ctx, msg)
				}
				close(sendWait)
				<-ctx.Done()
			})
		stream.On("Context").Return(ctx)
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"foo"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(msgs[0].Key),
					"Value": BeEquivalentTo(msgs[0].Value),
					"Headers": MatchAllElementsWithIndex(IndexIdentity, Elements{
						"0": PointTo(MatchFields(IgnoreExtras, Fields{
							"Key":   Equal("hi"),
							"Value": Equal("bye"),
						})),
					}),
					"Timestamp": Equal(msgs[0].Timestamp.UnixMilli()),
					"Topic":     Equal(msgs[0].Topic),
					"Partition": Equal(msgs[0].Partition),
					"Offset":    Equal(msgs[0].Offset),
				})))
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()
		Eventually(sendWait).Should(BeClosed())
		cancel()
	}, SpecTimeout(5*time.Second))

	It("should return records with a matching min offset", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})

		msgs := []*subscribe.Message{
			{
				Key:   "foo",
				Value: []byte("bar"),
				Headers: []subscribe.Header{
					{
						Key:   "hi",
						Value: "bye",
					},
				},
				Timestamp: time.Now(),
				Topic:     "my.topic",
				Partition: 1,
				Offset:    2,
			}, {
				Key:   "fuh",
				Value: []byte("bar"),
				Headers: []subscribe.Header{
					{
						Key:   "hi",
						Value: "bye",
					},
				},
				Timestamp: time.Now(),
				Topic:     "my.topic",
				Partition: 1,
				Offset:    4,
			},
		}

		broadcast.On("RegisterSession", mock.Anything, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())

				pub, ok := args.Get(1).(subscribe.Publisher)
				Ω(ok).Should(BeTrue())

				Eventually(recvWait).Should(BeClosed())
				for _, msg := range msgs {
					pub.Publish(ctx, msg)
				}
				close(sendWait)
				<-ctx.Done()
			})
		stream.On("Context").Return(ctx)
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys:      []string{"*"},
				MinOffset: 3,
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(msgs[1].Key),
					"Value": BeEquivalentTo(msgs[1].Value),
					"Headers": MatchAllElementsWithIndex(IndexIdentity, Elements{
						"0": PointTo(MatchFields(IgnoreExtras, Fields{
							"Key":   Equal("hi"),
							"Value": Equal("bye"),
						})),
					}),
					"Timestamp": Equal(msgs[1].Timestamp.UnixMilli()),
					"Topic":     Equal(msgs[1].Topic),
					"Partition": Equal(msgs[1].Partition),
					"Offset":    Equal(msgs[1].Offset),
				})))
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()
		Eventually(sendWait).Should(BeClosed())
		cancel()
	}, SpecTimeout(5*time.Second))

	It("should return records with a matching max age", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})

		msgs := []*subscribe.Message{
			{
				Key:   "foo",
				Value: []byte("bar"),
				Headers: []subscribe.Header{
					{
						Key:   "hi",
						Value: "bye",
					},
				},
				Timestamp: time.Now().Add(-10 * time.Second),
				Topic:     "my.topic",
				Partition: 1,
				Offset:    2,
			}, {
				Key:   "fuh",
				Value: []byte("bar"),
				Headers: []subscribe.Header{
					{
						Key:   "hi",
						Value: "bye",
					},
				},
				Timestamp: time.Now(),
				Topic:     "my.topic",
				Partition: 1,
				Offset:    4,
			},
		}

		broadcast.On("RegisterSession", mock.Anything, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())

				pub, ok := args.Get(1).(subscribe.Publisher)
				Ω(ok).Should(BeTrue())

				Eventually(recvWait).Should(BeClosed())
				for _, msg := range msgs {
					pub.Publish(ctx, msg)
				}
				close(sendWait)
				<-ctx.Done()
			})
		stream.On("Context").Return(ctx)
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys:   []string{"*"},
				MaxAge: "5s",
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(msgs[1].Key),
					"Value": BeEquivalentTo(msgs[1].Value),
					"Headers": MatchAllElementsWithIndex(IndexIdentity, Elements{
						"0": PointTo(MatchFields(IgnoreExtras, Fields{
							"Key":   Equal("hi"),
							"Value": Equal("bye"),
						})),
					}),
					"Timestamp": Equal(msgs[1].Timestamp.UnixMilli()),
					"Topic":     Equal(msgs[1].Topic),
					"Partition": Equal(msgs[1].Partition),
					"Offset":    Equal(msgs[1].Offset),
				})))
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()
		Eventually(sendWait).Should(BeClosed())
		cancel()

	}, SpecTimeout(5*time.Second))

	It("should update the filter", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		var recvWait, sendWait = make(chan struct{}), make(chan struct{})

		msgs := []*subscribe.Message{
			{
				Key:       "A",
				Value:     []byte(gofakeit.ProductName()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    1,
			},
			{
				Key:       "B",
				Value:     []byte(gofakeit.ProductName()),
				Timestamp: time.Now(),
				Topic:     "products",
				Partition: 1,
				Offset:    2,
			},
		}

		broadcast.On("RegisterSession", mock.Anything, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				ctx, ok := args.Get(0).(context.Context)
				Ω(ok).Should(BeTrue())

				pub, ok := args.Get(1).(subscribe.Publisher)
				Ω(ok).Should(BeTrue())

				for _, msg := range msgs {
					pub.Publish(ctx, msg)
				}
				Eventually(recvWait).Should(Receive())
				time.Sleep(100 * time.Millisecond)
				for _, msg := range msgs {
					pub.Publish(ctx, msg)
				}
				close(sendWait)
				<-ctx.Done()
			})
		stream.On("Context").Return(ctx)
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"A"},
			}, nil).
			Once()
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"B"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				recvWait <- struct{}{}
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				<-ctx.Done()
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":       BeEquivalentTo(msgs[0].Key),
					"Value":     BeEquivalentTo(msgs[0].Value),
					"Headers":   BeEmpty(),
					"Timestamp": Equal(msgs[0].Timestamp.UnixMilli()),
					"Topic":     Equal(msgs[0].Topic),
					"Partition": Equal(msgs[0].Partition),
					"Offset":    Equal(msgs[0].Offset),
				})))
			})
		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
			Once().
			Run(func(args mock.Arguments) {
				defer GinkgoRecover()
				Ω(args.Get(0)).Should(PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":       BeEquivalentTo(msgs[1].Key),
					"Value":     BeEquivalentTo(msgs[1].Value),
					"Headers":   BeEmpty(),
					"Timestamp": Equal(msgs[1].Timestamp.UnixMilli()),
					"Topic":     Equal(msgs[1].Topic),
					"Partition": Equal(msgs[1].Partition),
					"Offset":    Equal(msgs[1].Offset),
				})))
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())
		}()
		Eventually(sendWait).Should(BeClosed())
		cancel()
	}, SpecTimeout(5*time.Second))
})
