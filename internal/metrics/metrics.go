package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "kgs"
)

var (
	ActiveSessions = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_sessions",
		Help:      "Tracks the total number of active sessions.",
	})
	KafkaMaxOffset = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "kafka_max_offset",
		Help:      "The high watermark offset consumed by the server.",
	})
	BufferedBroadcastEvictionTimestampSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "buffered_broadcast_eviction_timestamp_seconds",
		Help:      "Tracks the timestamp in seconds of evicted messages from the buffer.",
	})
	BufferedBroadcastDroppedMessagesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "buffered_broadcast_dropped_messages_total",
		Help:      "Counts the total number of dropped messages across all sessions",
	})
)

func Register() {
	prometheus.MustRegister(
		ActiveSessions,
		KafkaMaxOffset,
		BufferedBroadcastEvictionTimestampSeconds,
		BufferedBroadcastDroppedMessagesTotal,
	)
}
