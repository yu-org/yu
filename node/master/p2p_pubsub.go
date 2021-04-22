package master

import (
	"context"
	"encoding/json"
	. "yu/blockchain"
	. "yu/common"
	. "yu/txn"
)

const (
	BlockTopic        = "block"
	PackedTxnsTopic   = "packed-txns"
	UnpackedTxnsTopic = "unpacked-txns"
)

func (m *Master) pubBlock(block IBlock) error {
	byt, err := block.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(BlockTopic, byt)
}

func (m *Master) subBlock() (IBlock, error) {
	byt, err := m.subFromP2P(BlockTopic)
	if err != nil {
		return nil, err
	}
	return m.chain.NewEmptyBlock().Decode(byt)
}

func (m *Master) pubPackedTxns(blockHash Hash, txns SignedTxns) error {
	pt, err := NewPackedTxns(blockHash, txns)
	if err != nil {
		return err
	}
	byt, err := json.Marshal(pt)
	if err != nil {
		return err
	}
	return m.pubToP2P(PackedTxnsTopic, byt)
}

func (m *Master) subPackedTxns() (Hash, SignedTxns, error) {
	byt, err := m.subFromP2P(PackedTxnsTopic)
	if err != nil {
		return NullHash, nil, err
	}
	var pt PackedTxns
	err = json.Unmarshal(byt, &pt)
	if err != nil {
		return NullHash, nil, err
	}
	return pt.Resolve()
}

func (m *Master) pubUnpackedTxns(txns SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.pubToP2P(UnpackedTxnsTopic, byt)
}

func (m *Master) subUnpackedTxns() (SignedTxns, error) {
	byt, err := m.subFromP2P(UnpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxns(byt)
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
