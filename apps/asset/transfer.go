package asset

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
	"math/big"
)

type Asset struct {
	*DefaultTripod
	validators []PubKey
	TokenName  string
}

func NewAsset(tokenName string, validators []PubKey) *Asset {
	df := NewDefaultTripod("asset")

	a := &Asset{df, validators, tokenName}
	a.SetExec(a.Transfer, 100).SetExec(a.CreateAccount, 10)
	a.SetQueries(a.QueryBalance)

	a.SetTxnChecker(func(txn *SignedTxn) error {
		if txn.Raw.Ecall.LeiPrice == 0 {
			return nil
		}

		if !a.existAccount(txn.Raw.Caller) {
			return AccountNotFound(txn.Raw.Caller)
		}

		balance := a.getBalance(txn.Raw.Caller)
		leiPrice := new(big.Int).SetUint64(txn.Raw.Ecall.LeiPrice)
		if balance.Cmp(leiPrice) < 0 {
			return InsufficientFunds
		}

		validatorsCount := len(validators)
		if validatorsCount > 0 {
			validatorsCountBigInt := new(big.Int).SetInt64(int64(validatorsCount))
			rewards := new(big.Int).Div(leiPrice, validatorsCountBigInt)
			for _, validator := range validators {
				err := a.transfer(txn.Raw.Caller, validator.Address(), rewards)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return a
}

func (a *Asset) QueryBalance(ctx *Context, _ Hash) (interface{}, error) {
	account := ctx.GetAddress("account")
	if !a.existAccount(account) {
		return nil, AccountNotFound(account)
	}
	amount := a.getBalance(account)
	return amount, nil
}

func (a *Asset) Transfer(ctx *Context, _ *CompactBlock) (err error) {
	from := ctx.Caller
	to := ctx.GetAddress("to")
	amount := big.NewInt(int64(ctx.GetUint64("amount")))

	err = a.transfer(from, to, amount)
	if err != nil {
		return
	}
	_ = ctx.EmitEvent("Transfer Completed!")

	return
}

func (a *Asset) transfer(from, to Address, amount *big.Int) error {
	if !a.existAccount(from) {
		return AccountNotFound(from)
	}

	fromBalance := a.getBalance(from)
	if fromBalance.Cmp(amount) < 0 {
		return InsufficientFunds
	}

	if !a.existAccount(to) {
		a.setBalance(to, amount)
	} else {
		toBalance := a.getBalance(to)
		toAdd := new(big.Int).Add(toBalance, amount)
		a.setBalance(to, toAdd)
	}

	fromSub := new(big.Int).Sub(fromBalance, amount)
	a.setBalance(from, fromSub)
	return nil
}

func (a *Asset) CreateAccount(ctx *Context, _ *CompactBlock) error {
	addr := ctx.Caller
	//if !a.isValidator(addr) {
	//	return NoPermission
	//}
	amount := big.NewInt(int64(ctx.GetUint64("amount")))

	logrus.Debugf("Create ACCOUNT(%s) amount(%d)", addr.String(), amount)

	if a.existAccount(addr) {
		_ = ctx.EmitEvent("Account Exists!")
		return nil
	}

	a.setBalance(addr, amount)
	_ = ctx.EmitEvent("Account Created Success!")
	return nil
}

func (a *Asset) existAccount(addr Address) bool {
	return a.State.Exist(a, addr.Bytes())
}

func (a *Asset) getBalance(addr Address) *big.Int {
	balanceByt, err := a.State.Get(a, addr.Bytes())
	if err != nil {
		logrus.Panic("get balance error: ", err)
	}

	b := new(big.Int)
	err = b.UnmarshalText(balanceByt)
	if err != nil {
		logrus.Panic("getBalance marshal error: ", err)
	}
	return b
}

func (a *Asset) setBalance(addr Address, amount *big.Int) {
	amountText, err := amount.MarshalText()
	if err != nil {
		logrus.Panic("amount marshal error: ", err)
	}

	a.State.Set(a, addr.Bytes(), amountText)
}

func (a *Asset) isValidator(addr Address) bool {
	for _, validator := range a.validators {
		if validator.Address() == addr {
			return true
		}
	}
	return false
}
