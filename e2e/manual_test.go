package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gstruct"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/smousa/kafka-grpc-stream/protobuf"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("Manual", func() {
	var (
		session *Session
		conn    *grpc.ClientConn
		cli     protobuf.KafkaStreamerClient
		topic   string
	)

	BeforeEach(func() {
		topic = "e2e.manual." + gofakeit.LetterN(12)

		var err error
		cmd := exec.Command(serverExe)
		cmd.Env = append(os.Environ(),
			"LISTEN_URL=localhost:9000",
			"LISTEN_ADVERTISE_URL=localhost:9000",
			"WORKER_TOPIC="+topic,
			"WORKER_PARTITION=0",
		)
		session, err = Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		time.Sleep(500 * time.Millisecond)
		conn, err = grpc.NewClient("localhost:9000",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		Ω(err).ShouldNot(HaveOccurred())
		cli = protobuf.NewKafkaStreamerClient(conn)
	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
		if session != nil {
			session.Terminate()
			Eventually(session).Should(Exit())
		}
		etcdClient.Delete(context.TODO(), "/topics/"+topic, clientv3.WithPrefix())
	})

	It("should shut down when there are no active connections", func(ctx SpecContext) {
		_, err := cli.Subscribe(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		session.Terminate()
		Consistently(session).ShouldNot(Exit())
	})

	It("should subscribe to the topic", func(ctx SpecContext) {
		stream, err := cli.Subscribe(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		req := &protobuf.SubscribeRequest{
			Keys: []string{"*"},
		}
		Ω(stream.Send(req)).Should(Succeed())

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
			Topic: topic,
		}
		Ω(kafkaClient.ProduceSync(ctx, record).FirstErr()).ShouldNot(HaveOccurred())
		Ω(stream.Recv()).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Key":   BeEquivalentTo(record.Key),
			"Value": BeEquivalentTo(record.Value),
			"Headers": Equal([]*protobuf.Header{
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
			"Timestamp": BeNumerically("~", time.Now().UnixMilli(), 10000),
			"Topic":     Equal(topic),
			"Partition": BeEquivalentTo(0),
			"Offset":    BeNumerically(">=", 0),
		})))
	}, SpecTimeout(10*time.Second))

	It("should filter messages with matching keys", func(ctx SpecContext) {
		stream, err := cli.Subscribe(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		req := &protobuf.SubscribeRequest{
			Keys: []string{"second"},
		}
		Ω(stream.Send(req)).Should(Succeed())

		car1 := gofakeit.Car()
		car2 := gofakeit.Car()

		records := []*kgo.Record{
			{
				Key:   []byte("first"),
				Value: []byte(car1.Model),
				Headers: []kgo.RecordHeader{
					{
						Key:   "type",
						Value: []byte(car1.Type),
					},
					{
						Key:   "fuel",
						Value: []byte(car1.Fuel),
					},
					{
						Key:   "transmission",
						Value: []byte(car1.Transmission),
					},
				},
				Topic: topic,
			}, {
				Key:   []byte("second"),
				Value: []byte(car2.Model),
				Headers: []kgo.RecordHeader{
					{
						Key:   "type",
						Value: []byte(car2.Type),
					},
					{
						Key:   "fuel",
						Value: []byte(car2.Fuel),
					},
					{
						Key:   "transmission",
						Value: []byte(car2.Transmission),
					},
				},
				Topic: topic,
			},
		}

		Ω(kafkaClient.ProduceSync(ctx, records...).FirstErr()).ShouldNot(HaveOccurred())
		Ω(stream.Recv()).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Key":   BeEquivalentTo(records[1].Key),
			"Value": BeEquivalentTo(records[1].Value),
			"Headers": Equal([]*protobuf.Header{
				{
					Key:   "type",
					Value: car2.Type,
				},
				{
					Key:   "fuel",
					Value: car2.Fuel,
				},
				{
					Key:   "transmission",
					Value: car2.Transmission,
				},
			}),
			"Timestamp": BeNumerically("~", time.Now().UnixMilli(), 10000),
			"Topic":     Equal(topic),
			"Partition": BeEquivalentTo(0),
			"Offset":    BeNumerically(">=", 0),
		})))
	}, SpecTimeout(10*time.Second))
})
