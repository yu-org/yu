package main

import (
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/startup"
)

func main() {
	startup.StartUp(config.InitDefaultCfg())
}
