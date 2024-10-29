package tripod

import (
	"context"
	"github.com/yu-org/yu/core/types/goproto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcLandServer struct {
	land *Land
}

func NewGrpcLand(land *Land) *GrpcLandServer {
	return &GrpcLandServer{land}
}

func (g *GrpcLandServer) SetTripods(_ context.Context, info *goproto.TripodsInfo) (*emptypb.Empty, error) {
	tripods := make([]*Tripod, 0)

	for _, triInfo := range info.Tripods {
		conn, err := grpc.NewClient(triInfo.Endpoint)
		if err != nil {
			return nil, err
		}
		tripod := NewTripodWithName(triInfo.Name).WithGrpcConn(conn)
		////set writings
		//for _, wrName := range triInfo.Writings {
		//	wrRd := dev.NewGrpcWrRd(triInfo.Endpoint, triInfo.Name, wrName)
		//	tripod.writings[wrName] = wrRd.Write
		//}
		////set readings
		//for _, rdName := range triInfo.Readings {
		//	wrRd := dev.NewGrpcWrRd(triInfo.Endpoint, triInfo.Name, rdName)
		//	tripod.readings[rdName] = wrRd.Read
		//}
		// todo: set p2pHandles

		tripods = append(tripods, tripod)
	}
	g.land.SetTripods(tripods...)
	return nil, nil
}
