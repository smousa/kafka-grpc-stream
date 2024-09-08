# kafka-grpc-stream

Scaleable streaming grpc service that allows external web clients to subscribe,
filter, and stream data from kafka topics. More details coming soon!

## Features

* 1.4.0 - Added TTL to key registry keys
* 1.3.0 - Max age and min offset filters
* 1.2.1 - More stable load balancer
* 1.2.0 - Monitoring and CLI tool
* 1.1.0 - Buffered broadcasting
* 1.0.0 - Key glob filtering
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

### CLI

Connect to the running server using the built-in CLI tool.  Default returns all
results.

```bash
docker compose exec -it server cli [--keys KEY1 --keys KEY2 ...]
```

### Redpanda

Access the local redpanda console at http://localhost:8000

### Etcd

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

### Generators

#### Mocks

Mocks are generated with mockery.  Update .mockery.yaml with the interface/s to
want to mock:

```bash
make mock
```

#### Protobuf

Protobufs are manages in the protobuf directory.  To generate protobufs:

```bash
make proto
```

## Examples

* **[Tuning Demo](./examples/tuning):** A grafana dashboard for tuning a single
  kafka-grpc-stream server instance.

## Build Reference

* [Redpanda](https://github.com/redpanda-data/redpanda)
* [Redpanda Console](https://github.com/redpanda-data/console)
* [Redpanda Connect](https://github.com/redpanda-data/connect)
* [Etcd](https://github.com/etcd-io/etcd)
* [CAdvisor](https://github.com/google/cadvisor)
* [Prometheus](https://github.com/prometheus/prometheus)
* [Grafana](https://github.com/grafana/grafana)
* [Alpine](https://hub.docker.com/_/alpine)
* [Golang](https://github.com/golang/go)
* [golangci-lint](https://github.com/golangci/golangci-lint)
* [protoc](https://github.com/protocolbuffers/protobuf)
* [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf)
* [protoc-gen-go-grpc](https://pkg.go.dev/google.golang.org/grpc/cmd/protoc-gen-go-grpc)
* [mockery](https://github.com/vektra/mockery)
* [go-test-coverage](https://github.com/vladopajic/go-test-coverage)
