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
)

func initTxnDBMetrics() {
	prometheus.MustRegister(TxnDBCounter)
}
