package kernel

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/types"
)

// Shake hand to the node of p2p network when starts up.
// If we have missing history block, fetch them.
func (m *Kernel) SyncHistory() error {
	logrus.Info("start to sync history from other node")

	resp, err := m.requestBlocks(nil)
	if err != nil {
		return err
	}
	if resp.Err != nil {
		return resp.Err
	}

	for resp.MissingRange != nil {
		// todo: the missing range maybe very huge and we need fetch them multiple times
		// the remote node will return new Missing blocks-range in this response.
		resp, err = m.requestBlocks(resp.MissingRange)
		if err != nil {
			return err
		}

		if resp.Err != nil {
			return resp.Err
		}

		if resp.BlocksByt != nil {
			blocks, err := m.chain.DecodeBlocks(resp.BlocksByt)
			if err != nil {
				return err
			}

			err = m.SyncHistoryBlocks(blocks)
			if err != nil {
				return err
			}

			resp.MissingRange = nil
		}

		if resp.TxnsByt != nil {
			for blockHash, byt := range resp.TxnsByt {
				txns, err := types.DecodeSignedTxns(byt)
				if err != nil {
					return err
				}
				err = m.base.SetTxns(blockHash, txns)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func (m *Kernel) handleHsReq(byt []byte) ([]byte, error) {
	remoteReq, err := DecodeHsRequest(byt)
	if err != nil {
		return nil, err
	}

	var (
		blocksByt []byte
		txnsByt   map[Hash][]byte
	)
	if remoteReq.FetchRange != nil {
		blocksByt, txnsByt, err = m.getMissingBlocksTxns(remoteReq)
		if err != nil {
			return nil, err
		}
	}

	missingRange, err := m.compareMissingRange(remoteReq.Info)

	if missingRange != nil {
		logrus.Debugf("missing range start-height is %d,  end-height is %d", missingRange.StartHeight, missingRange.EndHeight)
	}

	hsResp := &HandShakeResp{
		MissingRange: missingRange,
		BlocksByt:    blocksByt,
		TxnsByt:      txnsByt,
		Err:          err,
	}
	return hsResp.Encode()
}

func (m *Kernel) requestBlocks(fetchRange *BlocksRange) (*HandShakeResp, error) {
	hs, err := m.NewHsReq(fetchRange)
	if err != nil {
		return nil, err
	}

	if hs.FetchRange != nil {
		logrus.Infof("fetch history blocks from (%d) to (%d)", hs.FetchRange.StartHeight, hs.FetchRange.EndHeight)
	}

	byt, err := hs.Encode()
	if err != nil {
		return nil, err
	}

	respByt, err := m.p2pNetwork.RequestPeer(m.p2pNetwork.GetBootNodes()[0], HandshakeCode, byt)
	if err != nil {
		return nil, err
	}
	return DecodeHsResp(respByt)
}

func (m *Kernel) compareMissingRange(remoteInfo *HandShakeInfo) (*BlocksRange, error) {
	localInfo, err := m.NewHsInfo()
	if err != nil {
		return nil, err
	}
	return localInfo.Compare(remoteInfo)
}

func (m *Kernel) getMissingBlocksTxns(remoteReq *HandShakeRequest) ([]byte, map[Hash][]byte, error) {
	fetchRange := remoteReq.FetchRange
	blocks, err := m.chain.GetRangeBlocks(fetchRange.StartHeight, fetchRange.EndHeight)
	if err != nil {
		return nil, nil, err
	}
	blocksByt, err := m.chain.EncodeBlocks(blocks)
	if err != nil {
		return nil, nil, err
	}

	txnsByt := make(map[Hash][]byte)
	for _, block := range blocks {
		blockHash := block.Hash
		txns, err := m.base.GetTxns(blockHash)
		if err != nil {
			return nil, nil, err
		}
		byt, err := types.FromArray(txns...).Encode()
		if err != nil {
			return nil, nil, err
		}
		txnsByt[blockHash] = byt
	}

	return blocksByt, txnsByt, nil
}

func (m *Kernel) pubUnpackedTxns(txns types.SignedTxns) error {
	byt, err := txns.Encode()
	if err != nil {
		return err
	}
	return m.p2pNetwork.PubP2P(UnpackedTxnsTopic, byt)
}

func (m *Kernel) subUnpackedTxns() (types.SignedTxns, error) {
	byt, err := m.p2pNetwork.SubP2P(UnpackedTxnsTopic)
	if err != nil {
		return nil, err
	}
	return types.DecodeSignedTxns(byt)
}
