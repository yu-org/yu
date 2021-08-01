package yerror

import (
	"github.com/Lawliet-Chan/yu/common"
	"github.com/pkg/errors"
)

var InsufficientFunds = errors.New("Insufficient Funds")

type ErrAccountNotFound struct {
	account string
}

func (an ErrAccountNotFound) Error() string {
	return errors.Errorf("account(%s) not found", an.account).Error()
}

func AccountNotFound(addr common.Address) ErrAccountNotFound {
	return ErrAccountNotFound{account: addr.String()}
}

// hotstuff errors
var (
	NoValidQC       = errors.New("Target QC is empty.")
	NoValidParentId = errors.New("ParentId is empty.")
)
