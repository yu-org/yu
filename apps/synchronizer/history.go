package synchronizer

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/tripod"
)

const (
	FullSync int = iota
	FastSync
	LightSync
)

type Synchronizer struct {
	*Tripod
	syncMode int
}

func NewSynchronizer(syncMode int) *Synchronizer {
	tri := NewTripodWithName("synchronizer")
	fh := &Synchronizer{Tripod: tri, syncMode: syncMode}
	tri.SetInit(fh)
	tri.SetP2pHandler(HandshakeCode, fh.handleHsReq).SetP2pHandler(SyncTxnsCode, fh.handleSyncTxnsReq)
	return fh
}

func (b *Synchronizer) InitChain() {
	b.syncHistory()
}

func (b *Synchronizer) syncHistory() {
	if len(b.P2pNetwork.GetBootNodes()) == 0 {
		return
	}
	switch b.syncMode {
	case FullSync:
		err := b.syncFullHistory()
		if err != nil {
			logrus.Panic("sync full history failed, err: ", err)
		}
	case FastSync:

	case LightSync:

	}
}
