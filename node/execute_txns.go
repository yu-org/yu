package node

import (
	. "yu/blockchain"
	. "yu/common"
	"yu/context"
	"yu/tripod"
)

func ExecuteTxns(block IBlock, base IBlockBase, land *tripod.Land) error {
	blockHash := block.Header().Hash()
	stxns, err := base.GetTxns(blockHash)
	if err != nil {
		return err
	}
	for _, stxn := range stxns {
		ecall := stxn.GetRaw().Ecall()
		ctx, err := context.NewContext(ecall.Params)
		if err != nil {
			return err
		}
		err = land.Execute(ecall, ctx)
		if err != nil {
			return err
		}

		for _, event := range ctx.Events {
			event.Height = block.Header().Height()
			event.BlockHash = blockHash
			event.ExecName = ecall.ExecName
			event.TripodName = ecall.TripodName
			event.BlockStage = ExecuteTxnsStage
			event.Caller = stxn.GetRaw().Caller()
		}

		for _, e := range ctx.Errors {
			e.Caller = stxn.GetRaw().Caller()
			e.BlockStage = ExecuteTxnsStage
			e.TripodName = ecall.TripodName
			e.ExecName = ecall.ExecName
			e.BlockHash = blockHash
			e.Height = block.Header().Height()
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
