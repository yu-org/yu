package history

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/tripod"
)

const (
	Full = iota
	Snapshot
	Light
)

type History struct {
	*DefaultTripod
	mode int
}

func NewHistory(mode int) *History {
	tri := NewDefaultTripod("full_history")
	fh := &History{DefaultTripod: tri, mode: mode}
	tri.SetInit(fh)
	tri.SetP2pHandler(HandshakeCode, fh.handleHsReq).SetP2pHandler(SyncTxnsCode, fh.handleSyncTxnsReq)
	return fh
}

func (h *History) InitChain() {
	if len(h.P2pNetwork.GetBootNodes()) == 0 {
		return
	}
	switch h.mode {
	case Full:
		err := h.SyncFullHistory()
		if err != nil {
			logrus.Panic("sync full history failed, err: ", err)
		}
	case Snapshot:

	case Light:

	}
}
