package service_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/stretchr/testify/mock"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/service"
	"github.com/smousa/kafka-grpc-stream/internal/service"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
)

var _ = Describe("Publisher", func() {
	var (
		T      = GinkgoT()
		stream *mocks.MockSenderStream
		pub    *service.StreamPublisher
	)

	BeforeEach(func() {
		stream = mocks.NewMockSenderStream(T)
		pub = service.NewStreamPublisher("foo", stream)
	})

	It("should publish the message", func(ctx SpecContext) {
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

		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(nil).
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

		pub.Publish(ctx, msg)
	})

	It("should give up if it cannot publish the message", func(ctx SpecContext) {
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

		stream.On("Send", mock.AnythingOfType("*protobuf.Message")).
			Return(errors.New("error")).
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

		pub.Publish(ctx, msg)
	})
})
