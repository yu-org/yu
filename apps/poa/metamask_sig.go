package poa

import (
	"fmt"
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
	metamaskMsg := MetamaskMsg(hash)

	if !txn.Pubkey.VerifySignature(metamaskMsg, txn.Signature) {
		return yerror.TxnSignatureErr
	}
	return nil
}

func MetamaskMsg(hash []byte) []byte {
	hexHash := common.ToHex(hash)
	preambleStr := fmt.Sprintf("%s%d", prefix, len(hexHash))
	fmt.Println(preambleStr)
	preamble := []byte(preambleStr)
	fmt.Println("preable bytes: ", preamble)
	ethMsg := append(preamble, hash...)
	// return []byte(fmt.Sprintf("%s%d%s", prefix, len(hexHash), hexHash))
	return ethMsg
}
