package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	TypeLbl   = "type"
	OpLabel   = "op"
	StatusLbl = "status"
)

var (
	TxnDBCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "yu",
		Subsystem: "txndb",
		Name:      "op_counter",
		Help:      "Counter of txnDB",
	}, []string{TypeLbl, OpLabel, StatusLbl})
)

func initTxnDBMetrics() {
	prometheus.MustRegister(TxnDBCounter)
}
