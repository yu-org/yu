package startup

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/config"
	"os"
)

func InitKernelConfigFromPath(cfgPath string) *config.KernelConf {
	cfg := new(config.KernelConf)
	config.LoadTomlConf(cfgPath, cfg)
	initDataDir(cfg.DataDir)
	initLog(cfg)
	return cfg
}

//func InitKernelConfig(cfg *config.KernelConf) {
//	KernelCfg = cfg
//	initDataDir()
//	initLog(KernelCfg)
//}

func InitDefaultKernelConfig() *config.KernelConf {
	cfg := config.InitDefaultCfg()
	initDataDir(cfg.DataDir)
	initLog(cfg)
	return cfg
}

func initDataDir(fpath string) {
	err := os.MkdirAll(fpath, 0700)
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
