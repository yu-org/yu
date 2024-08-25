package startup

import (
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/utils/pprof"
)

func beforeStartUp(cfg *config.KernelConf) {
	if cfg.EnablePProf {
		pprof.StartPProf(cfg.PProfAddress)
	}
}
