package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	BlockNumLabel = "block_num"
	TripodLabel   = "tripod"
)

var (
	TxsPackCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yu",
			Subsystem: "transaction",
			Name:      "txs_pack",
			Help:      "Total number of count of packing txn",
		},
		[]string{BlockNumLabel},
	)

	StartBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "start_block",
			Help:      "Start Block duration",
		},
		[]string{BlockNumLabel, TripodLabel},
	)

	EndBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "end_block",
			Help:      "End Block duration",
		},
		[]string{BlockNumLabel, TripodLabel},
	)

	FinalizeBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "finalize_block",
			Help:      "Finalize Block duration",
		},
		[]string{BlockNumLabel, TripodLabel},
	)

	AppendBlockDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "block",
			Name:      "append_block",
			Help:      "append block duration",
		},
		[]string{BlockNumLabel},
	)

	StateCommitDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yu",
			Subsystem: "state",
			Name:      "state_commit",
			Help:      "State Commit duration",
		},
		[]string{BlockNumLabel},
	)
)

func init() {
	prometheus.MustRegister(TxsPackCounter)
	prometheus.MustRegister(AppendBlockDuration, StartBlockDuration, EndBlockDuration, FinalizeBlockDuration)
	prometheus.MustRegister(StateCommitDuration)
}
