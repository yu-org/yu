package node

import (
	. "yu/blockchain"
	. "yu/common"
	"yu/context"
	"yu/tripod"
)

func ExecuteTxns(block IBlock, base IBlockBase, land *tripod.Land, sub *Subscription) error {
	blockHash := block.GetHeader().GetHash()
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
		err = land.Execute(ecall, ctx, block)
		if err != nil {
			return err
		}

		for _, event := range ctx.Events {
			event.Height = block.GetHeader().GetHeight()
			event.BlockHash = blockHash
			event.ExecName = ecall.ExecName
			event.TripodName = ecall.TripodName
			event.BlockStage = ExecuteTxnsStage
			event.Caller = stxn.GetRaw().GetCaller()

			if sub != nil {
				sub.Push(event)
			}
		}

		for _, e := range ctx.Errors {
			e.Caller = stxn.GetRaw().GetCaller()
			e.BlockStage = ExecuteTxnsStage
			e.TripodName = ecall.TripodName
			e.ExecName = ecall.ExecName
			e.BlockHash = blockHash
			e.Height = block.GetHeader().GetHeight()

			if sub != nil {
				sub.Push(e)
			}
		}

		err = base.SetEvents(ctx.Events)
		if err != nil {
			return err
		}
		err = base.SetErrors(ctx.Errors)
		if err != nil {
			return err
		}
	}
	return nil
}
