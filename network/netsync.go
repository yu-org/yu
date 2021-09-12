package network

import (
	"bufio"
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/chain_env"
	. "github.com/yu-altar/yu/common"
	"github.com/yu-altar/yu/node"
	. "github.com/yu-altar/yu/tripod"
	. "github.com/yu-altar/yu/txn"
	. "github.com/yu-altar/yu/yerror"
	"io"
)

type DefaultNetSync struct {
	land *Land
}

func NewDefaultNetSync(land *Land) *DefaultNetSync {
	return &DefaultNetSync{land: land}
}

func (d *DefaultNetSync) ChooseBestNodes() {
	panic("implement me")
}

func (d *DefaultNetSync) SyncHistory(rw io.ReadWriter, env *ChainEnv) error {
	resp, err := d.pushFetchReq(rw, env.Chain, nil)
	if err != nil {
		return err
	}
	if resp.Err != nil {
		return resp.Err
	}

	for resp.MissingRange != nil {
		// todo: the missing range maybe very huge and we need fetch them multiple times
		// the remote node will return new Missing blocks-range in this response.
		resp, err = d.pushFetchReq(rw, env.Chain, resp.MissingRange)
		if err != nil {
			return err
		}
		if resp.Err != nil {
			return resp.Err
		}

		blocks, err := env.Chain.DecodeBlocks(resp.BlocksByt)
		if err != nil {
			return err
		}

		err = d.syncHistoryBlocks(env, blocks)
		if err != nil {
			return err
		}

		resp.MissingRange = nil

		for blockHash, byt := range resp.TxnsByt {
			txns, err := DecodeSignedTxns(byt)
			if err != nil {
				return err
			}
			err = env.Base.SetTxns(blockHash, txns)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DefaultNetSync) HandleSyncReq(rw io.ReadWriter, env *ChainEnv) error {
	byt, err := ReadFrom(rw)
	if err != nil {
		return err
	}

	remoteReq, err := DecodeHsRequest(byt)
	if err != nil {
		return err
	}

	var (
		blocksByt []byte
		txnsByt   map[Hash][]byte
	)
	if remoteReq.FetchRange != nil {
		blocksByt, txnsByt, err = getMissingBlocksTxns(env, remoteReq)
		if err != nil {
			return err
		}
	}

	missingRange, err := compareMissingRange(env.Chain, remoteReq.Info)

	hsResp := &HandShakeResp{
		MissingRange: missingRange,
		BlocksByt:    blocksByt,
		TxnsByt:      txnsByt,
		Err:          err,
	}
	byt, err = hsResp.Encode()
	if err != nil {
		return err
	}

	return WriteTo(byt, rw)
}

func (d *DefaultNetSync) pushFetchReq(rw io.ReadWriter, chain IBlockChain, fetchRange *BlocksRange) (*HandShakeResp, error) {
	hs, err := NewHsReq(chain, fetchRange)
	if err != nil {
		return nil, err
	}

	byt, err := hs.Encode()
	if err != nil {
		return nil, err
	}
	err = WriteTo(byt, rw)
	if err != nil {
		return nil, err
	}
	respByt, err := ReadFrom(rw)
	if err != nil {
		return nil, err
	}
	return DecodeHsResp(respByt)
}

// get the missing range of remote node
func getMissingBlocksTxns(env *ChainEnv, remoteReq *HandShakeRequest) ([]byte, map[Hash][]byte, error) {
	fetchRange := remoteReq.FetchRange
	blocks, err := env.Chain.GetRangeBlocks(fetchRange.StartHeight, fetchRange.EndHeight)
	if err != nil {
		return nil, nil, err
	}
	blocksByt, err := env.Chain.EncodeBlocks(blocks)
	if err != nil {
		return nil, nil, err
	}

	txnsByt := make(map[Hash][]byte)
	for _, block := range blocks {
		blockHash := block.GetHash()
		txns, err := env.Base.GetTxns(blockHash)
		if err != nil {
			return nil, nil, err
		}
		byt, err := FromArray(txns...).Encode()
		if err != nil {
			return nil, nil, err
		}
		txnsByt[blockHash] = byt
	}

	return blocksByt, txnsByt, nil
}

func (d *DefaultNetSync) syncHistoryBlocks(env *ChainEnv, blocks []IBlock) error {
	switch env.RunMode {
	case LocalNode:
		for _, block := range blocks {
			err := d.land.RangeList(func(tri Tripod) error {
				if tri.ValidateBlock(block, env) {
					return env.Chain.AppendBlock(block)
				}
				return BlockIllegal(block.GetHash())
			})
			if err != nil {
				return err
			}
		}

		return d.executeChainTxns(env)

	case MasterWorker:
		// todo
		return nil
	default:
		return NoRunMode
	}
}

func (d *DefaultNetSync) executeChainTxns(env *ChainEnv) error {
	chain, err := env.Chain.Chain()
	if err != nil {
		return err
	}
	return chain.Range(func(block IBlock) error {
		return node.ExecuteTxns(block, env, d.land)
	})
}

func compareMissingRange(chain IBlockChain, remoteInfo *HandShakeInfo) (*BlocksRange, error) {
	localInfo, err := NewHsInfo(chain)
	if err != nil {
		return nil, err
	}
	return localInfo.Compare(remoteInfo)
}

func ReadFrom(r io.Reader) ([]byte, error) {
	return bufio.NewReader(r).ReadBytes('\n')
}

func WriteTo(data []byte, w io.Writer) error {
	data = append(data, '\n')
	_, err := w.Write(data)
	return err
}
