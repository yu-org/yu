package poa

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/types"
)

var prefix = "\x19Ethereum Signed Message:\n"

func CheckMetamaskSig(txn *types.SignedTxn) error {
	wrCall := txn.Raw.WrCall
	hash, err := wrCall.Hash()
	if err != nil {
		return err
	}
	metamaskMsgHash := MetamaskMsgHash(hash)
	if len(txn.Signature) > 0 {
		txn.Signature[len(txn.Signature)-1] = 0
	}

	pubkey, err := crypto.Ecrecover(metamaskMsgHash, txn.Signature)
	if err != nil {
		return yerror.TxnSignatureIllegal(err)
	}

	return nil
}

func MetamaskMsgHash(hash []byte) []byte {
	hexHash := common.ToHex(hash)
	preambleStr := fmt.Sprintf("%s%d", prefix, len(hexHash))
	preamble := []byte(preambleStr)
	ethMsg := append(preamble, []byte(hexHash)...)
	return crypto.Keccak256(ethMsg)
}
