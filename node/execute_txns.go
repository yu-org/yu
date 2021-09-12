package node

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-altar/yu/blockchain"
	"github.com/yu-altar/yu/chain_env"
	. "github.com/yu-altar/yu/common"
	"github.com/yu-altar/yu/context"
	. "github.com/yu-altar/yu/subscribe"
	. "github.com/yu-altar/yu/tripod"
	"github.com/yu-altar/yu/txn"
	"github.com/yu-altar/yu/yerror"
)

func ExecuteTxns(block IBlock, env *chain_env.ChainEnv, land *Land) error {
	base := env.Base
	sub := env.Sub

	stxns, err := base.GetTxns(block.GetHash())
	if err != nil {
		return err
	}
	for _, stxn := range stxns {
		ecall := stxn.GetRaw().GetEcall()
		ctx, err := context.NewContext(stxn.GetPubkey().Address(), ecall.Params)
		if err != nil {
			return err
		}

		exec, lei, err := land.GetExecLei(ecall)
		if err != nil {
			handleError(err, ctx, block, stxn, sub)
			continue
		}

		if IfLeiOut(lei, block) {
			handleError(yerror.OutOfEnergy, ctx, block, stxn, sub)
			break
		}

		err = exec(ctx, block, env)
		if err != nil {
			env.Discard()
			handleError(err, ctx, block, stxn, sub)
		} else {
			env.NextTxn()
		}

		block.UseLei(lei)

		handleEvent(ctx, block, stxn, sub)

		err = base.SetEvents(ctx.Events)
		if err != nil {
			return err
		}
		err = base.SetError(ctx.Error)
		if err != nil {
			return err
		}
	}

	stateRoot, err := env.Commit()
	if err != nil {
		return err
	}
	block.SetStateRoot(stateRoot)

	return nil
}

func handleError(err error, ctx *context.Context, block IBlock, stxn *txn.SignedTxn, sub *Subscription) {
	ctx.EmitError(err)
	ecall := stxn.GetRaw().GetEcall()

	ctx.Error.Caller = stxn.GetRaw().GetCaller()
	ctx.Error.BlockStage = ExecuteTxnsStage
	ctx.Error.TripodName = ecall.TripodName
	ctx.Error.ExecName = ecall.ExecName
	ctx.Error.BlockHash = block.GetHash()
	ctx.Error.Height = block.GetHeight()

	logrus.Error("push error: ", ctx.Error.Error())
	if sub != nil {
		sub.Push(ctx.Error)
	}

}

func handleEvent(ctx *context.Context, block IBlock, stxn *txn.SignedTxn, sub *Subscription) {
	for _, event := range ctx.Events {
		ecall := stxn.GetRaw().GetEcall()

		event.Height = block.GetHeight()
		event.BlockHash = block.GetHash()
		event.ExecName = ecall.ExecName
		event.TripodName = ecall.TripodName
		event.BlockStage = ExecuteTxnsStage
		event.Caller = stxn.GetRaw().GetCaller()

		if sub != nil {
			sub.Push(event)
		}
	}
}
