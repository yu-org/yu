package node

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	"github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/context"
	. "github.com/Lawliet-Chan/yu/tripod"
)

func ExecuteTxns(block IBlock, env *chain_env.ChainEnv, land *Land) error {
	chain := env.Chain
	base := env.Base
	sub := env.Sub

	blockHash := block.GetHash()
	stxns, err := base.GetTxns(blockHash)
	if err != nil {
		return err
	}
	for _, stxn := range stxns {
		ecall := stxn.GetRaw().GetEcall()
		ctx, err := context.NewContext(stxn.GetPubkey().Address(), ecall.Params)
		if err != nil {
			return err
		}
		err = land.Execute(ecall, ctx, env)
		if err != nil {
			ctx.EmitError(err)
		}

		err = chain.UpdateBlock(block)
		if err != nil {
			return err
		}

		for _, event := range ctx.Events {
			event.Height = block.GetHeight()
			event.BlockHash = blockHash
			event.ExecName = ecall.ExecName
			event.TripodName = ecall.TripodName
			event.BlockStage = ExecuteTxnsStage
			event.Caller = stxn.GetRaw().GetCaller()

			if sub != nil {
				sub.Push(event)
			}
		}
		if ctx.Error != nil {
			ctx.Error.Caller = stxn.GetRaw().GetCaller()
			ctx.Error.BlockStage = ExecuteTxnsStage
			ctx.Error.TripodName = ecall.TripodName
			ctx.Error.ExecName = ecall.ExecName
			ctx.Error.BlockHash = blockHash
			ctx.Error.Height = block.GetHeight()
		}

		if sub != nil {
			sub.Push(ctx.Error)
		}

		err = base.SetEvents(ctx.Events)
		if err != nil {
			return err
		}
		err = base.SetError(ctx.Error)
		if err != nil {
			return err
		}
	}
	return nil
}
