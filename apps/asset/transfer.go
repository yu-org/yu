package asset

import (
	. "github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/context"
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

	return &Asset{df, tokenName}
}

func (a *Asset) Transfer(ctx *context.Context, env *ChainEnv) (err error) {
	from := ctx.Caller
	to := BytesToAddress(ctx.GetBytes("to"))
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

	return
}

func (a *Asset) CreateAccount(ctx *context.Context, env *ChainEnv) error {
	addr := ctx.Caller
	amount := ctx.GetUint64("amount")

	if env.KVDB.Exist(a, addr.Bytes()) {
		return nil
	}

	a.setBalance(env, addr, Amount(amount))
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
