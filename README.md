# kafka-grpc-stream

Scaleable streaming grpc service that allows external web clients to subscribe,
filter, and stream data from kafka topics. More details coming soon!

## Features

* Server-side host-session routing

## Local Development

### Pre-requisites

* go1.22.5
* docker >= 20.10.17
* make >= 4.3

### Running

Starts a local etcd server and the kafka-grpc-stream server.

```bash
docker compose up
```

### Testing

You can run the entire test suite using make:

```bash
make test
```

...or you can have more fine-grained control of tests by running ginkgo

```bash
docker compose run --rm dev ginkgo help
```
