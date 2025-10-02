package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	TypeLbl       = "type"
	OpLabel       = "op"
	StatusLbl     = "status"
	SourceTypeLbl = "source"
)

var (
	TxnDBCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "yu",
		Subsystem: "txndb",
		Name:      "op_counter",
		Help:      "Counter of txnDB",
	}, []string{TypeLbl, SourceTypeLbl, OpLabel, StatusLbl})

	TxnDBDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "eth",
			Subsystem: "txndb",
			Name:      "op_duration",
			Help:      "txn execute duration distribution.",
			Buckets:   prometheus.ExponentialBuckets(10, 2, 20), // 10us ~ 5s
		},
		[]string{TypeLbl, OpLabel},
	)
)

func initTxnDBMetrics() {
	prometheus.MustRegister(TxnDBCounter)
	prometheus.MustRegister(TxnDBDuration)
}
