package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gstruct"

	"github.com/twmb/franz-go/pkg/kgo"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ = PDescribe("Registry", func() {
	var (
		session *Session
		topic   string
	)

	BeforeEach(func() {
		topic = "e2e.registry." + gofakeit.LetterN(12)

		// wait for the server to join the registry
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		hostPath := path.Join("/t", topic, "p", "0", "host") + "/"
		hostCh := etcdClient.Watch(ctx, hostPath, clientv3.WithPrefix())

		var err error
		cmd := exec.Command(serverExe)
		cmd.Env = append(os.Environ(),
			"LISTEN_URL=localhost:9000",
			"LISTEN_ADVERTISE_URL=localhost:9000",
			"SERVER_BROADCAST_BUFFER_SIZE=5",
			"WORKER_TOPIC="+topic,
			"WORKER_PARTITION=0",
		)
		session, err = Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		Eventually(hostCh).Should(Receive())
	})

	AfterEach(func() {
		if session != nil {
			// wait for the server to leave the registry
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			hostPath := path.Join("/t", topic, "p", "0", "host") + "/"
			hostCh := etcdClient.Watch(ctx, hostPath, clientv3.WithPrefix())

			session.Terminate()
			Eventually(session).Should(Exit())
			Eventually(hostCh).WithTimeout(10 * time.Second).Should(Receive())
		}
		etcdClient.Delete(context.TODO(), "/t/"+topic, clientv3.WithPrefix())
	})

	It("should register all the new keys seen on the partition", func(ctx SpecContext) {
		keyPath := path.Join("/t", topic, "k") + "/"
		keyCh := etcdClient.Watch(ctx, keyPath, clientv3.WithPrefix())
		records := []*kgo.Record{
			{
				Key:   []byte("key1"),
				Topic: topic,
			},
			{
				Key:   []byte("key2"),
				Topic: topic,
			},
			{
				Key:   []byte("key2"),
				Topic: topic,
			},
			{
				Key:   []byte("key3"),
				Topic: topic,
			},
		}
		Ω(kafkaClient.ProduceSync(ctx, records...).FirstErr()).ShouldNot(HaveOccurred())
		Eventually(keyCh).Should(Receive())
		Eventually(keyCh).Should(Receive())
		Eventually(keyCh).Should(Receive())

		keyResp, err := etcdClient.Get(ctx, keyPath,
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(keyResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(3),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(path.Join(keyPath, "key1")),
					"Value": BeEquivalentTo("0"),
				})),
				"1": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(path.Join(keyPath, "key2")),
					"Value": BeEquivalentTo("0"),
				})),
				"2": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key":   BeEquivalentTo(path.Join(keyPath, "key3")),
					"Value": BeEquivalentTo("0"),
				})),
			}),
		})))
	})
})
