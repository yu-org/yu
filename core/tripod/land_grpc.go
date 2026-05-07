package tripod

type GrpcLand struct {
	land *Land
}

func NewGrpcLand(land *Land) *GrpcLand {
	return &GrpcLand{land}
}

//func (g *GrpcLand) RegisterTripods(_ context.Context, info *goproto.TripodsInfo) (*emptypb.Empty, error) {
//	tripods := make([]*Tripod, 0)
//	for _, triInfo := range info.Tripods {
//		tripod := NewTripodWithName(triInfo.Name)
//		for _, wrName := range triInfo.Writings {
//			wrRd := dev.NewGrpcWrRd(triInfo.Endpoint, triInfo.Name, wrName)
//			tripod.writings[wrName] = wrRd.Write
//		}
//		//for _, rdName := range triInfo.Readings {
//		//	wrRd := dev.NewGrpcWrRd(triInfo.Endpoint, triInfo.Name, rdName)
//		//	tripod.readings[rdName] = wrRd.Read
//		//}
//		// todo: set p2pHandles
//
//		tripods = append(tripods)
//	}
//	g.land.RegisterTripods(tripods...)
//	return nil, nil
//}
