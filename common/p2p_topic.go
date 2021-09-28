package common

import pubsub "github.com/libp2p/go-libp2p-pubsub"

const (
	StartBlockTopic    = "start-block"
	EndBlockTopic      = "end-block"
	FinalizeBlockTopic = "finalize-block"
	UnpackedTxnsTopic  = "unpacked-txns"
)

var (
	TopicsMap = make(map[string]*pubsub.Topic, 0)
	SubsMap   = make(map[string]*pubsub.Subscription, 0)
)
