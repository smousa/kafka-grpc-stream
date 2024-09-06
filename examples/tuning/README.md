# Tuning Demo

This example demonstrates how to use monitoring tools for configure the
kafka-grpc-stream server to support your use case.

## Local Setup

This example uses docker compose to set up the server and dependencies, along
with a data collector and monitoring tools.

To run, enter the following command:

```bash
docker compose --env-file ../../.env up -d
```

### Application and dependencies

* *****redpanda** - Server dependency where data is retrieved
* **etcd** - Used for routing and load balancing (coming soon!)
* **server** - kafka-grpc-stream service that we are instrumenting for sizing
  purposes

### Data generation and consumption

* **create-topic** - command used to provision topics in the local redpanda
  instance
* **collector** - This is a [redpanda connect](https://docs.redpanda.com/redpanda-connect/about/)
  pipeline that sources trade data from the [bitmex](https://www.bitmex.com/)
  api and writes the results to our kafka instance.  When validating with your
  own use case, it is expected that you would either provide your own collector
  or connect to kafka brokers that best simulate your expected ingestion
  patterns.
* **client** - kafka-grpc-stream command line tool to help simulate end-user
  traffic.  This tool writes the responses to stdout, so your mileage may vary
  based on expected usage patterns.

### Monitoring Tools

* **redpanda-console** - (localhost:8000) UI for the local kafka instance
* **[cadvisor](https://github.com/google/cadvisor)** - provides docker container resource metrics
* **[prom](https://prometheus.io/)** - prometheus server for scraping metrics
* **[grafana](https://grafana.com/)** - (localhost:3000) UI for metrics visualization

## Dashboard

Once your infrastructure is running, you'll want to check out the [tuning dashboard]()
to visualize the operational processes of your application. It highlights the
following metrics:

* CPU Usage
* Memory Usage
* Dropped Messages Per Session Rate
* Total Active Sessions
* Evicted Message Age
* Consumer Lag
* Kafka Topic Messages Produced Rate
* Kafka Topic Bytes Produced Rate

### Interactions and Recommendations

Each of these metrics have some relationship with the following configuration:

#### CPU Allocation**

This value is influenced largely based on the kafka topic ingest rate.

#### Memory Allocation

This value is influenced by the kafka topic ingest rate as well as the buffer
size.  This value also tends to increase when new sessions are added, but
eventually stablizes as the session reads the outstanding messages
in the buffer.

#### KAFKA_FETCH_MAX_BYTES

Defines the maximum amount of data the kafka server should return for a fetch
request (https://docs.confluent.io/platform/current/installation/configuration/consumer-configs.html#fetch-max-bytes)

This setting helps regulate memory usage at the cost of message latency. You
will want to be observant of the Consumer Lag and Kafka Topics Bytes Produced
Rate to ensure that you are able to consume messages at the rate messages are
produced.

#### SERVER_BROADCAST_BUFFER_SIZE

The kafka-grpc-stream server broadcasts messages by reading and writing messages
to a circular buffer.  If the buffer size is set too low, then client sessions
run the risk of missing messages as they get evicted faster than they can be
consumed.

To measure the performance of the buffer, you'll want to pay attention to
Dropped Messages Per Session Rate, Total Active Sessions, Evicted Message Age,
and Topic Messages Produced Rate.
