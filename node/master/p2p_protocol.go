package master

import (
	. "yu/common"
	. "yu/txn"
)

type HandShake struct {
	GenesisBlockHash Hash

	// When POW, these two will be always 0 and null.
	FinalizeHeight    BlockNum
	FinalizeBlockHash Hash

	EndHeight    BlockNum
	EndBlockHash Hash
}

type FetchBlocksRequest struct {
	StartHeight BlockNum
	Count       uint64
}

type PackedTxns struct {
	BlockHash string
	TxnsBytes []byte
}

func NewPackedTxns(blockHash Hash, txns SignedTxns) (*PackedTxns, error) {
	byt, err := txns.Encode()
	if err != nil {
		return nil, err
	}
	return &PackedTxns{
		BlockHash: blockHash.String(),
		TxnsBytes: byt,
	}, nil
}

func (pt *PackedTxns) Resolve() (Hash, SignedTxns, error) {
	txns := SignedTxns{}
	stxns, err := txns.Decode(pt.TxnsBytes)
	if err != nil {
		return NullHash, nil, err
	}
	return HexToHash(pt.BlockHash), stxns, nil
}
