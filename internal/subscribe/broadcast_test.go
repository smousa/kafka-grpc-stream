package subscribe_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
)

var _ = Describe("Broadcast", func() {
	var (
		T          = GinkgoT()
		pub1, pub2 *mocks.MockPublisher
		b          *subscribe.Broadcast
	)

	BeforeEach(func() {
		pub1, pub2 = mocks.NewMockPublisher(T), mocks.NewMockPublisher(T)
		b = subscribe.NewBroadcast()
	})

	It("should broadcast messages even if there are no sessions", func(ctx SpecContext) {
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
		b.Publish(ctx, expected)
	})

	It("should broadcast messages to a single session", func(ctx SpecContext) {
		By("adding the session", func() {
			b.RegisterSession("a", pub1)
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
			pub1.On("Publish", ctx, expected).Once()
			b.Publish(ctx, expected)
		})

		By("removing the session", func() {
			b.UnregisterSession("a")
			expected := &subscribe.Message{
				Key:   "fuh",
				Value: []byte("baz"),
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
			b.Publish(ctx, expected)
		})
	})

	It("should broadcast messages to multiple sessions", func(ctx SpecContext) {
		By("adding a session", func() {
			b.RegisterSession("a", pub1)
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
			pub1.On("Publish", ctx, expected).Once()
			b.Publish(ctx, expected)
		})

		By("adding another session", func() {
			b.RegisterSession("b", pub2)
			expected := &subscribe.Message{
				Key:   "fuh",
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
			pub1.On("Publish", ctx, expected).Once()
			pub2.On("Publish", ctx, expected).Once()
			b.Publish(ctx, expected)
		})

		By("removing the first session", func() {
			b.UnregisterSession("a")
			expected := &subscribe.Message{
				Key:   "fou",
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
			pub2.On("Publish", ctx, expected).Once()
			b.Publish(ctx, expected)
		})

		By("removing the second session", func() {
			b.UnregisterSession("b")
			expected := &subscribe.Message{
				Key:   "fto",
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
			b.Publish(ctx, expected)
		})
	})

	It("should panic when trying to add the same session more than once", func() {
		b.RegisterSession("a", pub1)
		Ω(func() { b.RegisterSession("a", pub1) }).Should(Panic())
	})
})
