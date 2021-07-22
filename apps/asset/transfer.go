package asset

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/context"
	. "github.com/Lawliet-Chan/yu/tripod"
	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
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

func (a *Asset) QueryBalance(ctx *Context, env *ChainEnv, _ Hash) (interface{}, error) {
	account := ctx.GetAddress("account")
	amount := a.getBalance(env, account)
	return amount, nil
}

func (a *Asset) Transfer(ctx *Context, _ IBlock, env *ChainEnv) (err error) {
	from := ctx.Caller
	to := ctx.GetAddress("to")
	amount := Amount(ctx.GetUint64("amount"))

	if !a.exsitAccount(env, from) {
		return AccountNotFound(from)
	}

	fromBalance := a.getBalance(env, from)
	if fromBalance < amount {
		return InsufficientFunds
	}

	if !a.exsitAccount(env, to) {
		a.setBalance(env, to, amount)
	} else {
		toBalance := a.getBalance(env, to)
		toBalance, err = checkAdd(toBalance, amount)
		if err != nil {
			return
		}
		a.setBalance(env, to, toBalance)
	}

	fromBalance, err = checkSub(fromBalance, amount)
	if err != nil {
		return
	}

	a.setBalance(env, from, fromBalance)

	_ = ctx.EmitEvent("Transfer Completed!")

	return
}

func (a *Asset) CreateAccount(ctx *Context, _ IBlock, env *ChainEnv) error {
	addr := ctx.Caller
	amount := ctx.GetUint64("amount")

	if a.exsitAccount(env, addr) {
		_ = ctx.EmitEvent("Account Exists!")
		return nil
	}

	a.setBalance(env, addr, Amount(amount))
	_ = ctx.EmitEvent("Account Created Success!")
	return nil
}

func (a *Asset) exsitAccount(env *ChainEnv, addr Address) bool {
	return env.KVDB.Exist(a, addr.Bytes())
}

func (a *Asset) getBalance(env *ChainEnv, addr Address) Amount {
	balanceByt, err := env.KVDB.Get(a, addr.Bytes())
	if err != nil {
		logrus.Panic("get balance error")
	}
	return MustDecodeToAmount(balanceByt)
}

func (a *Asset) setBalance(env *ChainEnv, addr Address, amount Amount) {
	env.KVDB.Set(a, addr.Bytes(), amount.MustEncode())
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
