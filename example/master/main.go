package main

import (
	"github.com/yu-altar/yu/apps/asset"
	"github.com/yu-altar/yu/apps/pow"
	"github.com/yu-altar/yu/startup"
)

func main() {
	startup.StartUp(pow.NewPow(1024), asset.NewAsset("YuCoin"))
}
