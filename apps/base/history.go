package base

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
)

const (
	Full = iota
	Snapshot
	Light
)

type Base struct {
	*Tripod
	mode int
}

func NewBase(mode int) *Base {
	tri := NewTripod("base")
	fh := &Base{Tripod: tri, mode: mode}
	tri.SetInit(fh)
	tri.SetP2pHandler(HandshakeCode, fh.handleHsReq).SetP2pHandler(SyncTxnsCode, fh.handleSyncTxnsReq)
	return fh
}

func (b *Base) InitChain() {
	b.defineGenesis()
	b.syncHistory()
}

func (b *Base) defineGenesis() {
	rootPubkey, rootPrivkey := GenSrKeyWithSecret([]byte("root"))
	genesisHash := HexToHash("genesis")
	signer, err := rootPrivkey.SignData(genesisHash.Bytes())
	if err != nil {
		logrus.Panic("sign genesis block failed: ", err)
	}

	gensisBlock := &CompactBlock{
		Header: &Header{
			Hash:           genesisHash,
			MinerPubkey:    rootPubkey.BytesWithType(),
			MinerSignature: signer,
		},
	}

	err = b.Chain.SetGenesis(gensisBlock)
	if err != nil {
		logrus.Panic("set genesis block failed: ", err)
	}
	err = b.Chain.Finalize(genesisHash)
	if err != nil {
		logrus.Panic("finalize genesis block failed: ", err)
	}
}

func (b *Base) syncHistory() {
	if len(b.P2pNetwork.GetBootNodes()) == 0 {
		return
	}
	switch b.mode {
	case Full:
		err := b.syncFullHistory()
		if err != nil {
			logrus.Panic("sync full history failed, err: ", err)
		}
	case Snapshot:

	case Light:

	}
}
