package startup

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/config"
	"os"
)

func InitKernelConfigFromPath(cfgPath string) {
	config.LoadTomlConf(cfgPath, KernelCfg)
	initDataDir()
	initLog(KernelCfg)
}

func InitKernelConfig(cfg *config.KernelConf) {
	KernelCfg = cfg
	initDataDir()
	initLog(KernelCfg)
}

func InitDefaultKernelConfig() {
	KernelCfg = config.InitDefaultCfg()
	initDataDir()
	initLog(KernelCfg)
}

func initDataDir() {
	err := os.MkdirAll(KernelCfg.DataDir, 0700)
	if err != nil {
		panic(err)
	}
}

func initLog(cfg *config.KernelConf) {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)

	var (
		logfile *os.File
		err     error
	)

	if cfg.LogOutput == "" {
		logfile = os.Stderr
	} else {
		logfile, err = os.OpenFile(cfg.LogOutput, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			panic("init log file error: " + err.Error())
		}
	}

	logrus.SetOutput(logfile)
	lvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic("parse log level error: " + err.Error())
	}

	logrus.SetLevel(lvl)
}
