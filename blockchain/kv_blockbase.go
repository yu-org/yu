package blockchain

import (
	. "yu/common"
	"yu/event"
	. "yu/storage/kv"
	"yu/txn"
)

type BlockBase struct {
	txnsBase   KV
	eventsBase KV
}

func NewBlockBase(txnsBase, eventsBase KV) IBlockBase {
	return &BlockBase{
		txnsBase:   txnsBase,
		eventsBase: eventsBase,
	}
}

func (bb *BlockBase) GetTxns(blockHash Hash) ([]txn.IsignedTxn, error) {
	hashByt := blockHash.Bytes()
	txnsByt, err := bb.txnsBase.Get(hashByt)
	if err != nil {
		return nil, err
	}
	txns, err := txn.DecodeSignedTxns(txnsByt)
	if err != nil {
		return nil, err
	}
	return txns.ToArray(), nil
}

func (bb *BlockBase) SetTxns(blockHash Hash, txns txn.SignedTxns) error {
	hashByt := blockHash.Bytes()
	txnsByt, err := txns.Encode()
	if err != nil {
		return err
	}
	return bb.txnsBase.Set(hashByt, txnsByt)
}

func (bb *BlockBase) GetEvents(blockHash Hash) ([]event.IEvent, error) {
	hashByt := blockHash.Bytes()
	eventsByt, err := bb.eventsBase.Get(hashByt)
	if err != nil {
		return nil, err
	}
	events, err := event.DecodeEvents(eventsByt)
	if err != nil {
		return nil, err
	}
	return events.ToArray(), nil
}

func (bb *BlockBase) SetEvents(blockHash Hash, events event.Events) error {
	hashByt := blockHash.Bytes()
	eventsByt, err := events.Encode()
	if err != nil {
		return err
	}
	return bb.eventsBase.Set(hashByt, eventsByt)
}
