package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	EthereumAPICounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ETH",
			Subsystem: "ethereum_api",
			Name:      "op_counter",
			Help:      "Total Operator number of counter",
		},
		[]string{TypeLbl, TypeStatusLbl},
	)

	EthereumAPICounterHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ETH",
		Subsystem: "ethereum_api",
		Name:      "op_execute_hist",
		Help:      "operation execute duration distribution.",
		Buckets:   prometheus.ExponentialBuckets(10, 2, 20), // 10us ~ 5s
	}, []string{TypeLbl, TypeStatusLbl})
)

func init() {
	prometheus.MustRegister(EthereumAPICounter)
	prometheus.MustRegister(EthereumAPICounterHist)
}
