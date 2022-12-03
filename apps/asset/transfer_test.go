package asset

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/infra/storage/kv"
	"github.com/yu-org/yu/utils/codec"
	"math/big"
	"testing"
)

var (
	Aaddr common.Address
	Baddr common.Address

	initAamount = big.NewInt(500)
	initBamount = big.NewInt(500)
)

func init() {
	aPubkey, _ := keypair.GenSrKeyWithSecret([]byte("a"))
	bPubkey, _ := keypair.GenSrKeyWithSecret([]byte("b"))
	Aaddr = aPubkey.Address()
	println("A addr is ", Aaddr.String())
	Baddr = bPubkey.Address()
	println("B addr is ", Baddr.String())
}

type TransferInfo struct {
	To     string `json:"to"`
	Amount uint64 `json:"amount"`
}

func TestTransfer(t *testing.T) {
	asset := newAsset(t)
	byt, _ := json.Marshal(TransferInfo{
		To:     Baddr.String(),
		Amount: 200,
	})
	ctx, err := context.NewWriteContext(Aaddr, string(byt), nil)
	if err != nil {
		panic(err)
	}
	err = asset.Transfer(ctx)
	if err != nil {
		panic(err)
	}

	abalance := asset.GetBalance(Aaddr)
	bbalance := asset.GetBalance(Baddr)

	assert.Equal(t, big.NewInt(300), abalance)
	assert.Equal(t, big.NewInt(700), bbalance)
}

type QryAccount struct {
	Account string `json:"account"`
}

func TestQueryBalance(t *testing.T) {
	asset := newAsset(t)
	byt, _ := json.Marshal(QryAccount{Aaddr.String()})
	ctx, err := context.NewWriteContext(Aaddr, string(byt), nil)
	if err != nil {
		panic(err)
	}
	Aamount, err := asset.QueryBalance(ctx)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, initAamount, Aamount)

	byt, _ = json.Marshal(QryAccount{Baddr.String()})
	ctx, err = context.NewWriteContext(Baddr, string(byt), nil)
	if err != nil {
		panic(err)
	}
	Bamount, err := asset.QueryBalance(ctx)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, initBamount, Bamount)
}

func newAsset(t *testing.T) *Asset {
	asset := NewAsset("test_asset")
	cfg := config.InitDefaultCfg()
	kvdb, err := kv.NewKvdb(&cfg.KVDB)
	assert.NoError(t, err)
	statedb := state.NewStateDB(kvdb)
	codec.GlobalCodec = &codec.RlpCodec{}
	env := &env.ChainEnv{State: statedb}
	asset.SetChainEnv(env)
	asset.SetBalance(Aaddr, initAamount)
	asset.SetBalance(Baddr, initBamount)
	return asset
}
