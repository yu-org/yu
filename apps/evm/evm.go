package evm

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	gcore "github.com/ethereum/go-ethereum/core"
	gstate "github.com/ethereum/go-ethereum/core/state"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/types"
	"math/big"
)

func ApplyTxn(
	block *types.Block,
	chain *blockchain.BlockChain,
	statedb *gstate.StateDB,
	to common.Address,
	value, gasPrice,
	gasFeeCap, gasTipCap uint64,
) error {
	gaspool := new(gcore.GasPool).AddGas(block.LeiLimit)
	gasprice := new(big.Int).SetUint64(gasPrice)
	amount := new(big.Int).SetUint64(value)
	gfc := new(big.Int).SetUint64(gasFeeCap)
	gtc := new(big.Int).SetUint64(gasTipCap)
	toAddr := gcommon.Address(to)

	for i, stxn := range block.Txns {
		msg := gtypes.NewMessage(
			gcommon.Address(stxn.Raw.Caller),
			&toAddr, stxn.Raw.Nonce,
			amount, block.LeiLimit,
			gasprice, gfc,
			gtc, stxn.Raw.Code,
			nil, false,
		)

		evm := NewDefaultEVM(block.Header, chain, statedb)
		txCtx := vm.TxContext{
			Origin:   gcommon.Address(stxn.Raw.Caller),
			GasPrice: gasprice,
		}
		evm.Reset(txCtx, statedb)

		statedb.Prepare(gcommon.Hash(stxn.TxnHash), i)
		result, err := gcore.ApplyMessage(evm, msg, gaspool)
		if err != nil {
			return err
		}

		block.LeiUsed += result.UsedGas
	}
	return nil
}

func NewDefaultEVM(header *types.Header, chain *blockchain.BlockChain, statedb vm.StateDB) *vm.EVM {
	blockCtx := NewEVMBlockContext(header, chain, nil)
	return vm.NewEVM(blockCtx, vm.TxContext{}, statedb, DefaultEthChainCfg, vm.Config{})
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
