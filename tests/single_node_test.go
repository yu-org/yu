package tests

import (
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/core/startup"
	"testing"
	"time"
)

func TestSingleNode(t *testing.T) {
	go runChain()

}

func runChain() {
	poaCfg := poa.DefaultCfg(0)
	startup.InitDefaultKernelConfig()

	assetTri := asset.NewAsset("yu-coin")
	poaTri := poa.NewPoa(poaCfg)

	chain := startup.InitDefaultKernel(poaTri, assetTri)
	go chain.Startup()

	blockInterval := time.Duration(poaCfg.BlockInterval) * time.Second
	time.Sleep(blockInterval * 10)

	chain.Stop()
}

func transferAsset() {

}
