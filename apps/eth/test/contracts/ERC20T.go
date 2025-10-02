// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ERC20TMetaData contains all meta data concerning the ERC20T contract.
var ERC20TMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"blockNumberT\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526012600560006101000a81548160ff021916908360ff1602179055503480156200002d57600080fd5b506040516200137038038062001370833981810160405281019062000053919062000212565b8160039081620000649190620004e2565b508060049081620000769190620004e2565b505050620005c9565b6000604051905090565b600080fd5b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b620000e8826200009d565b810181811067ffffffffffffffff821117156200010a5762000109620000ae565b5b80604052505050565b60006200011f6200007f565b90506200012d8282620000dd565b919050565b600067ffffffffffffffff82111562000150576200014f620000ae565b5b6200015b826200009d565b9050602081019050919050565b60005b83811015620001885780820151818401526020810190506200016b565b60008484015250505050565b6000620001ab620001a58462000132565b62000113565b905082815260208101848484011115620001ca57620001c962000098565b5b620001d784828562000168565b509392505050565b600082601f830112620001f757620001f662000093565b5b81516200020984826020860162000194565b91505092915050565b600080604083850312156200022c576200022b62000089565b5b600083015167ffffffffffffffff8111156200024d576200024c6200008e565b5b6200025b85828601620001df565b925050602083015167ffffffffffffffff8111156200027f576200027e6200008e565b5b6200028d85828601620001df565b9150509250929050565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b60006002820490506001821680620002ea57607f821691505b6020821081036200030057620002ff620002a2565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b6000600883026200036a7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826200032b565b6200037686836200032b565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b6000620003c3620003bd620003b7846200038e565b62000398565b6200038e565b9050919050565b6000819050919050565b620003df83620003a2565b620003f7620003ee82620003ca565b84845462000338565b825550505050565b600090565b6200040e620003ff565b6200041b818484620003d4565b505050565b5b8181101562000443576200043760008262000404565b60018101905062000421565b5050565b601f82111562000492576200045c8162000306565b62000467846200031b565b8101602085101562000477578190505b6200048f62000486856200031b565b83018262000420565b50505b505050565b600082821c905092915050565b6000620004b76000198460080262000497565b1980831691505092915050565b6000620004d28383620004a4565b9150826002028217905092915050565b620004ed8262000297565b67ffffffffffffffff811115620005095762000508620000ae565b5b620005158254620002d1565b6200052282828562000447565b600060209050601f8311600181146200055a576000841562000545578287015190505b620005518582620004c4565b865550620005c1565b601f1984166200056a8662000306565b60005b8281101562000594578489015182556001820191506020850194506020810190506200056d565b86831015620005b45784890151620005b0601f891682620004a4565b8355505b6001600288020188555050505b505050505050565b610d9780620005d96000396000f3fe608060405234801561001057600080fd5b50600436106100b45760003560e01c806370a082311161007157806370a082311461018f578063785c15ed146101bf57806395d89b41146101dd578063a0712d68146101fb578063a9059cbb14610217578063dd62ed3e14610247576100b4565b806306fdde03146100b9578063095ea7b3146100d757806318160ddd1461010757806323b872dd14610125578063313ce5671461015557806342966c6814610173575b600080fd5b6100c1610277565b6040516100ce91906109eb565b60405180910390f35b6100f160048036038101906100ec9190610aa6565b610305565b6040516100fe9190610b01565b60405180910390f35b61010f6103f7565b60405161011c9190610b2b565b60405180910390f35b61013f600480360381019061013a9190610b46565b6103fd565b60405161014c9190610b01565b60405180910390f35b61015d6105ac565b60405161016a9190610bb5565b60405180910390f35b61018d60048036038101906101889190610bd0565b6105bf565b005b6101a960048036038101906101a49190610bfd565b610696565b6040516101b69190610b2b565b60405180910390f35b6101c76106ae565b6040516101d49190610b2b565b60405180910390f35b6101e56106b6565b6040516101f291906109eb565b60405180910390f35b61021560048036038101906102109190610bd0565b610744565b005b610231600480360381019061022c9190610aa6565b61081b565b60405161023e9190610b01565b60405180910390f35b610261600480360381019061025c9190610c2a565b610936565b60405161026e9190610b2b565b60405180910390f35b6003805461028490610c99565b80601f01602080910402602001604051908101604052809291908181526020018280546102b090610c99565b80156102fd5780601f106102d2576101008083540402835291602001916102fd565b820191906000526020600020905b8154815290600101906020018083116102e057829003601f168201915b505050505081565b600081600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040516103e59190610b2b565b60405180910390a36001905092915050565b60025481565b600081600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825461048b9190610cf9565b92505081905550816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546104e09190610cf9565b92505081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546105359190610d2d565b925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040516105999190610b2b565b60405180910390a3600190509392505050565b600560009054906101000a900460ff1681565b806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825461060d9190610cf9565b9250508190555080600260008282546106269190610cf9565b92505081905550600073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8360405161068b9190610b2b565b60405180910390a350565b60006020528060005260406000206000915090505481565b600043905090565b600480546106c390610c99565b80601f01602080910402602001604051908101604052809291908181526020018280546106ef90610c99565b801561073c5780601f106107115761010080835404028352916020019161073c565b820191906000526020600020905b81548152906001019060200180831161071f57829003601f168201915b505050505081565b806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546107929190610d2d565b9250508190555080600260008282546107ab9190610d2d565b925050819055503373ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040516108109190610b2b565b60405180910390a350565b6000816000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825461086b9190610cf9565b92505081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546108c09190610d2d565b925050819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040516109249190610b2b565b60405180910390a36001905092915050565b6001602052816000526040600020602052806000526040600020600091509150505481565b600081519050919050565b600082825260208201905092915050565b60005b8381101561099557808201518184015260208101905061097a565b60008484015250505050565b6000601f19601f8301169050919050565b60006109bd8261095b565b6109c78185610966565b93506109d7818560208601610977565b6109e0816109a1565b840191505092915050565b60006020820190508181036000830152610a0581846109b2565b905092915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610a3d82610a12565b9050919050565b610a4d81610a32565b8114610a5857600080fd5b50565b600081359050610a6a81610a44565b92915050565b6000819050919050565b610a8381610a70565b8114610a8e57600080fd5b50565b600081359050610aa081610a7a565b92915050565b60008060408385031215610abd57610abc610a0d565b5b6000610acb85828601610a5b565b9250506020610adc85828601610a91565b9150509250929050565b60008115159050919050565b610afb81610ae6565b82525050565b6000602082019050610b166000830184610af2565b92915050565b610b2581610a70565b82525050565b6000602082019050610b406000830184610b1c565b92915050565b600080600060608486031215610b5f57610b5e610a0d565b5b6000610b6d86828701610a5b565b9350506020610b7e86828701610a5b565b9250506040610b8f86828701610a91565b9150509250925092565b600060ff82169050919050565b610baf81610b99565b82525050565b6000602082019050610bca6000830184610ba6565b92915050565b600060208284031215610be657610be5610a0d565b5b6000610bf484828501610a91565b91505092915050565b600060208284031215610c1357610c12610a0d565b5b6000610c2184828501610a5b565b91505092915050565b60008060408385031215610c4157610c40610a0d565b5b6000610c4f85828601610a5b565b9250506020610c6085828601610a5b565b9150509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b60006002820490506001821680610cb157607f821691505b602082108103610cc457610cc3610c6a565b5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610d0482610a70565b9150610d0f83610a70565b9250828203905081811115610d2757610d26610cca565b5b92915050565b6000610d3882610a70565b9150610d4383610a70565b9250828201905080821115610d5b57610d5a610cca565b5b9291505056fea2646970667358221220d45fbc2d549fa0f927666815f229ca5a68fac394bbbf583c0e1f4f372e5c344c64736f6c63430008150033",
}

