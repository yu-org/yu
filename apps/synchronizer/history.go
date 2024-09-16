package synchronizer

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
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

func (b *Synchronizer) InitChain(block *Block) {
	b.defineGenesis(block)
	b.syncHistory()
}

func (b *Synchronizer) defineGenesis(genesisBlock *Block) {
	// FIXME: must NOT generate private key onchain.
	rootPubkey, rootPrivkey := GenSrKeyWithSecret([]byte("root"))
	genesisHash := HexToHash("genesis")
	signer, err := rootPrivkey.SignData(genesisHash.Bytes())
	if err != nil {
		logrus.Panic("sign genesis block failed: ", err)
	}
	genesisBlock.Hash = genesisHash
	genesisBlock.MinerSignature = signer
	genesisBlock.MinerPubkey = rootPubkey.BytesWithType()

	err = b.Chain.SetGenesis(genesisBlock)
	if err != nil {
		logrus.Panic("set genesis block failed: ", err)
	}
	err = b.Chain.Finalize(genesisBlock)
	if err != nil {
		logrus.Panic("finalize genesis block failed: ", err)
	}
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
