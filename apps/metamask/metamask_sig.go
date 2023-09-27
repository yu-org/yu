package metamask

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/types"
)

var prefix = "\x19Ethereum Signed Message:\n"

func CheckMetamaskSig(txn *types.SignedTxn) error {
	wrCall := txn.Raw.WrCall
	msgByt, err := json.Marshal(wrCall)
	if err != nil {
		return err
	}
	metamaskMsgHash := MetamaskMsgHash(msgByt)
	if len(txn.Signature) > 0 {
		// for eth sig.v
		txn.Signature[len(txn.Signature)-1] = 1
	}

	crypto.VerifySignature(txn.Pubkey, metamaskMsgHash, txn.Signature)
	if err != nil {
		return yerror.TxnSignatureIllegal(err)
	}

	//pubkey, err := crypto.Ecrecover(metamaskMsgHash, txn.Signature)
	//if err != nil {
	//	return yerror.TxnSignatureIllegal(err)
	//}
	//if !bytes.Equal(pubkey, txn.Pubkey) {
	//	return errors.Errorf("pubkey mismatch: ec_recover pubkey: %x, expected pubkey %x", pubkey, txn.Pubkey)
	//}

	return nil
}

func MetamaskMsgHash(hash []byte) []byte {
	hexHash := common.ToHex(hash)
	preambleStr := fmt.Sprintf("%s%d", prefix, len(hexHash))
	preamble := []byte(preambleStr)
	ethMsg := append(preamble, []byte(hexHash)...)
	return crypto.Keccak256(ethMsg)
}
