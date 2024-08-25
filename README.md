# kafka-grpc-stream

Scaleable streaming grpc service that allows external web clients to subscribe,
filter, and stream data from kafka topics. More details coming soon!

## Features

* 1.0.0 - Key glob filtering
* 0.0.5 - Manual subscribe to a topic/partition server through grpc client
* 0.0.4 - Key registration (maps keys to partitions)
* 0.0.3 - Server-side host-session routing

## Usage

Use the grpc go client to connect to the server.

```go
package main

import (
    "encoding/json"
    "fmt"
	"os/signal"
	"syscall"

    "github.com/smousa/kafka-grpc-stream/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const ClientUrl = "localhost:9000"

func main() {
    // set up the client
    cli, err := grpc.NewClient(ClientUrl,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        fmt.Println("Error connecting to client:", err)
        return
    }
    defer cli.Close()

    ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // create the subsciption
    stream, err := cli.Subscribe(ctx)
    if err != nil {
        fmt.Println("Error creating subscription:", err)
        return
    }

    // provide the filter
    err = stream.Send(&protobuf.SubscribeRequest{
        Keys: []string{"*"}, // send everything
    })
    if err != nil {
        fmt.Println("Error sending subscription filter:", err)
        return
    }

    // watch the data stream in
    for {
        msg, err := stream.Recv()
        if err != nil {
            fmt.Println("Closing stream:", err)
            return
        }

        msgBytes, err := json.Marshal(msg)
        if err != nil {
            fmt.Println("Could not marshal message: ", err)
            return
        }

        fmt.Println(string(msgBytes))
    }
}

```

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
