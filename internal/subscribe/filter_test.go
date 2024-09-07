package subscribe_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

	Context("FilterFunc", func() {
		It("should filter messages with the minimum offset", func() {
			f := subscribe.FilterMinOffset(2)

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
					Offset:    3,
				}
				Ω(f(expected)).Should(BeTrue())
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
					Offset:    1,
				}
				Ω(f(expected)).Should(BeFalse())
			})

		})

		It("should return an error if provided an invalid key glob", func() {
			_, err := subscribe.FilterKeys([]string{"[]"})
			Ω(err).Should(HaveOccurred())
		})

		It("should filter messages that match the key glob", func() {
			f, err := subscribe.FilterKeys([]string{"hi", "by*", "*o"})
			Ω(err).ShouldNot(HaveOccurred())

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
					Offset:    3,
				}
				Ω(f(expected)).Should(BeTrue())
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
					Offset:    1,
				}
				Ω(f(expected)).Should(BeFalse())
			})
		})

		It("should return an error if the age is invalid", func() {
			_, err := subscribe.FilterMaxAge("foo")
			Ω(err).Should(HaveOccurred())
		})

		It("should filter messages that are less than the max age", func() {
			f, err := subscribe.FilterMaxAge("5s")
			Ω(err).ShouldNot(HaveOccurred())

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
					Offset:    3,
				}
				Ω(f(expected)).Should(BeTrue())
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
					Timestamp: time.Now().Add(-10 * time.Second),
					Topic:     "my.topic",
					Partition: 3,
					Offset:    1,
				}
				Ω(f(expected)).Should(BeFalse())
			})
		})

		It("should perform a logical AND of multiple filter functions", func() {
			msg := &subscribe.Message{
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
				Offset:    1,
			}

			f1 := mocks.NewMockFilterFunc(T)
			f2 := mocks.NewMockFilterFunc(T)

			f := subscribe.FilterAnd(f1.Execute, f2.Execute)

			f1.On("Execute", msg).Return(false).Once()
			Ω(f(msg)).Should(BeFalse())

			f1.On("Execute", msg).Return(true).Once()
			f2.On("Execute", msg).Return(false).Once()
			Ω(f(msg)).Should(BeFalse())

			f1.On("Execute", msg).Return(true).Once()
			f2.On("Execute", msg).Return(true).Once()
			Ω(f(msg)).Should(BeTrue())
		})
	})
})
