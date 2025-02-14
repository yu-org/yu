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

	TxnDBDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "yu",
		Subsystem: "txndb",
		Name:      "duration",
		Help:      "Hist Duration of txnDB",
		Buckets:   prometheus.ExponentialBuckets(10, 2, 20), // 10us ~ 5s
	}, []string{TypeLbl, SourceTypeLbl, OpLabel, StatusLbl})
)

func initTxnDBMetrics() {
	prometheus.MustRegister(TxnDBCounter)
	prometheus.MustRegister(TxnDBDurationHistogram)
}
