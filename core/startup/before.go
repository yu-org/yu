package startup

import (
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/utils/banner"
	"github.com/yu-org/yu/utils/ip"
	"github.com/yu-org/yu/utils/pprof"
)

func beforeStartUp(cfg *config.KernelConf) {
	// show what this chain is, before anything else runs.
	banner.PrintChainSpec(&cfg.ChainSpec)

	if cfg.EnablePProf {
		pprof.StartPProf(ip.MakePort(cfg.PProfPort))
	}
}
