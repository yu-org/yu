package master

import (
	"encoding/json"
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/yerror"
	"github.com/libp2p/go-libp2p-core/peer"
	"strconv"
)

const (
	HandshakeType int = iota
	SyncTxnsType

	RequestTypeBytesLen = 1
)

var (
	HandshakeReqByt = []byte(strconv.Itoa(HandshakeType))
	SyncTxnsReqByt  = []byte(strconv.Itoa(SyncTxnsType))
)

type HandShakeRequest struct {
	FetchRange *BlocksRange
	Info       *HandShakeInfo
}

func (m *Master) NewHsReq(fetchRange *BlocksRange) (*HandShakeRequest, error) {
	info, err := m.NewHsInfo()
	if err != nil {
		return nil, err
	}
	return &HandShakeRequest{
		FetchRange: fetchRange,
		Info:       info,
	}, nil
}

func (hs *HandShakeRequest) Encode() ([]byte, error) {
	byt, err := json.Marshal(hs)
	if err != nil {
		return nil, err
	}
	return append(HandshakeReqByt, byt...), nil
}

func DecodeHsRequest(data []byte) (*HandShakeRequest, error) {
	var hs HandShakeRequest
	err := json.Unmarshal(data[RequestTypeBytesLen:], &hs)
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

func (m *Master) NewHsInfo() (*HandShakeInfo, error) {
	gBlock, err := m.chain.GetGenesis()
	if err != nil {
		return nil, err
	}

	eBlock, err := m.chain.GetEndBlock()
	if err != nil {
		return nil, err
	}

	return &HandShakeInfo{
		GenesisBlockHash: gBlock.GetHash(),
		EndHeight:        eBlock.GetHeight(),
		EndBlockHash:     eBlock.GetHash(),
	}, nil
}

// return a BlocksRange if other node's height is lower
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
	// compressed blocks bytes
	BlocksByt []byte
	// key: block-hash,
	// value: compressed txns bytes
	TxnsByt map[Hash][]byte
	Err     error
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

//type PackedTxns struct {
//	BlockHash string
//	TxnsBytes []byte
//}
//
//func NewPackedTxns(blockHash Hash, txns SignedTxns) (*PackedTxns, error) {
//	byt, err := txns.Encode()
//	if err != nil {
//		return nil, err
//	}
//	return &PackedTxns{
//		BlockHash: blockHash.String(),
//		TxnsBytes: byt,
//	}, nil
//}
//
//func (pt *PackedTxns) Resolve() (Hash, SignedTxns, error) {
//	stxns, err := DecodeSignedTxns(pt.TxnsBytes)
//	if err != nil {
//		return NullHash, nil, err
//	}
//	return HexToHash(pt.BlockHash), stxns, nil
//}

type TxnsRequest struct {
	Hashes        []Hash
	BlockProducer peer.ID
}

func (tr TxnsRequest) Encode() ([]byte, error) {
	byt, err := json.Marshal(tr)
	if err != nil {
		return nil, err
	}
	return append(SyncTxnsReqByt, byt...), nil
}

func DecodeTxnsRequest(data []byte) (tr TxnsRequest, err error) {
	err = json.Unmarshal(data[RequestTypeBytesLen:], &tr)
	return
}
