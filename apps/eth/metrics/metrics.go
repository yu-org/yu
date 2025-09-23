package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	TypeLbl       = "type"
	TypeCountLbl  = "count"
	TypeStatusLbl = "status"
)

var (
	TxnCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "txn",
			Name:      "total_count",
			Help:      "Total number of count for txn",
		},
		[]string{TypeLbl},
	)

	BatchTxnCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "batch_txn",
			Name:      "total_count",
			Help:      "Total number of redo count for batch txn",
		},
		[]string{TypeLbl},
	)

	BlockTxnCommitDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eth",
			Subsystem: "block_txn",
			Name:      "commit_duration_seconds",
			Help:      "txn commit duration seconds",
		},
		[]string{},
	)

	BatchTxnDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "eth",
			Subsystem: "batch_txn",
			Name:      "execute_duration_seconds",
			Help:      "txn execute duration distribution.",
			Buckets:   prometheus.ExponentialBuckets(10, 2, 20), // 10us ~ 5s
		},
		[]string{TypeLbl},
	)

	BatchTxnSplitCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "batch_txn",
			Name:      "split_txn_count",
			Help:      "split sub batch txn count",
		},
		[]string{TypeCountLbl},
	)

	BlockExecuteTxnDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eth",
			Subsystem: "block",
			Name:      "execute_duration_seconds",
			Help:      "block execute txn duration",
		},
		[]string{},
	)

	BlockExecuteTxnCountGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eth",
			Subsystem: "block",
			Name:      "execute_txn_count",
			Help:      "txn count for each block",
		}, []string{})

	BlockTxnPrepareDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eth",
			Subsystem: "block_txn",
			Name:      "prepare_txn_duration_seconds",
			Help:      "split batch txn duration",
		},
		[]string{},
	)

	BlockTxnAllExecuteDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eth",
			Subsystem: "block_txn",
			Name:      "execute_all_duration_seconds",
			Help:      "split batch txn duration",
		},
		[]string{},
	)
	DownwardMessageSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "downward_success_total",
			Help:      "Total number of successfully processed downward messages",
		},
		[]string{TypeLbl},
	)

	DownwardMessageFailureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "downward_failure_total",
			Help:      "Total number of failed downward message processing attempts",
		},
		[]string{TypeLbl},
	)

	DownwardMessageReceivedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "downward_received_total",
			Help:      "Total number of received downward messages",
		},
		[]string{TypeLbl},
	)

	UpwardMessageSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "upward_success_total",
			Help:      "Total number of successfully processed upward messages",
		},
		[]string{TypeLbl},
	)

	UpwardMessageFailureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "upward_failure_total",
			Help:      "Total number of failed upward message processing attempts",
		},
		[]string{TypeLbl},
	)

	UpwardMessageReceivedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "upward_received_total",
			Help:      "Total number of received upward messages",
		},
		[]string{TypeLbl},
	)
	L1EventWatcherFailureCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "l1_event_watcher",
			Name:      "failure_total",
			Help:      "Total number of L1 event watcher failures",
		},
	)

	L1EventWatcherRetryCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "l1_event_watcher",
			Name:      "retry_total",
			Help:      "Total number of L1 event watcher retries",
		},
	)

	WithdrawMessageNonceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "withdraw_message_nonce",
			Help:      "The current value of the bridge event nonce.",
		},
		[]string{"table"},
	)

	WithdrawMessageNonceGap = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eth",
			Subsystem: "bridge",
			Name:      "withdraw_message_nonce_gap",
			Help:      "The gap between the current value of the bridge event nonce and the expected value.",
		},
		[]string{TypeLbl, TypeStatusLbl},
	)
)

func init() {
	prometheus.MustRegister(TxnCounter)

	prometheus.MustRegister(BlockExecuteTxnCountGauge)
	prometheus.MustRegister(BlockTxnPrepareDurationGauge)
	prometheus.MustRegister(BlockTxnAllExecuteDurationGauge)
	prometheus.MustRegister(BlockTxnCommitDurationGauge)
	prometheus.MustRegister(BlockExecuteTxnDurationGauge)

	prometheus.MustRegister(BatchTxnCounter)
	prometheus.MustRegister(BatchTxnSplitCounter)
	prometheus.MustRegister(BatchTxnDuration)

	prometheus.MustRegister(DownwardMessageSuccessCounter)
	prometheus.MustRegister(DownwardMessageFailureCounter)
	prometheus.MustRegister(DownwardMessageReceivedCounter)
	prometheus.MustRegister(UpwardMessageSuccessCounter)
	prometheus.MustRegister(UpwardMessageFailureCounter)
	prometheus.MustRegister(UpwardMessageReceivedCounter)

	prometheus.MustRegister(L1EventWatcherFailureCounter)
	prometheus.MustRegister(L1EventWatcherRetryCounter)
	prometheus.MustRegister(WithdrawMessageNonceGauge)
	prometheus.MustRegister(WithdrawMessageNonceGap)

}

var TxnBuckets = []float64{.00005, .0001, .00025, .0005, .001, .0025, .005, 0.01, 0.025, 0.05, 0.1}
