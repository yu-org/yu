package startup

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/chain_env"
	"github.com/yu-org/yu/core/kernel"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txpool"
	"github.com/yu-org/yu/core/yudb"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/utils/codec"
)

var (
	kernelCfgPath string
	kernelCfg     config.KernelConf
)

func StartUp(tripods ...tripod.Tripod) {
	initCfgFromFlags()
	initLog(kernelCfg.LogLevel)

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.ReleaseMode)

	land := tripod.NewLand()
	land.SetTripods(tripods...)

	chain := blockchain.NewBlockChain(&kernelCfg.BlockChain)

	base := yudb.NewYuDB(&kernelCfg.BlockBase)

	statedb := state.NewStateDB(&kernelCfg.State)

	env := &chain_env.ChainEnv{
		State:      statedb,
		Chain:      chain,
		YuDB:       base,
		Pool:       txpool.LocalWithDefaultChecks(&kernelCfg.Txpool, base),
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: p2p.NewP2P(&kernelCfg.P2P),
	}

	for _, t := range tripods {
		t.SetChainEnv(env)
	}

	k := kernel.NewKernel(&kernelCfg, env, land)

	k.Startup()
}

func initCfgFromFlags() {
	useDefaultCfg := flag.Bool("dc", false, "default config files")

	flag.StringVar(&kernelCfgPath, "k", "yu_conf/kernel.toml", "Kernel config file path")

	flag.Parse()
	if *useDefaultCfg {
		kernelCfg = config.InitDefaultCfg()
		return
	}

	config.LoadConf(kernelCfgPath, &kernelCfg)
}

func initLog(level string) {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		panic("parse log level error: " + err.Error())
	}

	logrus.SetLevel(lvl)
}
