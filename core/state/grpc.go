package state

import (
	"context"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types/goproto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcMptKV struct {
	kv *MptKV
}

func NewGrpcMptKV(kv *MptKV) *GrpcMptKV {
	return &GrpcMptKV{kv}
}

func (g *GrpcMptKV) Get(ctx context.Context, key *goproto.Key) (*goproto.ValueResponse, error) {
	value, err := g.kv.get(key.GetTripodName(), key.GetKey())
	if err != nil {
		return nil, err
	}
	return &goproto.ValueResponse{Value: value}, nil
}

func (g *GrpcMptKV) Set(ctx context.Context, keyValue *goproto.KeyValue) (*emptypb.Empty, error) {
	g.kv.set(keyValue.GetTripodName(), keyValue.GetKey(), keyValue.GetValue())
	return nil, nil
}

func (g *GrpcMptKV) Delete(ctx context.Context, key *goproto.Key) (*emptypb.Empty, error) {
	g.kv.delete(key.GetTripodName(), key.GetKey())
	return nil, nil
}

func (g *GrpcMptKV) Exist(ctx context.Context, key *goproto.Key) (*goproto.Bool, error) {
	ok := g.kv.exist(key.GetTripodName(), key.GetKey())
	return &goproto.Bool{Ok: ok}, nil
}

func (g *GrpcMptKV) GetByBlockHash(ctx context.Context, hash *goproto.KeyByHash) (*goproto.ValueResponse, error) {
	value, err := g.kv.getByBlockHash(hash.GetTripodName(), hash.GetKey(), common.BytesToHash(hash.GetBlockHash()))
	if err != nil {
		return nil, err
	}
	return &goproto.ValueResponse{Value: value}, nil
}

func (g *GrpcMptKV) GetFinalized(ctx context.Context, key *goproto.Key) (*goproto.ValueResponse, error) {
	value, err := g.kv.getFinalized(key.GetTripodName(), key.GetKey())
	if err != nil {
		return nil, err
	}
	return &goproto.ValueResponse{Value: value}, nil
}

func (g *GrpcMptKV) StartBlock(ctx context.Context, hash *goproto.TxnHash) (*emptypb.Empty, error) {
	g.kv.StartBlock(common.BytesToHash(hash.Hash))
	return nil, nil
}

func (g *GrpcMptKV) Commit(ctx context.Context, empty *emptypb.Empty) (*goproto.TxnHashResponse, error) {
	hash, err := g.kv.Commit()
	if err != nil {
		return nil, err
	}
	return &goproto.TxnHashResponse{Hash: hash.Bytes()}, nil
}

func (g *GrpcMptKV) Discard(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	g.kv.Discard()
	return nil, nil
}

func (g *GrpcMptKV) DiscardAll(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	g.kv.DiscardAll()
	return nil, nil
}

func (g *GrpcMptKV) NextTxn(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	g.kv.NextTxn()
	return nil, nil
}
