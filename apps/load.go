package apps

import (
	"github.com/Lawliet-Chan/yu/apps/asset"
	"github.com/Lawliet-Chan/yu/apps/pow"
	"github.com/Lawliet-Chan/yu/tripod"
)

func LoadLand() *tripod.Land {
	land := tripod.NewLand()
	powTripod := pow.NewPow(1024)
	land.SetTripods(powTripod)

	assetTripod := asset.NewAsset("YuCoin")
	land.SetTripods(assetTripod)
	return land
}
