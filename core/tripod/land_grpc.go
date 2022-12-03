package tripod

import (
	"context"
	"github.com/yu-org/yu/core/tripod/dev"
	"github.com/yu-org/yu/core/types/goproto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcLand struct {
	land *Land
}

func NewGrpcLand(land *Land) *GrpcLand {
	return &GrpcLand{land}
}

func (g *GrpcLand) SetTripods(_ context.Context, info *goproto.TripodsInfo) (*emptypb.Empty, error) {
	tripods := make([]*Tripod, 0)
	for _, triInfo := range info.Tripods {
		tripod := NewTripod(triInfo.Name)
		for _, wrName := range triInfo.Writings {
			wrRd := dev.NewGrpcWrRd(triInfo.Endpoint, triInfo.Name, wrName)
			tripod.writings[wrName] = wrRd.Write
		}
		for _, rdName := range triInfo.Readings {
			wrRd := dev.NewGrpcWrRd(triInfo.Endpoint, triInfo.Name, rdName)
			tripod.readings[rdName] = wrRd.Read
		}
		// todo: set p2pHandles

		tripods = append(tripods)
	}
	g.land.SetTripods(tripods...)
	return nil, nil
}
