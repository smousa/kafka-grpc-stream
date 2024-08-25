package subscribe_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	//. "github.com/onsi/gomega"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
)

var _ = Describe("Filter", func() {

	var (
		T   = GinkgoT()
		pub *mocks.MockPublisher
		p   *subscribe.FilterPublisher
	)

	BeforeEach(func() {
		pub = mocks.NewMockPublisher(T)
		p = subscribe.NewFilterPublisher(pub)
	})

	It("should not deliver the message if a filter is not defined", func(ctx SpecContext) {
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
		p.Publish(ctx, expected)
	})

	It("should deliver the message only if the filter matches", func(ctx SpecContext) {
		p.SetFilter(func(m *subscribe.Message) bool {
			return m.Key == "foo"
		})
		By("having a match", func() {
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
			pub.On("Publish", ctx, expected).Once()
			p.Publish(ctx, expected)
		})

		By("not having a match", func() {
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
			p.Publish(ctx, expected)
		})
	})
})
