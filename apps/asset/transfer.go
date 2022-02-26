package asset

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
	"math/big"
)

type Asset struct {
	*DefaultTripod
	TokenName string
}

func NewAsset(tokenName string) *Asset {
	df := NewDefaultTripod("asset")

	a := &Asset{df, tokenName}
	a.SetExec(a.Transfer, 100).SetExec(a.CreateAccount, 10)
	a.SetQueries(a.QueryBalance)

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

	_ = ctx.EmitEvent("Transfer Completed!")

	return
}

func (a *Asset) CreateAccount(ctx *Context, _ *CompactBlock) error {
	addr := ctx.Caller
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
