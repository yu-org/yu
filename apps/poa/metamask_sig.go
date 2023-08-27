package poa

import (
	"fmt"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/types"
)

var prefix = "\u0019Ethereum Signed Message:\n"
var msgLength = "130"

func CheckMetamaskSig(txn *types.SignedTxn) error {
	wrCall := txn.Raw.WrCall
	hash, err := wrCall.Hash()
	if err != nil {
		return err
	}
	metamaskMsg := fmt.Sprintf("%s%s%s", prefix, msgLength, common.ToHex(hash))

	if !txn.Pubkey.VerifySignature([]byte(metamaskMsg), txn.Signature) {
		return yerror.TxnSignatureErr
	}
	return nil
}
