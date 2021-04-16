package master

import (
	"context"
	. "yu/blockchain"
	. "yu/txn"
)

const (
	BlockTopic = "block"
	TxnsTopic  = "txns"
)

func (m *Master) pubBlockToP2P(block IBlock) error {
	byt, err := block.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(BlockTopic, byt)
}

func (m *Master) subBlockFromP2P() (IBlock, error) {
	byt, err := m.subFromP2P(BlockTopic)
	if err != nil {
		return nil, err
	}
	return m.chain.NewEmptyBlock().Decode(byt)
}

func (m *Master) pubTxnsToP2P(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(TxnsTopic, byt)
}

func (m *Master) subTxnsFromP2P() (SignedTxns, error) {
	byt, err := m.subFromP2P(TxnsTopic)
	if err != nil {
		return nil, err
	}
	return m.txPool.NewEmptySignedTxns().Decode(byt)
}

func (m *Master) pubToP2P(topic string, msg []byte) error {
	tp, err := m.ps.Join(topic)
	if err != nil {
		return err
	}
	return tp.Publish(context.Background(), msg)
}

func (m *Master) subFromP2P(topic string) ([]byte, error) {
	tp, err := m.ps.Join(topic)
	if err != nil {
		return nil, err
	}
	sub, err := tp.Subscribe()
	if err != nil {
		return nil, err
	}
	msg, err := sub.Next(context.Background())
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}
