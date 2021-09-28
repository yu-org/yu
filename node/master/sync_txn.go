package master

import (
	"context"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/txn"
)

func (m *Master) handleSyncTxnsReq(byt []byte, s network.Stream) error {
	txnsReq, err := DecodeTxnsRequest(byt)
	if err != nil {
		return err
	}
	var (
		txns             SignedTxns
		missingTxnHashes []Hash
	)
	for _, hash := range txnsReq.Hashes {
		stxn, err := m.txPool.GetTxn(hash)
		if err != nil {
			return err
		}

		if stxn != nil {
			txns = append(txns, stxn)
		} else {
			missingTxnHashes = append(missingTxnHashes, hash)
		}
	}

	// request the node of block-producer for missingTxnHashes
	if txnsReq.BlockProducer != m.host.ID() {
		stxns, err := m.requestTxns(txnsReq.BlockProducer, txnsReq.BlockProducer, missingTxnHashes)
		if err != nil {
			return err
		}

		txns = append(txns, stxns...)
	}

	var txnsByt []byte
	if txns != nil {
		txnsByt, err = txns.Encode()
		if err != nil {
			return err
		}
	}

	return writeToStream(txnsByt, s)
}

func (m *Master) requestTxns(connectPeer, blockProducer peerstore.ID, txnHashes []Hash) (SignedTxns, error) {
	s, err := m.host.NewStream(context.Background(), connectPeer, m.protocolID)
	if err != nil {
		return nil, err
	}
	txnsRequest := TxnsRequest{
		Hashes:        txnHashes,
		BlockProducer: blockProducer,
	}
	reqByt, err := txnsRequest.Encode()
	if err != nil {
		return nil, err
	}
	err = writeToStream(reqByt, s)
	if err != nil {
		return nil, err
	}
	respByt, err := readFromStream(s)
	if err != nil {
		return nil, err
	}
	return DecodeSignedTxns(respByt)
}
