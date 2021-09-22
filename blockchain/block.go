package blockchain

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/trie"
	"github.com/yu-org/yu/txn"
	. "github.com/yu-org/yu/utils/codec"
)

type Block struct {
	Header     *Header
	TxnsHashes []Hash

	ChainLen uint64
}

func (b *Block) GetHeight() BlockNum {
	return b.Header.GetHeight()
}

func (b *Block) GetHash() Hash {
	return b.Header.GetHash()
}

func (b *Block) GetPrevHash() Hash {
	return b.Header.GetPrevHash()
}

func (b *Block) GetTxnRoot() Hash {
	return b.Header.GetTxnRoot()
}

func (b *Block) GetStateRoot() Hash {
	return b.Header.GetStateRoot()
}

func (b *Block) GetTimestamp() uint64 {
	return b.Header.GetTimestamp()
}

func (b *Block) GetHeader() IHeader {
	return b.Header
}

func (b *Block) GetTxnsHashes() []Hash {
	return b.TxnsHashes
}

func (b *Block) SetTxnsHashes(hashes []Hash) {
	b.TxnsHashes = hashes
}

func (b *Block) SetHash(hash Hash) {
	b.Header.Hash = hash
}

func (b *Block) SetPreHash(preHash Hash) {
	b.Header.PrevHash = preHash
}

func (b *Block) SetTxnRoot(hash Hash) {
	b.Header.TxnRoot = hash
}

func (b *Block) SetHeight(height BlockNum) {
	b.Header.Height = height
}

func (b *Block) GetBlockId() BlockId {
	return NewBlockId(b.Header.GetHeight(), b.Header.GetHash())
}

func (b *Block) SetStateRoot(hash Hash) {
	b.Header.StateRoot = hash
}

func (b *Block) SetTimestamp(ts uint64) {
	b.Header.Timestamp = ts
}

func (b *Block) GetPeerID() peer.ID {
	return b.Header.PeerID
}

func (b *Block) SetPeerID(peerID peer.ID) {
	b.Header.PeerID = peerID
}

func (b *Block) GetLeiLimit() uint64 {
	return b.Header.LeiLimit
}

func (b *Block) SetLeiLimit(e uint64) {
	b.Header.LeiLimit = e
}

func (b *Block) GetLeiUsed() uint64 {
	return b.Header.LeiUsed
}

func (b *Block) UseLei(e uint64) {
	b.Header.LeiUsed += e
}

//
//func (b *Block) SetSign(sign []byte) {
//	b.Header.Signature = sign
//}
//
//func (b *Block) GetSign() []byte {
//	return b.Header.GetSign()
//}
//
//func (b *Block) SetPubkey(key PubKey) {
//	b.Header.Pubkey = key.BytesWithType()
//}
//
//func (b *Block) GetPubkey() PubKey {
//	return b.Header.GetPubkey()
//}

func (b *Block) SetNonce(nonce uint64) {
	b.Header.Nonce = nonce
}

func (b *Block) SetChainLen(len uint64) {
	b.ChainLen = len
}

func (b *Block) Encode() ([]byte, error) {
	return GlobalCodec.EncodeToBytes(b)
}

func (b *Block) Decode(data []byte) (IBlock, error) {
	var block Block
	err := GlobalCodec.DecodeBytes(data, &block)
	return &block, err
}

func (b *Block) CopyFrom(other IBlock) {
	otherBlock := other.(*Block)
	*b = *otherBlock
}

func IfLeiOut(Lei uint64, block IBlock) bool {
	return Lei+block.GetLeiUsed() > block.GetLeiLimit()
}

func MakeTxnRoot(txns []*txn.SignedTxn) (Hash, error) {
	txnsBytes := make([]Hash, 0)
	for _, tx := range txns {
		hash := tx.GetTxnHash()
		txnsBytes = append(txnsBytes, hash)
	}
	mTree := trie.NewMerkleTree(txnsBytes)
	return mTree.RootNode.Data, nil
}

//func getPubkeyStrOrNil(b IBlock) (pubkeyStr string) {
//	pubkey := b.GetPubkey()
//	if pubkey == nil {
//		return
//	}
//	return pubkey.StringWithType()
//}

//
//func (b *Block) getSignStr() (signStr string) {
//	sign := b.GetSign()
//	if sign == nil {
//		return
//	}
//	return ToHex(sign)
//}
