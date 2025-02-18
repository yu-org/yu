package txpool

import "github.com/prometheus/client_golang/prometheus"

const (
	TypeLbl = "type"
)

var (
	TxnPoolCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "yu",
		Subsystem: "txn_pool",
		Name:      "op_counter",
		Help:      "Counter of txnPool",
	}, []string{TypeLbl})

	TxnPoolDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "reddio",
			Subsystem: "txn_pool",
			Name:      "op_duration",
			Help:      "txn execute duration distribution.",
			Buckets:   prometheus.ExponentialBuckets(10, 2, 20), // 10us ~ 5s
		},
		[]string{TypeLbl},
	)
)

func init() {
	prometheus.MustRegister(TxnPoolCounter)
	prometheus.MustRegister(TxnPoolDuration)
}
