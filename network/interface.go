package network

import (
	. "github.com/yu-org/yu/chain_env"
	"io"
)

type NetSync interface {
	ChooseBestNodes()
	HandleSyncReq(rw io.ReadWriter, env *ChainEnv) error
	SyncHistory(rw io.ReadWriter, env *ChainEnv) error
}
