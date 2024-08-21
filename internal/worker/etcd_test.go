package worker_test

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"github.com/smousa/kafka-grpc-stream/internal/worker"
)

var _ = Describe("EtcdRegistry", func() {

	var reg *worker.EtcdRegistry

	BeforeEach(func(ctx SpecContext) {
		reg = worker.NewEtcdRegistry(worker.WithEtcdClient(etcdClient))

		// delete all the resources
		_, err := etcdClient.Delete(ctx, "/", clientv3.WithPrefix())
		Ω(err).ShouldNot(HaveOccurred())
	}, NodeTimeout(5*time.Second))

	It("should register itself within the cluster", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		var w = worker.Worker{
			HostId:    "worker-0",
			HostAddr:  "127.0.0.1:1234",
			Topic:     "foo",
			Partition: "0",
		}
		watchCh := etcdClient.Watch(ctx, "/topics/foo/0/hosts/", clientv3.WithPrefix())
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Ω(reg.Register(ctx, w, 15)).Should(Succeed())
		}()
		Eventually(watchCh).Should(Receive())

		hostResp, err := etcdClient.Get(ctx, "/topics/foo/0/hosts/worker-0")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(hostResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(1),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/hosts/worker-0"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
			}),
		})))

		cancel()
	}, SpecTimeout(30*time.Second))

	It("assigns sessions when a host becomes available", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		var w = worker.Worker{
			HostId:    "worker-0",
			HostAddr:  "127.0.0.1:1234",
			Topic:     "foo",
			Partition: "0",
		}

		By("adding a session before the host is added", func() {
			_, err := etcdClient.Put(ctx, "/topics/foo/0/sessions/session-0", "")
			Ω(err).ShouldNot(HaveOccurred())
		})

		watchCh := etcdClient.Watch(ctx, "/topics/foo/0/routes/", clientv3.WithPrefix())
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Ω(reg.Register(ctx, w, 15)).Should(Succeed())
		}()
		Eventually(watchCh).Should(Receive())

		routeResp, err := etcdClient.Get(ctx, "/topics/foo/0/routes/",
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(routeResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(1),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-0"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
			}),
		})))

		By("adding a session after the host is added", func() {
			_, err := etcdClient.Put(ctx, "/topics/foo/0/sessions/session-1", "")
			Ω(err).ShouldNot(HaveOccurred())
		})
		Eventually(watchCh).Should(Receive())

		routeResp, err = etcdClient.Get(ctx, "/topics/foo/0/routes/",
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(routeResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(2),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-0"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
				"1": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-1"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
			}),
		})))

		cancel()
	}, SpecTimeout(30*time.Second))

	It("should should only rebalance if it is the leader", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		var w = worker.Worker{
			HostId:    "worker-0",
			HostAddr:  "127.0.0.1:1234",
			Topic:     "foo",
			Partition: "0",
		}

		By("adding a host and session before the host is added", func() {
			_, err := etcdClient.Put(ctx, "/topics/foo/0/hosts/hosts-foo", "")
			Ω(err).ShouldNot(HaveOccurred())
			_, err = etcdClient.Put(ctx, "/topics/foo/0/sessions/session-0", "")
			Ω(err).ShouldNot(HaveOccurred())
		})

		watchCh := etcdClient.Watch(ctx, "/topics/foo/0/routes/", clientv3.WithPrefix())
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Ω(reg.Register(ctx, w, 15)).Should(Succeed())
		}()
		Consistently(watchCh).ShouldNot(Receive())

		By("removing the leader", func() {
			_, err := etcdClient.Delete(ctx, "/topics/foo/0/hosts/hosts-foo")
			Ω(err).ShouldNot(HaveOccurred())
		})
		Eventually(watchCh).Should(Receive())

		routeResp, err := etcdClient.Get(ctx, "/topics/foo/0/routes/",
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(routeResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(1),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-0"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
			}),
		})))
		cancel()
	}, SpecTimeout(30*time.Second))

	It("should rebalance if the host or session list changes", func(sCtx SpecContext) {
		ctx, cancel := context.WithCancel(sCtx)
		var w = worker.Worker{
			HostId:    "worker-0",
			HostAddr:  "127.0.0.1:1234",
			Topic:     "foo",
			Partition: "0",
		}

		By("adding sessions", func() {
			_, err := etcdClient.Put(ctx, "/topics/foo/0/sessions/session-0", "")
			Ω(err).ShouldNot(HaveOccurred())
			_, err = etcdClient.Put(ctx, "/topics/foo/0/sessions/session-1", "")
			Ω(err).ShouldNot(HaveOccurred())
		})

		watchCh := etcdClient.Watch(ctx, "/topics/foo/0/routes/", clientv3.WithPrefix())
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Ω(reg.Register(ctx, w, 15)).Should(Succeed())
		}()
		Eventually(watchCh).Should(Receive())

		routeResp, err := etcdClient.Get(ctx, "/topics/foo/0/routes/",
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(routeResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(2),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-0"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
				"1": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-1"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
			}),
		})))

		By("adding a new host")

		var w2 = worker.Worker{
			HostId:    "worker-1",
			HostAddr:  "127.0.0.1:1234",
			Topic:     "foo",
			Partition: "0",
		}
		w2bytes, err := json.Marshal(w2)
		Ω(err).ShouldNot(HaveOccurred())
		_, err = etcdClient.Put(ctx, "/topics/foo/0/hosts/worker-1", string(w2bytes))
		Ω(err).ShouldNot(HaveOccurred())
		Eventually(watchCh).Should(Receive())

		routeResp, err = etcdClient.Get(ctx, "/topics/foo/0/routes/",
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(routeResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(2),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-0"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
				"1": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-1"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w2)),
				})),
			}),
		})))

		By("removing a session", func() {
			_, err := etcdClient.Delete(ctx, "/topics/foo/0/sessions/session-0")
			Ω(err).ShouldNot(HaveOccurred())
		})
		Eventually(watchCh).Should(Receive())

		routeResp, err = etcdClient.Get(ctx, "/topics/foo/0/routes/",
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(routeResp).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Count": BeEquivalentTo(1),
			"Kvs": MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": PointTo(MatchFields(IgnoreExtras, Fields{
					"Key": BeEquivalentTo("/topics/foo/0/routes/session-1"),
					"Value": WithTransform(func(v []byte) (worker.Worker, error) {
						w := worker.Worker{}
						err := json.Unmarshal(v, &w)
						return w, err
					}, Equal(w)),
				})),
			}),
		})))

		cancel()
	}, SpecTimeout(30*time.Second))

	Context("registering the partition key", func() {

		It("should not create the partition key if it is not the leader", func(ctx SpecContext) {
			reg.Publish(ctx, &subscribe.Message{Topic: "foo", Partition: 0, Key: "abc"})
			res, err := etcdClient.Get(ctx, "/topics/foo/keys/abc")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res.Count).Should(BeZero())
		})

		It("should not update the partition key if it is not the leader", func(ctx SpecContext) {
			_, err := etcdClient.Put(ctx, "/topics/foo/keys/abc", "1")
			Ω(err).ShouldNot(HaveOccurred())
			reg.Publish(ctx, &subscribe.Message{Topic: "foo", Partition: 0, Key: "abc"})
			res, err := etcdClient.Get(ctx, "/topics/foo/keys/abc")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res.Kvs[0].Value).Should(BeEquivalentTo("1"))
		})

		It("should not update the partition key if the value hasn't changed", func(sCtx SpecContext) {
			ctx, cancel := context.WithCancel(sCtx)

			var w = worker.Worker{
				HostId:    "worker-0",
				HostAddr:  "127.0.0.1:1234",
				Topic:     "foo",
				Partition: "0",
			}
			watchCh := etcdClient.Watch(ctx, "/topics/foo/0/routes/", clientv3.WithPrefix())
			var wg sync.WaitGroup
			defer wg.Wait()
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				Ω(reg.Register(ctx, w, 15)).Should(Succeed())
			}()
			By("adding sessions", func() {
				_, err := etcdClient.Put(ctx, "/topics/foo/0/sessions/session-0", "")
				Ω(err).ShouldNot(HaveOccurred())
			})
			Eventually(watchCh).Should(Receive())

			putResp, err := etcdClient.Put(sCtx, "/topics/foo/keys/abc", "0")
			Ω(err).ShouldNot(HaveOccurred())
			reg.Publish(sCtx, &subscribe.Message{Topic: "foo", Partition: 0, Key: "abc"})

			getResp, err := etcdClient.Get(sCtx, "/topics/foo/keys/abc")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(getResp.Count).Should(BeEquivalentTo(1))
			Ω(getResp.Kvs[0].ModRevision).Should(Equal(putResp.Header.Revision))

			cancel()
		}, SpecTimeout(10*time.Second))

		It("should create the partition key", func(sCtx SpecContext) {
			ctx, cancel := context.WithCancel(sCtx)

			var w = worker.Worker{
				HostId:    "worker-0",
				HostAddr:  "127.0.0.1:1234",
				Topic:     "foo",
				Partition: "0",
			}
			watchCh := etcdClient.Watch(ctx, "/topics/foo/0/routes/", clientv3.WithPrefix())
			var wg sync.WaitGroup
			defer wg.Wait()
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				Ω(reg.Register(ctx, w, 15)).Should(Succeed())
			}()
			By("adding sessions", func() {
				_, err := etcdClient.Put(ctx, "/topics/foo/0/sessions/session-0", "")
				Ω(err).ShouldNot(HaveOccurred())
			})
			Eventually(watchCh).Should(Receive())

			reg.Publish(sCtx, &subscribe.Message{Topic: "foo", Partition: 0, Key: "abc"})

			getResp, err := etcdClient.Get(sCtx, "/topics/foo/keys/abc")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(getResp.Count).Should(BeEquivalentTo(1))

			cancel()
		}, SpecTimeout(10*time.Second))

		It("should update the partition key", func(sCtx SpecContext) {
			ctx, cancel := context.WithCancel(sCtx)
			var w = worker.Worker{
				HostId:    "worker-0",
				HostAddr:  "127.0.0.1:1234",
				Topic:     "foo",
				Partition: "0",
			}
			watchCh := etcdClient.Watch(ctx, "/topics/foo/0/routes/", clientv3.WithPrefix())
			var wg sync.WaitGroup
			defer wg.Wait()
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				Ω(reg.Register(ctx, w, 15)).Should(Succeed())
			}()
			By("adding sessions", func() {
				_, err := etcdClient.Put(ctx, "/topics/foo/0/sessions/session-0", "")
				Ω(err).ShouldNot(HaveOccurred())
			})
			Eventually(watchCh).Should(Receive())

			putResp, err := etcdClient.Put(sCtx, "/topics/foo/keys/abc", "1")
			Ω(err).ShouldNot(HaveOccurred())
			reg.Publish(sCtx, &subscribe.Message{Topic: "foo", Partition: 0, Key: "abc"})

			getResp, err := etcdClient.Get(sCtx, "/topics/foo/keys/abc")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(getResp.Count).Should(BeEquivalentTo(1))
			Ω(getResp.Kvs[0].ModRevision).ShouldNot(Equal(putResp.Header.Revision))
			Ω(getResp.Kvs[0].Value).Should(BeEquivalentTo("0"))

			cancel()
		}, SpecTimeout(10*time.Second))
	})
})