// ERC20TABI is the input ABI used to generate the binding from.
// Deprecated: Use ERC20TMetaData.ABI instead.
var ERC20TABI = ERC20TMetaData.ABI

// ERC20TBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ERC20TMetaData.Bin instead.
var ERC20TBin = ERC20TMetaData.Bin

// DeployERC20T deploys a new Ethereum contract, binding an instance of ERC20T to it.
func DeployERC20T(auth *bind.TransactOpts, backend bind.ContractBackend, name_ string, symbol_ string) (common.Address, *types.Transaction, *ERC20T, error) {
	parsed, err := ERC20TMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ERC20TBin), backend, name_, symbol_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ERC20T{ERC20TCaller: ERC20TCaller{contract: contract}, ERC20TTransactor: ERC20TTransactor{contract: contract}, ERC20TFilterer: ERC20TFilterer{contract: contract}}, nil
}

// ERC20T is an auto generated Go binding around an Ethereum contract.
type ERC20T struct {
	ERC20TCaller     // Read-only binding to the contract
	ERC20TTransactor // Write-only binding to the contract
	ERC20TFilterer   // Log filterer for contract events
}

// ERC20TCaller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20TCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20TTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20TTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20TFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20TFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20TSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20TSession struct {
	Contract     *ERC20T           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20TCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20TCallerSession struct {
	Contract *ERC20TCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ERC20TTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20TTransactorSession struct {
	Contract     *ERC20TTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20TRaw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20TRaw struct {
	Contract *ERC20T // Generic contract binding to access the raw methods on
}

// ERC20TCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20TCallerRaw struct {
	Contract *ERC20TCaller // Generic read-only contract binding to access the raw methods on
}

// ERC20TTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20TTransactorRaw struct {
	Contract *ERC20TTransactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20T creates a new instance of ERC20T, bound to a specific deployed contract.
func NewERC20T(address common.Address, backend bind.ContractBackend) (*ERC20T, error) {
	contract, err := bindERC20T(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20T{ERC20TCaller: ERC20TCaller{contract: contract}, ERC20TTransactor: ERC20TTransactor{contract: contract}, ERC20TFilterer: ERC20TFilterer{contract: contract}}, nil
}

// NewERC20TCaller creates a new read-only instance of ERC20T, bound to a specific deployed contract.
func NewERC20TCaller(address common.Address, caller bind.ContractCaller) (*ERC20TCaller, error) {
	contract, err := bindERC20T(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20TCaller{contract: contract}, nil
}

// NewERC20TTransactor creates a new write-only instance of ERC20T, bound to a specific deployed contract.
func NewERC20TTransactor(address common.Address, transactor bind.ContractTransactor) (*ERC20TTransactor, error) {
	contract, err := bindERC20T(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20TTransactor{contract: contract}, nil
}

// NewERC20TFilterer creates a new log filterer instance of ERC20T, bound to a specific deployed contract.
func NewERC20TFilterer(address common.Address, filterer bind.ContractFilterer) (*ERC20TFilterer, error) {
	contract, err := bindERC20T(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20TFilterer{contract: contract}, nil
}

// bindERC20T binds a generic wrapper to an already deployed contract.
func bindERC20T(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ERC20TMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20T *ERC20TRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20T.Contract.ERC20TCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20T *ERC20TRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20T.Contract.ERC20TTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20T *ERC20TRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20T.Contract.ERC20TTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20T *ERC20TCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20T.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20T *ERC20TTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20T.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20T *ERC20TTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20T.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_ERC20T *ERC20TCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_ERC20T *ERC20TSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _ERC20T.Contract.Allowance(&_ERC20T.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_ERC20T *ERC20TCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _ERC20T.Contract.Allowance(&_ERC20T.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_ERC20T *ERC20TCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_ERC20T *ERC20TSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _ERC20T.Contract.BalanceOf(&_ERC20T.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_ERC20T *ERC20TCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _ERC20T.Contract.BalanceOf(&_ERC20T.CallOpts, arg0)
}

// BlockNumberT is a free data retrieval call binding the contract method 0x785c15ed.
//
// Solidity: function blockNumberT() view returns(uint256 blockNumber)
func (_ERC20T *ERC20TCaller) BlockNumberT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "blockNumberT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BlockNumberT is a free data retrieval call binding the contract method 0x785c15ed.
//
// Solidity: function blockNumberT() view returns(uint256 blockNumber)
func (_ERC20T *ERC20TSession) BlockNumberT() (*big.Int, error) {
	return _ERC20T.Contract.BlockNumberT(&_ERC20T.CallOpts)
}

// BlockNumberT is a free data retrieval call binding the contract method 0x785c15ed.
//
// Solidity: function blockNumberT() view returns(uint256 blockNumber)
func (_ERC20T *ERC20TCallerSession) BlockNumberT() (*big.Int, error) {
	return _ERC20T.Contract.BlockNumberT(&_ERC20T.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ERC20T *ERC20TCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ERC20T *ERC20TSession) Decimals() (uint8, error) {
	return _ERC20T.Contract.Decimals(&_ERC20T.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ERC20T *ERC20TCallerSession) Decimals() (uint8, error) {
	return _ERC20T.Contract.Decimals(&_ERC20T.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ERC20T *ERC20TCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ERC20T *ERC20TSession) Name() (string, error) {
	return _ERC20T.Contract.Name(&_ERC20T.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ERC20T *ERC20TCallerSession) Name() (string, error) {
	return _ERC20T.Contract.Name(&_ERC20T.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ERC20T *ERC20TCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ERC20T *ERC20TSession) Symbol() (string, error) {
	return _ERC20T.Contract.Symbol(&_ERC20T.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ERC20T *ERC20TCallerSession) Symbol() (string, error) {
	return _ERC20T.Contract.Symbol(&_ERC20T.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20T *ERC20TCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ERC20T.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20T *ERC20TSession) TotalSupply() (*big.Int, error) {
	return _ERC20T.Contract.TotalSupply(&_ERC20T.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20T *ERC20TCallerSession) TotalSupply() (*big.Int, error) {
	return _ERC20T.Contract.TotalSupply(&_ERC20T.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20T *ERC20TTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20T *ERC20TSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Approve(&_ERC20T.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20T *ERC20TTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Approve(&_ERC20T.TransactOpts, spender, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 amount) returns()
func (_ERC20T *ERC20TTransactor) Burn(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.contract.Transact(opts, "burn", amount)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 amount) returns()
func (_ERC20T *ERC20TSession) Burn(amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Burn(&_ERC20T.TransactOpts, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 amount) returns()
func (_ERC20T *ERC20TTransactorSession) Burn(amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Burn(&_ERC20T.TransactOpts, amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) returns()
func (_ERC20T *ERC20TTransactor) Mint(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.contract.Transact(opts, "mint", amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) returns()
func (_ERC20T *ERC20TSession) Mint(amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Mint(&_ERC20T.TransactOpts, amount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount) returns()
func (_ERC20T *ERC20TTransactorSession) Mint(amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Mint(&_ERC20T.TransactOpts, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20T *ERC20TTransactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20T *ERC20TSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Transfer(&_ERC20T.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20T *ERC20TTransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.Transfer(&_ERC20T.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20T *ERC20TTransactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20T *ERC20TSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.TransferFrom(&_ERC20T.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20T *ERC20TTransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20T.Contract.TransferFrom(&_ERC20T.TransactOpts, sender, recipient, amount)
}

// ERC20TApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ERC20T contract.
type ERC20TApprovalIterator struct {
	Event *ERC20TApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20TApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20TApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20TApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20TApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20TApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20TApproval represents a Approval event raised by the ERC20T contract.
type ERC20TApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20T *ERC20TFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ERC20TApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20T.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ERC20TApprovalIterator{contract: _ERC20T.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20T *ERC20TFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ERC20TApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20T.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20TApproval)
				if err := _ERC20T.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20T *ERC20TFilterer) ParseApproval(log types.Log) (*ERC20TApproval, error) {
	event := new(ERC20TApproval)
	if err := _ERC20T.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20TTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ERC20T contract.
type ERC20TTransferIterator struct {
	Event *ERC20TTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ERC20TTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20TTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ERC20TTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ERC20TTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20TTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20TTransfer represents a Transfer event raised by the ERC20T contract.
type ERC20TTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20T *ERC20TFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ERC20TTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20T.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ERC20TTransferIterator{contract: _ERC20T.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20T *ERC20TFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ERC20TTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20T.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20TTransfer)
				if err := _ERC20T.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20T *ERC20TFilterer) ParseTransfer(log types.Log) (*ERC20TTransfer, error) {
	event := new(ERC20TTransfer)
	if err := _ERC20T.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
