package main

import (
	"github.com/Lawliet-Chan/yu/apps/asset"
	"github.com/Lawliet-Chan/yu/apps/pow"
	"github.com/Lawliet-Chan/yu/startup"
)

func main() {
	startup.StartUp(pow.NewPow(1024), asset.NewAsset("YuCoin"))
}
