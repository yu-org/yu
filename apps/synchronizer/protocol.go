package synchronizer

import (
	"bytes"
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
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(hs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeHsRequest(data []byte) (*HandShakeRequest, error) {
	var hs HandShakeRequest
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(&hs)
	return &hs, err
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

	eBlock, err := b.Chain.GetEndCompactBlock()
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
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(hs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeHsResp(data []byte) (*HandShakeResp, error) {
	var hs HandShakeResp
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(&hs)
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
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	err = decoder.Decode(&tr)
	return
}
