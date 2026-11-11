package kernel

import (
	"fmt"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/types"
)

// reconcileStateFinality repairs a crash between durable state finalization and
// chain-marker finalization. It never guesses a branch from height alone.
func (k *Kernel) reconcileStateFinality() error {
	backend, ok := k.State.(state.FinalizedStateHead)
	if !ok {
		return nil
	}
	hash, root, err := backend.FinalizedStateHead()
	if err != nil {
		return err
	}
	tip, err := k.Chain.GetBlock(hash)
	if err != nil {
		return fmt.Errorf("finalized state block is missing: %w", err)
	}
	if (tip.Height != 0 || tip.StateRoot != common.NullHash) && tip.StateRoot != root {
		return fmt.Errorf("finalized block and SQL state roots disagree")
	}
	last, err := k.Chain.LastFinalized()
	if err != nil {
		return err
	}
	var blocks []*types.Block
	for tip.Hash != last.Hash {
		if tip.Height <= last.Height {
			return fmt.Errorf("chain finality is ahead of or conflicts with SQL state")
		}
		blocks = append(blocks, tip)
		tip, err = k.Chain.GetBlock(tip.PrevHash)
		if err != nil {
			return err
		}
	}
	for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}
	if len(blocks) == 0 {
		return nil
	}
	return k.Chain.FinalizeFork(&types.Fork{Blocks: blocks})
}
