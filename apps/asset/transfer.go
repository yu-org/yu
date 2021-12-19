package asset

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/chain_env"
	. "github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
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
	amount := a.getBalance(a.ChainEnv, account)
	return amount, nil
}

func (a *Asset) Transfer(ctx *Context, _ *CompactBlock) (err error) {
	from := ctx.Caller
	to := ctx.GetAddress("to")
	amount := Amount(ctx.GetUint64("amount"))

	if !a.exsitAccount(a.ChainEnv, from) {
		return AccountNotFound(from)
	}

	fromBalance := a.getBalance(a.ChainEnv, from)
	if fromBalance < amount {
		return InsufficientFunds
	}

	if !a.exsitAccount(a.ChainEnv, to) {
		a.setBalance(a.ChainEnv, to, amount)
	} else {
		toBalance := a.getBalance(a.ChainEnv, to)
		toBalance, err = checkAdd(toBalance, amount)
		if err != nil {
			return
		}
		a.setBalance(a.ChainEnv, to, toBalance)
	}

	fromBalance, err = checkSub(fromBalance, amount)
	if err != nil {
		return
	}

	a.setBalance(a.ChainEnv, from, fromBalance)

	_ = ctx.EmitEvent("Transfer Completed!")

	return
}

func (a *Asset) CreateAccount(ctx *Context, _ *CompactBlock) error {
	addr := ctx.Caller
	amount := ctx.GetUint64("amount")

	if a.exsitAccount(a.ChainEnv, addr) {
		_ = ctx.EmitEvent("Account Exists!")
		return nil
	}

	a.setBalance(a.ChainEnv, addr, Amount(amount))
	_ = ctx.EmitEvent("Account Created Success!")
	return nil
}

func (a *Asset) exsitAccount(env *ChainEnv, addr Address) bool {
	return env.Exist(a, addr.Bytes())
}

func (a *Asset) getBalance(env *ChainEnv, addr Address) Amount {
	balanceByt, err := env.Get(a, addr.Bytes())
	if err != nil {
		logrus.Panic("get balance error")
	}
	return MustDecodeToAmount(balanceByt)
}

func (a *Asset) setBalance(env *ChainEnv, addr Address, amount Amount) {
	env.Set(a, addr.Bytes(), amount.MustEncode())
}

func checkAdd(origin, add Amount) (Amount, error) {
	result := origin + add
	if result < origin && result < add {
		return 0, IntegerOverflow
	}
	return result, nil
}

func checkSub(origin, sub Amount) (Amount, error) {
	if origin < sub {
		return 0, IntegerOverflow
	}
	return origin - sub, nil
}
