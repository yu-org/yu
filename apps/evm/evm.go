package evm

import (
	"encoding/binary"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// Evm struct
type Evm struct {
	cfg *runtime.Config
}

func NewEvm(sdb *state.StateDB, chainId int64, gasLimit, gasPrice uint64) *Evm {
	e := &Evm{cfg: setDefaults(chainId, gasLimit, gasPrice)}
	e.cfg.State = sdb
	return e
}

// Get evm dump
func (e *Evm) RawDump() state.Dump {
	if e.cfg.State != nil {
		return e.cfg.State.RawDump(nil)
	}
	return state.Dump{}
}

// Create a contract
func (e *Evm) Create(code []byte, origin common.Address) ([]byte, common.Address, uint64, error) {
	return runtime.Create(code, e.cfg)
}

// Call contract
func (e *Evm) Call(contAddr common.Address, origin common.Address, inputCode []byte) ([]byte, uint64, error) {
	e.cfg.State.SetCode(contAddr, e.cfg.State.GetCode(contAddr))
	return runtime.Call(contAddr, inputCode, e.cfg)
}

// Get contract bytecode
func (e *Evm) GetCode(contAddr common.Address) []byte {
	return e.cfg.State.GetCode(contAddr)
}

func (e *Evm) SetCode(contAddr common.Address, code []byte) {
	e.cfg.State.SetCode(contAddr, code)
	return
}

// Prepare hash into evm
func (e *Evm) Prepare(txhash, blhash common.Hash, txindex int) {
	e.cfg.State.Prepare(txhash, txindex)
}

// Add log
func (e *Evm) AddLog(log *types.Log) {
	e.cfg.State.AddLog(log)
}

// Get logs
func (e *Evm) GetLogs(txHash, blockH common.Hash) []*types.Log {
	log := e.cfg.State.GetLogs(txHash, blockH)
	return log
}

// Get logs
func (e *Evm) Logs() []*types.Log {
	return e.cfg.State.Logs()
}

// SetBlockInfo set block info into evm
func (e *Evm) SetBlockInfo(num, tm uint64, miner common.Address, difficulty *big.Int) {
	e.cfg.BlockNumber = new(big.Int).SetUint64(num)
	e.cfg.Coinbase = miner
	e.cfg.Time = new(big.Int).SetUint64(tm)
	e.cfg.Difficulty = difficulty
}

// Set evm cfg
func (e *Evm) SetConfig(val, price *big.Int, limit uint64, origin common.Address) {
	e.cfg.Value = val
	e.cfg.GasPrice = price
	e.cfg.GasLimit = limit
	e.cfg.Origin = origin
}

// Get evm Config
func (e *Evm) GetConfig() *runtime.Config {
	return e.cfg
}

// Add Balance
func (e *Evm) AddBalance(addr common.Address, amount *big.Int) {
	e.cfg.State.AddBalance(addr, amount)
}

// Sub Balance
func (e *Evm) SubBalance(addr common.Address, amount *big.Int) {
	e.cfg.State.SubBalance(addr, amount)
}

// Set Balance
func (e *Evm) SetBalance(addr common.Address, amount *big.Int) {
	e.cfg.State.SetBalance(addr, amount)
}

// Get Balance
func (e *Evm) GetBalance(addr common.Address) *big.Int {
	bi := e.cfg.State.GetBalance(addr)
	return bi

}

// Get Nonce
func (e *Evm) GetNonce(addr common.Address) uint64 {
	non := e.cfg.State.GetNonce(addr)
	return non
}

// Set Nonce
func (e *Evm) SetNonce(addr common.Address, nonce uint64) {
	e.cfg.State.SetNonce(addr, nonce)
}

// Get Storage At address
func (e *Evm) GetStorageAt(addr common.Address, hash common.Hash) common.Hash {
	proof := e.cfg.State.GetState(addr, hash)
	return proof
}

// Get Snapshot
func (e *Evm) GetSnapshot() int {
	i := e.cfg.State.Snapshot()
	return i
}

// Revert Snapshot to a position
func (e *Evm) RevertToSnapshot(sp int) {
	e.cfg.State.RevertToSnapshot(sp)
}

// sets defaults config
func setDefaults(chainId int64, gasLimit, gasPrice uint64) *runtime.Config {
	cfg := new(runtime.Config)
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = &params.ChainConfig{
			ChainID:             big.NewInt(chainId),
			HomesteadBlock:      new(big.Int),
			DAOForkBlock:        new(big.Int),
			DAOForkSupport:      false,
			EIP150Block:         new(big.Int),
			EIP150Hash:          common.Hash{},
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
			MuirGlacierBlock:    new(big.Int),
			//YoloV3Block:         nil,
		}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.Time == nil {
		cfg.Time = big.NewInt(time.Now().Unix())
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = gasLimit
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = big.NewInt(int64(gasPrice))
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
	return cfg
}

// parse token name or symbol by call result.
func ParseCallResultToString(result string) string {
	var res string
	resBt := common.Hex2Bytes(result)
	if len(resBt) > 64 {
		l := binary.BigEndian.Uint32(resBt[60:64])
		res = string(resBt[64 : 64+l])
	}
	return res
}

// parse token decimal or totalSupply by call result.
func ParseCallResultToBig(result string) *big.Int {
	if res, ok := big.NewInt(0).SetString(result, 16); ok {
		return res
	}
	return nil
}
