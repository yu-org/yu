package eth

import (
	"github.com/yu-org/yu/apps/eth/config"
	"github.com/yu-org/yu/apps/eth/ethrpc"
	"github.com/yu-org/yu/apps/eth/evm"
	"github.com/yu-org/yu/apps/poa"
	yuConfig "github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/kernel"
	"github.com/yu-org/yu/core/startup"
)

func StartupEthChain(yuCfg *yuConfig.KernelConf, poaCfg *poa.PoaConfig, gethCfg *config.GethConfig) *kernel.Kernel {
	poaTri := poa.NewPoa(poaCfg)
	solidityTri := evm.NewSolidity(gethCfg)

	chain := startup.InitDefaultKernel(yuCfg).WithTripods(poaTri, solidityTri)
	ethrpc.StartupEthRPC(chain, gethCfg)
	chain.Startup()
	return chain
}
