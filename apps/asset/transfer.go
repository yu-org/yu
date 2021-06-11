package pow

import (
	"github.com/Lawliet-Chan/yu/chain_env"
	"github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/context"
)

func (p *Pow) Transfer(ctx *context.Context, env *chain_env.ChainEnv) error {
	sendAddr := ctx.Caller

	to, err := ctx.GetBytes("to")
	if err != nil {
		return err
	}
	toAddr := common.BytesToAddress(to)
	amount, err := ctx.GetUint64("amount")
	if err != nil {
		return err
	}


	return nil
}

func (p *Pow) CreateAccount(ctx *context.Context, env *chain_env.ChainEnv) error {
	addr := ctx.Caller

	amount, err := ctx.GetUint64("amount")
	if err != nil {
		return err
	}
	
}