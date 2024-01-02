package synchronizer

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p/core/peer"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
)

const (
	HandshakeCode int = 100
	SyncTxnsCode      = 101
)

type HandShakeRequest struct {
	FetchRange *BlocksRange
	Info       *HandShakeInfo
}

func (b *Synchronizer) NewHsReq(fetchRange *BlocksRange) (*HandShakeRequest, error) {
	info, err := b.NewHsInfo()
	if err != nil {
		return nil, err
	}
	return &HandShakeRequest{
		FetchRange: fetchRange,
		Info:       info,
	}, nil
}

func (hs *HandShakeRequest) Encode() ([]byte, error) {
	return json.Marshal(hs)
}

func DecodeHsRequest(data []byte) (*HandShakeRequest, error) {
	var hs HandShakeRequest
	err := json.Unmarshal(data, &hs)
	if err != nil {
		return nil, err
	}
	return &hs, nil
}

type HandShakeInfo struct {
	GenesisBlockHash Hash

	// when chain is finlaized chain, end block is the finalized block
	EndHeight    BlockNum
	EndBlockHash Hash
}

func (b *Synchronizer) NewHsInfo() (*HandShakeInfo, error) {
	gBlock, err := b.Chain.GetGenesis()
	if err != nil {
		return nil, err
	}

	eBlock, err := b.Chain.GetEndBlock()
	if err != nil {
		return nil, err
	}

	return &HandShakeInfo{
		GenesisBlockHash: gBlock.Hash,
		EndHeight:        eBlock.Height,
		EndBlockHash:     eBlock.Hash,
	}, nil
}

// Compare return a BlocksRange if other node's height is lower
func (hs *HandShakeInfo) Compare(other *HandShakeInfo) (*BlocksRange, error) {
	if hs.GenesisBlockHash != other.GenesisBlockHash {
		return nil, yerror.GenesisBlockIllegal
	}

	if hs.EndHeight > other.EndHeight {
		return &BlocksRange{
			StartHeight: other.EndHeight + 1,
			EndHeight:   hs.EndHeight,
		}, nil
	}

	return nil, nil
}

type HandShakeResp struct {
	// missing blocks range
	MissingRange *BlocksRange
	// blocks bytes
	BlocksByt []byte
	Err       error
}

func (hs *HandShakeResp) Encode() ([]byte, error) {
	return json.Marshal(hs)
}

func DecodeHsResp(data []byte) (*HandShakeResp, error) {
	var hs HandShakeResp
	err := json.Unmarshal(data, &hs)
	return &hs, err
}

type BlocksRange struct {
	StartHeight BlockNum
	EndHeight   BlockNum
}

type TxnsRequest struct {
	Hashes        []Hash
	BlockProducer peer.ID
}

func (tr TxnsRequest) Encode() ([]byte, error) {
	return json.Marshal(tr)
}

func DecodeTxnsRequest(data []byte) (tr TxnsRequest, err error) {
	err = json.Unmarshal(data, &tr)
	return
}
