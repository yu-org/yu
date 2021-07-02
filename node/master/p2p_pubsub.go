package master

import (
	"context"
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/txn"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	BlockTopic        = "block"
	UnpackedTxnsTopic = "unpacked-txns"
)

func (m *Master) initTopics() error {
	blockTopic, err := m.ps.Join(BlockTopic)
	if err != nil {
		return err
	}
	unpkgTxnsTopic, err := m.ps.Join(UnpackedTxnsTopic)
	if err != nil {
		return err
	}
	m.blockTopic = blockTopic
	m.unpackedTxnsTopic = unpkgTxnsTopic
	return nil
}

func (m *Master) pubBlock(block IBlock) error {
	byt, err := block.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(m.blockTopic, byt)
}

func (m *Master) subBlock() (IBlock, error) {
	byt, err := m.subFromP2P(m.blockTopic)
	if err != nil {
		return nil, err
	}
	return m.chain.NewEmptyBlock().Decode(byt)
}

//func (m *Master) pubPackedTxns(blockHash Hash, txns SignedTxns) error {
//	pt, err := NewPackedTxns(blockHash, txns)
//	if err != nil {
//		return err
//	}
//	byt, err := json.Marshal(pt)
//	if err != nil {
//		return err
//	}
//	return m.pubToP2P(m.packedTxnsTopic, byt)
//}
//
//func (m *Master) subPackedTxns() (Hash, SignedTxns, error) {
//	byt, err := m.subFromP2P(m.packedTxnsTopic)
//	if err != nil {
//		return NullHash, nil, err
//	}
//	var pt PackedTxns
//	err = json.Unmarshal(byt, &pt)
//	if err != nil {
//		return NullHash, nil, err
//	}
//
//	return pt.Resolve()
//}

func (m *Master) pubUnpackedTxns(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(m.unpackedTxnsTopic, byt)
}

func (m *Master) subUnpackedTxns() (SignedTxns, error) {
	byt, err := m.subFromP2P(m.unpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxns(byt)
}

func (m *Master) pubToP2P(topic *pubsub.Topic, msg []byte) error {
	return topic.Publish(context.Background(), msg)
}

func (m *Master) subFromP2P(topic *pubsub.Topic) ([]byte, error) {
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	msg, err := sub.Next(context.Background())
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}
