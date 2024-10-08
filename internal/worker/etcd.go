package worker

import (
	"context"
	"encoding/json"
	"path"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smousa/kafka-grpc-stream/internal/subscribe"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistry struct {
	lease   clientv3.Lease
	kv      clientv3.KV
	watcher clientv3.Watcher

	mu       sync.Mutex
	isLeader bool

	workerTTL int64
	keyTTL    int64
}

type EtcdRegistryOp func(*EtcdRegistry)

func WithEtcdClient(cli *clientv3.Client) EtcdRegistryOp {
	return func(r *EtcdRegistry) {
		r.lease = clientv3.NewLease(cli)
		r.kv = clientv3.NewKV(cli)
		r.watcher = clientv3.NewWatcher(cli)
	}
}

func WithWorkerTTL(ttl int64) EtcdRegistryOp {
	return func(r *EtcdRegistry) {
		r.workerTTL = ttl
	}
}

func WithKeyTTL(ttl int64) EtcdRegistryOp {
	return func(r *EtcdRegistry) {
		r.keyTTL = ttl
	}
}

func NewEtcdRegistry(ops ...EtcdRegistryOp) *EtcdRegistry {
	r := &EtcdRegistry{}
	for _, op := range ops {
		op(r)
	}

	return r
}

func (r *EtcdRegistry) Register(ctx context.Context, worker Worker) error {
	log := zerolog.Ctx(ctx)

	// Get the key value data
	var (
		base        = path.Join("/t", worker.Topic, "p", worker.Partition)
		hostPath    = path.Join(base, "host") + "/"
		sessionPath = path.Join(base, "sess") + "/"
		routePath   = path.Join(base, "rout") + "/"
	)

	hostBytes, err := json.Marshal(worker)
	if err != nil {
		return errors.Wrap(err, "could not serialize worker info")
	}

	// Grant the lease
	grant, err := r.lease.Grant(ctx, r.workerTTL)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return errors.Wrap(err, "could not grant lease")
	}

	// Revoke leader privileges
	defer func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.isLeader = false
	}()

	//nolint:errcheck
	defer r.lease.Revoke(context.WithoutCancel(ctx), grant.ID)

	// Register the host
	hostResp, err := r.kv.Put(ctx, path.Join(hostPath, worker.HostId), string(hostBytes),
		clientv3.WithLease(grant.ID),
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return errors.Wrap(err, "could not register host")
	}

	// Track the revision
	rev := hostResp.Header.Revision

	// initialize state
	err = r.resetState(ctx, worker.HostId, hostPath, sessionPath, routePath, rev)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return err
	}

	// Monitor the lease
	keepAlive, err := r.lease.KeepAlive(ctx, grant.ID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return errors.Wrap(err, "could not monitor lease")
	}

	// Monitor the hosts
	hostCh := r.watcher.Watch(ctx, hostPath,
		clientv3.WithPrefix(),
		clientv3.WithRev(rev),
	)

	// Monitor the sessions
	sessionCh := r.watcher.Watch(ctx, sessionPath,
		clientv3.WithPrefix(),
		clientv3.WithRev(rev),
	)

	for {
		select {
		case _, ok := <-keepAlive:
			if !ok {
				return nil
			}
		case resp, ok := <-hostCh:
			if !ok {
				return nil
			}

			// Handle connection closed by the server
			if resp.Canceled {
				if err := resp.Err(); errors.Is(err, rpctypes.ErrCompacted) {
					if rev < resp.CompactRevision {
						rev = resp.CompactRevision

						err = r.resetState(ctx, worker.HostId, hostPath, sessionPath, routePath, rev)
						if err != nil {
							if errors.Is(err, context.Canceled) {
								return nil
							}

							return err
						}
					}
				} else {
					log.Err(err).Str("watch_path", hostPath).Msg("Watcher exited")
				}

				// Recreate the watcher
				hostCh = r.watcher.Watch(ctx, hostPath,
					clientv3.WithPrefix(),
					clientv3.WithRev(rev),
				)

				break
			}

			if rev >= resp.Header.Revision {
				break
			}

			// Update leader and rebalance resources
			rev = resp.Header.Revision

			err = r.resetState(ctx, worker.HostId, hostPath, sessionPath, routePath, rev)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				} else if errors.Is(err, rpctypes.ErrCompacted) {
					break
				}

				return err
			}
		case resp, ok := <-sessionCh:
			if !ok {
				return nil
			}

			// Handle connection closed by the server
			if resp.Canceled {
				if err := resp.Err(); errors.Is(err, rpctypes.ErrCompacted) {
					if rev < resp.CompactRevision {
						rev = resp.CompactRevision

						err = r.resetState(ctx, worker.HostId, hostPath, sessionPath, routePath, rev)
						if err != nil {
							if errors.Is(err, context.Canceled) {
								return nil
							}

							return err
						}
					}
				} else {
					log.Err(err).Str("watch_path", sessionPath).Msg("Watcher exited")
				}

				// Recreate the watcher
				sessionCh = r.watcher.Watch(ctx, sessionPath,
					clientv3.WithPrefix(),
					clientv3.WithRev(rev),
				)

				break
			}

			if rev >= resp.Header.Revision {
				break
			}

			// Update leader and rebalance resources
			rev = resp.Header.Revision

			err = r.resetState(ctx, worker.HostId, hostPath, sessionPath, routePath, rev)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				} else if errors.Is(err, rpctypes.ErrCompacted) {
					break
				}

				return err
			}
		}
	}
}

