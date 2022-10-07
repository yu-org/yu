package yerror

import (
	"github.com/pkg/errors"
	"github.com/yu-org/yu/common"
	"math/big"
)

var (
	InsufficientFunds = errors.New("Insufficient Funds")
	NoPermission      = errors.New("No Permission")
)

type ErrAccountNotFound struct {
	account string
}

func (an ErrAccountNotFound) Error() string {
	return errors.Errorf("account(%s) not found", an.account).Error()
}

func AccountNotFound(addr common.Address) ErrAccountNotFound {
	return ErrAccountNotFound{account: addr.String()}
}

type ErrAmountNeg struct {
	amount *big.Int
}

func (an ErrAmountNeg) Error() string {
	return errors.Errorf("amount(%d) is negative", an.amount).Error()
}

func AmountNeg(amount *big.Int) ErrAmountNeg {
	return ErrAmountNeg{amount: amount}
}

// hotstuff errors
var (
	NoValidQC       = errors.New("Target QC is empty.")
	NoValidParentId = errors.New("ParentId is empty.")
)
