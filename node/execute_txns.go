package node

import (
	. "yu/blockchain"
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
		err := land.Execute(ecall)
		if err != nil {
			return err
		}
	}
	return nil
}