func (r *EtcdRegistry) resetState(ctx context.Context, hostId, hostPath, sessionPath, routePath string, rev int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	leaderHostId, err := r.getLeader(ctx, hostPath, rev)
	if err != nil {
		return errors.Wrap(err, "could not reset leader state")
	}

	if r.isLeader = leaderHostId == hostId; !r.isLeader {
		return nil
	}

	err = r.rebalance(ctx, hostPath, sessionPath, routePath, rev)

	return errors.Wrap(err, "could not reset routes")
}

func (r *EtcdRegistry) getLeader(ctx context.Context, hostPath string, rev int64) (string, error) {
	resp, err := r.kv.Get(ctx, hostPath, append(
		clientv3.WithFirstCreate(),
		clientv3.WithPrefix(),
		clientv3.WithKeysOnly(),
		clientv3.WithRev(rev),
	)...)
	if err != nil {
		return "", errors.Wrap(err, "could not look up leader")
	}

	return path.Base(string(resp.Kvs[0].Key)), nil
}

func (r *EtcdRegistry) rebalance(ctx context.Context, hostPath, sessionPath, routePath string, rev int64) error {
	// Get the list of hosts
	hostResp, err := r.kv.Get(ctx, hostPath,
		clientv3.WithPrefix(),
		clientv3.WithRev(rev),
		clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
	)
	if err != nil {
		return errors.Wrap(err, "could not look up hosts")
	}

	// Get the list of sessions
	sessionResp, err := r.kv.Get(ctx, sessionPath,
		clientv3.WithPrefix(),
		clientv3.WithRev(rev),
		clientv3.WithKeysOnly(),
		clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
	)
	if err != nil {
		return errors.Wrap(err, "could not look up sessions")
	}

	// Rebalance sessions
	var (
		routeMap       = make(map[string]string, sessionResp.Count)
		hostIndex      = -1
		sessionEnd     = 0
		hostSessionDiv = int(sessionResp.Count / hostResp.Count)
		hostSessionMod = int(sessionResp.Count % hostResp.Count)
	)

	for i, session := range sessionResp.Kvs {
		if i >= sessionEnd {
			sessionEnd += hostSessionDiv
			if hostIndex += 1; hostIndex < hostSessionMod {
				sessionEnd += 1
			}
		}

		key := path.Join(routePath, path.Base(string(session.Key)))
		value := string(hostResp.Kvs[hostIndex].Value)
		routeMap[key] = value
	}

	// Get the list of existing routes
	routeResp, err := r.kv.Get(ctx, routePath,
		clientv3.WithPrefix(),
		clientv3.WithRev(rev),
	)
	if err != nil {
		return errors.Wrap(err, "could not look up routes")
	}

	// Setup and execute transaction
	ops := make([]clientv3.Op, 0, sessionResp.Count+routeResp.Count)

	// Update or delete existing route assignments
	for _, route := range routeResp.Kvs {
		key, value := string(route.Key), string(route.Value)
		if v, ok := routeMap[key]; ok {
			if v != value {
				ops = append(ops, clientv3.OpPut(key, v))
			}

			delete(routeMap, key)
		} else {
			ops = append(ops, clientv3.OpDelete(key))
		}
	}

	// Add new route assignments
	for key, value := range routeMap {
		ops = append(ops, clientv3.OpPut(key, value))
	}

	// If no changes, don't update route assignments
	if len(ops) == 0 {
		return nil
	}

	_, err = r.kv.Txn(ctx).
		If(
			clientv3.Compare(clientv3.ModRevision(hostPath).WithPrefix(), "<", rev+1),
			clientv3.Compare(clientv3.ModRevision(sessionPath).WithPrefix(), "<", rev+1),
		).
		Then(ops...).
		Commit()

	return errors.Wrap(err, "could not update routes")
}

func (r *EtcdRegistry) Publish(ctx context.Context, message *subscribe.Message) {
	log := zerolog.Ctx(ctx).With().Str("key", message.Key).Logger()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Skip registration if not the leader
	if !r.isLeader {
		return
	}

	// Load the key
	key := path.Join("/t", message.Topic, "k", message.Key)
	value := strconv.Itoa(int(message.Partition))

	resp, err := r.kv.Get(ctx, key)
	if err != nil {
		log.Err(err).Msg("Could not search key")

		return
	}

	var ops []clientv3.OpOption

	if resp.Count == 0 {
		grant, err := r.lease.Grant(ctx, r.keyTTL)
		if err != nil {
			log.Err(err).Msg("Could not grant lease")

			return
		}

		ops = append(ops, clientv3.WithLease(grant.ID))
	} else {
		keyRoute := resp.Kvs[0]

		if _, err := r.lease.KeepAliveOnce(ctx, clientv3.LeaseID(keyRoute.Lease)); err != nil {
			if !errors.Is(err, rpctypes.ErrLeaseNotFound) {
				log.Err(err).Msg("Could not renew lease")

				return
			}

			grant, err := r.lease.Grant(ctx, r.keyTTL)
			if err != nil {
				log.Err(err).Msg("Could not grant lease")

				return
			}

			ops = append(ops, clientv3.WithLease(grant.ID))
		}

		if len(ops) == 0 && string(keyRoute.Value) == value {
			// nothing to update
			return
		}
	}

	if _, err := r.kv.Put(ctx, key, value, ops...); err != nil {
		log.Err(err).Msg("Could not register key")
	}
}
