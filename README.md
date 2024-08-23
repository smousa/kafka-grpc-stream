# kafka-grpc-stream

Scaleable streaming grpc service that allows external web clients to subscribe,
filter, and stream data from kafka topics. More details coming soon!

## Features

* 0.0.5 - Manual subscribe to a topic/partition server through grpc client
* 0.0.4 - Key registration (maps keys to partitions)
* 0.0.3 - Server-side host-session routing

## Local Development

### Pre-requisites

* go1.22.5
* docker >= 20.10.17
* make >= 4.3

### Running

Starts a local etcd, redpanda server and the kafka-grpc-stream server.

```bash
docker compose up
```

Access the local redpanda console at http://localhost:8000

Access the etcd via command line:

```bash
docker compose exec etcdctl help
```

### Testing

You can run the entire test suite using make:

```bash
make test
```

And the e2e test suite:

```bash
make e2e-test
```

...or you can have more fine-grained control of tests by running ginkgo

```bash
docker compose run --rm dev ginkgo help
```

Mocks are generated with mockery.  Update .mockery.yaml with the interface/s to
want to mock:

```bash
make mock
```
