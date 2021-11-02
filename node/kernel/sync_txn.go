package kernel

import (
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/types"
)

func (m *Kernel) handleSyncTxnsReq(byt []byte) ([]byte, error) {
	txnsReq, err := DecodeTxnsRequest(byt)
	if err != nil {
		return nil, err
	}
	var (
		txns             types.SignedTxns
		missingTxnHashes []Hash
	)
	for _, hash := range txnsReq.Hashes {
		stxn, err := m.txPool.GetTxn(hash)
		if err != nil {
			return nil, err
		}

		if stxn != nil {
			txns = append(txns, stxn)
		} else {
			missingTxnHashes = append(missingTxnHashes, hash)
		}
	}

	// request the node of block-producer for missingTxnHashes
	if txnsReq.BlockProducer != m.p2pNetwork.LocalID() {
		stxns, err := m.requestTxns(txnsReq.BlockProducer, txnsReq.BlockProducer, missingTxnHashes)
		if err != nil {
			return nil, err
		}

		txns = append(txns, stxns...)
	}

	var txnsByt []byte
	if txns != nil {
		txnsByt, err = txns.Encode()
		if err != nil {
			return nil, err
		}
	}

	return txnsByt, nil
}

func (m *Kernel) requestTxns(connectPeer, blockProducer peerstore.ID, txnHashes []Hash) (types.SignedTxns, error) {
	txnsRequest := TxnsRequest{
		Hashes:        txnHashes,
		BlockProducer: blockProducer,
	}
	reqByt, err := txnsRequest.Encode()
	if err != nil {
		return nil, err
	}

	respByt, err := m.p2pNetwork.RequestPeer(connectPeer, SyncTxnsCode, reqByt)
	if err != nil {
		return nil, err
	}
	return types.DecodeSignedTxns(respByt)
}
