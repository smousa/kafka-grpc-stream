package service_test

import (
	"context"
	"errors"
	"sync"
	"time"

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
		broadcast *subscribe.Broadcast
		svc       *service.Service
	)

	BeforeEach(func() {
		stream = mocks.NewMockBidiStream(T)
		broadcast = subscribe.NewBroadcast()
		svc = service.New(broadcast)
	})

	It("should exit if the client closes the connection", func(ctx SpecContext) {
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

		stream.On("Recv").Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).Once()
		Ω(svc.Subscribe(stream)).Should(Succeed())

		By("ensuring that the stream is no longer sending messages")
		broadcast.Publish(ctx, msg)
	})

	It("should return an error if the client raises an error", func(ctx SpecContext) {
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

		stream.On("Recv").Return(nil, status.Error(codes.Unknown, "error")).Once()
		Ω(svc.Subscribe(stream)).ShouldNot(Succeed())

		By("ensuring that the stream is no longer sending messages")
		broadcast.Publish(ctx, msg)
	})

	It("should deliver the message", func(ctx SpecContext) {
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

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"*"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				defer GinkgoRecover()
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				defer GinkgoRecover()
				Eventually(sendWait).Should(BeClosed())
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
				close(sendWait)
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())

			By("ensuring that the stream is no longer sending messages")
			broadcast.Publish(ctx, msg)
		}()
		Eventually(recvWait).Should(BeClosed())
		broadcast.Publish(ctx, msg)
	}, SpecTimeout(5*time.Second))

	It("should not panic if it cannot deliver the message", func(ctx SpecContext) {
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

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"*"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				defer GinkgoRecover()
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				defer GinkgoRecover()
				Eventually(sendWait).Should(BeClosed())
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
				close(sendWait)
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())

			By("ensuring that the stream is no longer sending messages")
			broadcast.Publish(ctx, msg)
		}()
		Eventually(recvWait).Should(BeClosed())
		broadcast.Publish(ctx, msg)
	})

	It("should return an error if the key filter cannot be parsed", func() {
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"[]"},
			}, nil).
			Once()
		Ω(svc.Subscribe(stream)).ShouldNot(Succeed())
	})

	It("should return records that have matching keys", func(ctx SpecContext) {
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

		var recvWait, sendWait = make(chan struct{}), make(chan struct{})
		stream.On("Recv").
			Return(&protobuf.SubscribeRequest{
				Keys: []string{"foo"},
			}, nil).
			Once().
			Run(func(mock.Arguments) {
				defer GinkgoRecover()
				close(recvWait)
			})
		stream.On("Recv").
			Return(nil, status.Error(codes.Canceled, context.Canceled.Error())).
			Once().
			Run(func(mock.Arguments) {
				defer GinkgoRecover()
				Eventually(sendWait).Should(BeClosed())
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
				close(sendWait)
			})

		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Ω(svc.Subscribe(stream)).Should(Succeed())

			By("ensuring that the stream is no longer sending messages")
			broadcast.Publish(ctx, msgs[0])
		}()
		Eventually(recvWait).Should(BeClosed())
		broadcast.Publish(ctx, msgs[1])
		broadcast.Publish(ctx, msgs[0])
	})
})
