package evm

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/types"
	"math/big"
)

func NewDefaultEVM(header *types.Header, chain *blockchain.BlockChain, statedb vm.StateDB) *vm.EVM {
	ctx := NewEVMBlockContext(header, chain, nil)
	return vm.NewEVM(ctx, vm.TxContext{}, statedb, DefaultEthChainCfg, vm.Config{})
}

func NewEVM(
	header *types.Header,
	chain *blockchain.BlockChain,
	statedb vm.StateDB,
	chainCfg *params.ChainConfig,
	txCtx vm.TxContext,
	cfg vm.Config,
	author common.Address,
) *vm.EVM {
	gaddr := gcommon.Address(author)
	ctx := NewEVMBlockContext(header, chain, &gaddr)
	return vm.NewEVM(ctx, txCtx, statedb, chainCfg, cfg)
}

var onFork = big.NewInt(0)

var DefaultEthChainCfg = &params.ChainConfig{
	ChainID:             big.NewInt(1),
	HomesteadBlock:      onFork,
	DAOForkBlock:        onFork,
	DAOForkSupport:      true,
	EIP150Block:         onFork,
	EIP150Hash:          gcommon.Hash{},
	EIP155Block:         onFork,
	EIP158Block:         onFork,
	ByzantiumBlock:      onFork,
	ConstantinopleBlock: onFork,
	PetersburgBlock:     onFork,
	IstanbulBlock:       onFork,
	MuirGlacierBlock:    onFork,
	BerlinBlock:         onFork,
	LondonBlock:         onFork,
	CatalystBlock:       onFork,
	Ethash:              nil,
	Clique:              nil,
}
