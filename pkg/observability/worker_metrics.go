package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	natsMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veil_nats_messages_total",
			Help: "NATS messages processed by Veil workers.",
		},
		[]string{"worker", "subject", "status"},
	)
	natsMessageDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veil_nats_message_duration_seconds",
			Help:    "NATS message processing duration.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"worker", "subject"},
	)
	neo4jOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veil_neo4j_operations_total",
			Help: "Neo4j operations executed by Veil services.",
		},
		[]string{"operation", "status"},
	)
	neo4jOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veil_neo4j_operation_duration_seconds",
			Help:    "Neo4j operation duration.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// RecordNatsMessage increments worker NATS metrics.
func RecordNatsMessage(worker, subject, status string, seconds float64) {
	natsMessagesTotal.WithLabelValues(worker, subject, status).Inc()
	natsMessageDuration.WithLabelValues(worker, subject).Observe(seconds)
}

// RecordNeo4jOperation increments Neo4j client metrics.
func RecordNeo4jOperation(operation, status string, seconds float64) {
	neo4jOperationsTotal.WithLabelValues(operation, status).Inc()
	neo4jOperationDuration.WithLabelValues(operation).Observe(seconds)
}
