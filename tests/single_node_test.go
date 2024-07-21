package tests

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/startup"
	"github.com/yu-org/yu/core/types"
	cliAsset "github.com/yu-org/yu/example/client/asset"
	"github.com/yu-org/yu/example/client/callchain"
	"sync"
	"testing"
	"time"
)

func TestSingleNode(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go runChain(&wg)
	transferAsset(t)
	wg.Wait()
}

func runChain(wg *sync.WaitGroup) {
	poaCfg := poa.DefaultCfg(0)
	startup.InitDefaultKernelConfig()

	assetTri := asset.NewAsset("yu-coin")
	poaTri := poa.NewPoa(poaCfg)

	chain := startup.InitDefaultKernel(poaTri, assetTri)
	go chain.Startup()

	blockInterval := time.Duration(poaCfg.BlockInterval) * time.Second
	time.Sleep(blockInterval * 16)

	chain.Stop()

	wg.Done()
}

func transferAsset(t *testing.T) {
	pubkey, privkey, err := keypair.GenKeyPair(keypair.Sr25519)
	if err != nil {
		panic("generate key error: " + err.Error())
	}

	toPubkey, _, err := keypair.GenKeyPair(keypair.Sr25519)
	if err != nil {
		panic("generate To Address key error: " + err.Error())
	}

	sub, err := callchain.NewSubscriber()
	if err != nil {
		panic("new subscriber failed: " + err.Error())
	}

	resultCh := make(chan *types.Receipt)
	go sub.SubEvent(resultCh)

	var (
		createAmount uint64 = 500
		transfer1    uint64 = 50
		transfer2    uint64 = 100
	)

	logrus.Info("--- send Creating Account ---")
	cliAsset.CreateAccount(privkey, pubkey, createAmount)
	time.Sleep(4 * time.Second)

	logrus.Info("--- send Transferring 1 ---")
	cliAsset.TransferBalance(privkey, pubkey, toPubkey.Address(), transfer1, 0)
	time.Sleep(4 * time.Second)

	logrus.Info("--- send Transferring 2 ---")
	cliAsset.TransferBalance(privkey, pubkey, toPubkey.Address(), transfer2, 0)
	time.Sleep(6 * time.Second)

	balance1 := cliAsset.QueryAccount(pubkey)
	assert.Equal(t, createAmount-transfer1, balance1)
	balance2 := cliAsset.QueryAccount(toPubkey)
	assert.Equal(t, createAmount-transfer1-transfer2, balance2)
}
