package startup

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/base"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/chain_env"
	"github.com/yu-org/yu/core/kernel"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txdb"
	"github.com/yu-org/yu/core/txpool"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/utils/codec"
	"os"
)

var (
	kernelCfgPath string
	kernelCfg     config.KernelConf
	isInitedLog   bool
)

func StartUpFullNode(tripodInstances ...interface{}) {
	tripodInstances = append([]interface{}{base.NewBase(base.Full)}, tripodInstances...)
	StartUp(tripodInstances...)
}

func StartUp(tripodInstances ...interface{}) {
	tripods := make([]*tripod.Tripod, 0)
	for _, v := range tripodInstances {
		tripods = append(tripods, tripod.ResolveTripod(v))
	}

	initCfgFromFlags()
	if !isInitedLog {
		InitLog(kernelCfg.LogLevel, kernelCfg.LogOutput)
	}

	codec.GlobalCodec = &codec.RlpCodec{}
	gin.SetMode(gin.ReleaseMode)

	land := tripod.NewLand()

	kvdb, err := kv.NewKvdb(&kernelCfg.KVDB)
	if err != nil {
		logrus.Fatal("init kvdb error: ", err)
	}

	base := txdb.NewTxDB(kvdb)
	chain := blockchain.NewBlockChain(&kernelCfg.BlockChain, base)
	statedb := state.NewStateDB(kvdb)
	pool := txpool.WithDefaultChecks(&kernelCfg.Txpool, base)

	for _, tri := range tripods {
		pool.WithTripodCheck(tri)
	}

	env := &chain_env.ChainEnv{
		State:      statedb,
		Chain:      chain,
		TxDB:       base,
		Pool:       pool,
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: p2p.NewP2P(&kernelCfg.P2P),
	}

	for i, t := range tripods {
		t.SetChainEnv(env)
		t.SetLand(land)
		t.SetInstance(tripodInstances[i])
	}

	land.SetTripods(tripods...)

	for _, tripodInterface := range tripodInstances {
		err = tripod.Inject(tripodInterface)
		if err != nil {
			logrus.Fatal("inject tripod failed: ", err)
		}
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

	config.LoadTomlConf(kernelCfgPath, &kernelCfg)
}

func InitLog(level, output string) {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)

	var (
		logfile *os.File
		err     error
	)

	if output == "" {
		logfile = os.Stderr
	} else {
		logfile, err = os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			panic("init log file error: " + err.Error())
		}
	}

	logrus.SetOutput(logfile)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		panic("parse log level error: " + err.Error())
	}

	logrus.SetLevel(lvl)
	isInitedLog = true
}
