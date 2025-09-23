package testx

import (
	"github.com/yu-org/yu/apps/eth/config"
	"github.com/yu-org/yu/apps/poa"
	yuConfig "github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/startup"
)

func GenerateConfig(yuConfigPath, evmConfigPath, poaConfigPath string) (yuCfg *yuConfig.KernelConf, poaCfg *poa.PoaConfig, evmConfig *config.GethConfig) {
	yuCfg = startup.InitKernelConfigFromPath(yuConfigPath)
	evmConfig = config.LoadGethConfig(evmConfigPath)
	poaCfg = poa.LoadCfgFromPath(poaConfigPath)
	return yuCfg, poaCfg, evmConfig
}
