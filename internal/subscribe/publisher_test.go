package subscribe_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	//. "github.com/onsi/gomega"

	mocks "github.com/smousa/kafka-grpc-stream/internal/mocks/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
)

var _ = Describe("Publisher", func() {
	var (
		T          = GinkgoT()
		pub1, pub2 *mocks.MockPublisher
	)

	BeforeEach(func() {
		pub1, pub2 = mocks.NewMockPublisher(T), mocks.NewMockPublisher(T)
	})

	It("should chain publishers together", func(ctx SpecContext) {
		p := subscribe.Publishers{pub1, pub2}
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
		pub2.On("Publish", ctx, expected).Once()
		p.Publish(ctx, expected)
	})
})
