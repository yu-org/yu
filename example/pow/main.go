package main

import (
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/pow"
	"github.com/yu-org/yu/core/startup"
)

func main() {
	startup.StartUp(pow.NewPow(1024), asset.NewAsset("YuCoin"))
}
