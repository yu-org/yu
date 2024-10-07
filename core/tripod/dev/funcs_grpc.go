package dev

import (
	"context"
	"errors"
	. "github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/core/types/goproto"
	"google.golang.org/grpc"
)

type GrpcWrRd struct {
	targetAddr string
	tripodName string
	funcName   string
}

func NewGrpcWrRd(targetAddr, tripodName, funcName string) *GrpcWrRd {
	return &GrpcWrRd{
		targetAddr: targetAddr,
		tripodName: tripodName,
		funcName:   funcName,
	}
}

func (rpc *GrpcWrRd) Write(ctx *WriteContext) error {
	conn, err := grpc.NewClient(rpc.targetAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	cli := goproto.NewWritingClient(conn)
	res, err := cli.Write(context.Background(), &goproto.WriteContext{
		ReadContext: &goproto.ReadContext{
			ParamsStr:  ctx.ParamsStr,
			Response:   nil,
			TripodName: rpc.tripodName,
			FuncName:   rpc.funcName,
		},
		Block:   ctx.Block.ToPb(),
		Txn:     ctx.Txn.ToPb(),
		LeiCost: ctx.LeiCost,
	})
	if err != nil {
		return err
	}
	if res.Error != nil {
		return errors.New(res.Error.GetMsg())
	}

	events := make([]*types.Event, 0)
	for _, value := range res.Values {
		events = append(events, &types.Event{Value: value})
	}
	ctx.Events = events
	return nil
}

func (rpc *GrpcWrRd) Read(ctx *ReadContext) error {
	conn, err := grpc.NewClient(rpc.targetAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	cli := goproto.NewReadingClient(conn)
	res, err := cli.Read(context.Background(), &goproto.ReadContext{
		ParamsStr:  ctx.GetParams(),
		Response:   nil,
		TripodName: rpc.tripodName,
		FuncName:   rpc.funcName,
	})
	if err != nil {
		return err
	}
	if res.Error != nil {
		return errors.New(res.Error.GetMsg())
	}

	ctx.Data(int(res.GetCode()), res.GetContentType(), res.GetData())
	return nil
}
