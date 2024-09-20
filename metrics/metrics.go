package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	TripodLabel = "tripod"
)

var (
	TxpoolSizeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "yu",
			Subsystem: "txpool",
			Name:      "txpool_size",
			Help:      "Total number of txpool",
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
	prometheus.MustRegister(TxpoolSizeGauge)
	// prometheus.MustRegister(AppendBlockDuration, StartBlockDuration, EndBlockDuration, FinalizeBlockDuration)
	prometheus.MustRegister(StateCommitDuration)
}
