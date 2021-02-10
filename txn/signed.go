package txn

import (
	. "yu/common"
	"yu/keypair"
)

type SignedTxn struct {
	Raw       UnsignedTxn
	TxnHash   Hash
	Pubkey    keypair.PubKey
	Signature []byte
}
