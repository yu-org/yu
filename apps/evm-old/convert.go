package evm_old

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/yu-org/yu/core/types"
	"math/big"
)

func HeaderToGeth(header *types.Header) *gtypes.Header {
	return &gtypes.Header{
		ParentHash:  gcommon.Hash(header.PrevHash),
		UncleHash:   gcommon.Hash{},
		Coinbase:    gcommon.Address{},
		Root:        gcommon.Hash(header.StateRoot),
		TxHash:      gcommon.Hash(header.TxnRoot),
		ReceiptHash: gcommon.Hash{},
		Bloom:       gtypes.Bloom{},
		Difficulty:  nil,
		Number:      big.NewInt(int64(header.Height)),
		GasLimit:    header.LeiLimit,
		GasUsed:     header.LeiUsed,
		Time:        header.Timestamp,
		Extra:       header.Extra,
		MixDigest:   gcommon.Hash{},
		Nonce:       gtypes.BlockNonce{},
	}
}
