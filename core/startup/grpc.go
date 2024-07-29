package startup

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/types/goproto"
	"google.golang.org/grpc"
	"net"
)

func StartGrpcServer(cfg *config.KernelConf) {
	if cfg.RunMode != common.MasterWorker {
		return
	}
	lis, err := net.Listen("tcp", cfg.GrpcPort)
	if err != nil {
		logrus.Fatal("listen for grpc failed: ", err)
	}
	grpcServer := grpc.NewServer()
	goproto.RegisterStateDBServer(grpcServer, state.NewGrpcMptKV(StateDB))
	goproto.RegisterLandServer(grpcServer, tripod.NewGrpcLand(Land))
	// TODO: add chain server, pool server, txndb server.

	err = grpcServer.Serve(lis)
	if err != nil {
		logrus.Fatal("failed to serve grpc: ", err)
	}
}
