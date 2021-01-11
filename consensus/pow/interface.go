package pow

import "yu/common"

type PoW interface {
	Run() (nonce int, hash common.Hash, err error)
	Validate() bool
}
