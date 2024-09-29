package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	TripodLabel = "tripod"
)

var (
	KernelHandleTxnCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "yu",
		Subsystem: "kernel",
		Name:      "handle_txn_count",
		Help:      "Counter of Kernel Handle Txn",
	}, []string{})

	TxPoolInsertCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "yu",
		Subsystem: "txpool",
		Name:      "insert_count",
		Help:      "Counter of txpool insert",
	}, []string{})

	TxpoolSizeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "yu",
			Subsystem: "txpool",
			Name:      "size_gauge",
			Help:      "Gauge of number of txpool",
		},
	)

	StartBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "start_block",
			Help:      "Start Block duration",
		},
		[]string{TripodLabel},
	)

	EndBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "end_block",
			Help:      "End Block duration",
		},
		[]string{TripodLabel},
	)

	FinalizeBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "finalize_block",
			Help:      "Finalize Block duration",
		},
		[]string{TripodLabel},
	)

	AppendBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "append_block",
			Help:      "append block duration",
		},
		[]string{},
	)

	StateCommitDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "state",
			Name:      "state_commit",
			Help:      "State Commit duration",
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(KernelHandleTxnCounter)
	prometheus.MustRegister(TxPoolInsertCounter)
	prometheus.MustRegister(TxpoolSizeGauge)
	// prometheus.MustRegister(AppendBlockDuration, StartBlockDuration, EndBlockDuration, FinalizeBlockDuration)
	prometheus.MustRegister(StateCommitDuration)
}
