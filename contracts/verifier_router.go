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

// VerifierRouterMetaData contains all meta data concerning the VerifierRouter contract.
var VerifierRouterMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"batch1Verifier_\",\"type\":\"address\"},{\"name\":\"batch3Verifier_\",\"type\":\"address\"},{\"name\":\"batch10Verifier_\",\"type\":\"address\"},{\"name\":\"batch100Verifier_\",\"type\":\"address\"},{\"name\":\"shieldedPaymentVerifier_\",\"type\":\"address\"},{\"name\":\"creditDraw1Verifier_\",\"type\":\"address\"},{\"name\":\"creditDraw3Verifier_\",\"type\":\"address\"},{\"name\":\"creditDraw10Verifier_\",\"type\":\"address\"},{\"name\":\"creditDraw100Verifier_\",\"type\":\"address\"},{\"name\":\"creditReclaimVerifier_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"batch100Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"batch10Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"batch1Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"batch3Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDraw100Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDraw10Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDraw1Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDraw3Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditReclaimVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditDrawVerifier\",\"inputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"outputs\":[{\"name\":\"verifier\",\"type\":\"address\"},{\"name\":\"maxSize\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditReclaimVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getShieldedPaymentVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getWithdrawVerifier\",\"inputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"outputs\":[{\"name\":\"verifier\",\"type\":\"address\"},{\"name\":\"maxSize\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"shieldedPaymentVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"InvalidBatchCount\",\"inputs\":[{\"name\":\"count\",\"type\":\"uint256\"}]}]",
}

// VerifierRouterABI is the input ABI used to generate the binding from.
// Deprecated: Use VerifierRouterMetaData.ABI instead.
var VerifierRouterABI = VerifierRouterMetaData.ABI

// VerifierRouter is an auto generated Go binding around an Ethereum contract.
type VerifierRouter struct {
	VerifierRouterCaller     // Read-only binding to the contract
	VerifierRouterTransactor // Write-only binding to the contract
	VerifierRouterFilterer   // Log filterer for contract events
}

// VerifierRouterCaller is an auto generated read-only Go binding around an Ethereum contract.
type VerifierRouterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifierRouterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VerifierRouterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifierRouterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VerifierRouterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifierRouterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VerifierRouterSession struct {
	Contract     *VerifierRouter   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VerifierRouterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VerifierRouterCallerSession struct {
	Contract *VerifierRouterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// VerifierRouterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VerifierRouterTransactorSession struct {
	Contract     *VerifierRouterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// VerifierRouterRaw is an auto generated low-level Go binding around an Ethereum contract.
type VerifierRouterRaw struct {
	Contract *VerifierRouter // Generic contract binding to access the raw methods on
}

// VerifierRouterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VerifierRouterCallerRaw struct {
	Contract *VerifierRouterCaller // Generic read-only contract binding to access the raw methods on
}

// VerifierRouterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VerifierRouterTransactorRaw struct {
	Contract *VerifierRouterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVerifierRouter creates a new instance of VerifierRouter, bound to a specific deployed contract.
func NewVerifierRouter(address common.Address, backend bind.ContractBackend) (*VerifierRouter, error) {
	contract, err := bindVerifierRouter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &VerifierRouter{VerifierRouterCaller: VerifierRouterCaller{contract: contract}, VerifierRouterTransactor: VerifierRouterTransactor{contract: contract}, VerifierRouterFilterer: VerifierRouterFilterer{contract: contract}}, nil
}

// NewVerifierRouterCaller creates a new read-only instance of VerifierRouter, bound to a specific deployed contract.
func NewVerifierRouterCaller(address common.Address, caller bind.ContractCaller) (*VerifierRouterCaller, error) {
	contract, err := bindVerifierRouter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VerifierRouterCaller{contract: contract}, nil
}

// NewVerifierRouterTransactor creates a new write-only instance of VerifierRouter, bound to a specific deployed contract.
func NewVerifierRouterTransactor(address common.Address, transactor bind.ContractTransactor) (*VerifierRouterTransactor, error) {
	contract, err := bindVerifierRouter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VerifierRouterTransactor{contract: contract}, nil
}

// NewVerifierRouterFilterer creates a new log filterer instance of VerifierRouter, bound to a specific deployed contract.
func NewVerifierRouterFilterer(address common.Address, filterer bind.ContractFilterer) (*VerifierRouterFilterer, error) {
	contract, err := bindVerifierRouter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VerifierRouterFilterer{contract: contract}, nil
}

// bindVerifierRouter binds a generic wrapper to an already deployed contract.
func bindVerifierRouter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := VerifierRouterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VerifierRouter *VerifierRouterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VerifierRouter.Contract.VerifierRouterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VerifierRouter *VerifierRouterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRouter.Contract.VerifierRouterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VerifierRouter *VerifierRouterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VerifierRouter.Contract.VerifierRouterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VerifierRouter *VerifierRouterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VerifierRouter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VerifierRouter *VerifierRouterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRouter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VerifierRouter *VerifierRouterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VerifierRouter.Contract.contract.Transact(opts, method, params...)
}

// Batch100Verifier is a free data retrieval call binding the contract method 0xce455515.
//
// Solidity: function batch100Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) Batch100Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "batch100Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Batch100Verifier is a free data retrieval call binding the contract method 0xce455515.
//
// Solidity: function batch100Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) Batch100Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch100Verifier(&_VerifierRouter.CallOpts)
}

// Batch100Verifier is a free data retrieval call binding the contract method 0xce455515.
//
// Solidity: function batch100Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) Batch100Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch100Verifier(&_VerifierRouter.CallOpts)
}

// Batch10Verifier is a free data retrieval call binding the contract method 0x0c64ef86.
//
// Solidity: function batch10Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) Batch10Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "batch10Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Batch10Verifier is a free data retrieval call binding the contract method 0x0c64ef86.
//
// Solidity: function batch10Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) Batch10Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch10Verifier(&_VerifierRouter.CallOpts)
}

// Batch10Verifier is a free data retrieval call binding the contract method 0x0c64ef86.
//
// Solidity: function batch10Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) Batch10Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch10Verifier(&_VerifierRouter.CallOpts)
}

// Batch1Verifier is a free data retrieval call binding the contract method 0xab2df0f9.
//
// Solidity: function batch1Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) Batch1Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "batch1Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Batch1Verifier is a free data retrieval call binding the contract method 0xab2df0f9.
//
// Solidity: function batch1Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) Batch1Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch1Verifier(&_VerifierRouter.CallOpts)
}

// Batch1Verifier is a free data retrieval call binding the contract method 0xab2df0f9.
//
// Solidity: function batch1Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) Batch1Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch1Verifier(&_VerifierRouter.CallOpts)
}

// Batch3Verifier is a free data retrieval call binding the contract method 0x4c72c129.
//
// Solidity: function batch3Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) Batch3Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "batch3Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Batch3Verifier is a free data retrieval call binding the contract method 0x4c72c129.
//
// Solidity: function batch3Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) Batch3Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch3Verifier(&_VerifierRouter.CallOpts)
}

// Batch3Verifier is a free data retrieval call binding the contract method 0x4c72c129.
//
// Solidity: function batch3Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) Batch3Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.Batch3Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw100Verifier is a free data retrieval call binding the contract method 0x3c671715.
//
// Solidity: function creditDraw100Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) CreditDraw100Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "creditDraw100Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDraw100Verifier is a free data retrieval call binding the contract method 0x3c671715.
//
// Solidity: function creditDraw100Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) CreditDraw100Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw100Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw100Verifier is a free data retrieval call binding the contract method 0x3c671715.
//
// Solidity: function creditDraw100Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) CreditDraw100Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw100Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw10Verifier is a free data retrieval call binding the contract method 0xff33c99f.
//
// Solidity: function creditDraw10Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) CreditDraw10Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "creditDraw10Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDraw10Verifier is a free data retrieval call binding the contract method 0xff33c99f.
//
// Solidity: function creditDraw10Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) CreditDraw10Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw10Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw10Verifier is a free data retrieval call binding the contract method 0xff33c99f.
//
// Solidity: function creditDraw10Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) CreditDraw10Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw10Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw1Verifier is a free data retrieval call binding the contract method 0xc8d64835.
//
// Solidity: function creditDraw1Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) CreditDraw1Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "creditDraw1Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDraw1Verifier is a free data retrieval call binding the contract method 0xc8d64835.
//
// Solidity: function creditDraw1Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) CreditDraw1Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw1Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw1Verifier is a free data retrieval call binding the contract method 0xc8d64835.
//
// Solidity: function creditDraw1Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) CreditDraw1Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw1Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw3Verifier is a free data retrieval call binding the contract method 0xca6ddfe1.
//
// Solidity: function creditDraw3Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) CreditDraw3Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "creditDraw3Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDraw3Verifier is a free data retrieval call binding the contract method 0xca6ddfe1.
//
// Solidity: function creditDraw3Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) CreditDraw3Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw3Verifier(&_VerifierRouter.CallOpts)
}

// CreditDraw3Verifier is a free data retrieval call binding the contract method 0xca6ddfe1.
//
// Solidity: function creditDraw3Verifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) CreditDraw3Verifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditDraw3Verifier(&_VerifierRouter.CallOpts)
}

// CreditReclaimVerifier is a free data retrieval call binding the contract method 0x6ae56772.
//
// Solidity: function creditReclaimVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) CreditReclaimVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "creditReclaimVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditReclaimVerifier is a free data retrieval call binding the contract method 0x6ae56772.
//
// Solidity: function creditReclaimVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) CreditReclaimVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditReclaimVerifier(&_VerifierRouter.CallOpts)
}

// CreditReclaimVerifier is a free data retrieval call binding the contract method 0x6ae56772.
//
// Solidity: function creditReclaimVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) CreditReclaimVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.CreditReclaimVerifier(&_VerifierRouter.CallOpts)
}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 count) view returns(address verifier, uint256 maxSize)
func (_VerifierRouter *VerifierRouterCaller) GetCreditDrawVerifier(opts *bind.CallOpts, count *big.Int) (struct {
	Verifier common.Address
	MaxSize  *big.Int
}, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCreditDrawVerifier", count)

	outstruct := new(struct {
		Verifier common.Address
		MaxSize  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Verifier = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.MaxSize = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 count) view returns(address verifier, uint256 maxSize)
func (_VerifierRouter *VerifierRouterSession) GetCreditDrawVerifier(count *big.Int) (struct {
	Verifier common.Address
	MaxSize  *big.Int
}, error) {
	return _VerifierRouter.Contract.GetCreditDrawVerifier(&_VerifierRouter.CallOpts, count)
}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 count) view returns(address verifier, uint256 maxSize)
func (_VerifierRouter *VerifierRouterCallerSession) GetCreditDrawVerifier(count *big.Int) (struct {
	Verifier common.Address
	MaxSize  *big.Int
}, error) {
	return _VerifierRouter.Contract.GetCreditDrawVerifier(&_VerifierRouter.CallOpts, count)
}

// GetCreditReclaimVerifier is a free data retrieval call binding the contract method 0x0b0d79b3.
//
// Solidity: function getCreditReclaimVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetCreditReclaimVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCreditReclaimVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditReclaimVerifier is a free data retrieval call binding the contract method 0x0b0d79b3.
//
// Solidity: function getCreditReclaimVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetCreditReclaimVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditReclaimVerifier(&_VerifierRouter.CallOpts)
}

// GetCreditReclaimVerifier is a free data retrieval call binding the contract method 0x0b0d79b3.
//
// Solidity: function getCreditReclaimVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetCreditReclaimVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditReclaimVerifier(&_VerifierRouter.CallOpts)
}

// GetShieldedPaymentVerifier is a free data retrieval call binding the contract method 0x7749a31b.
//
// Solidity: function getShieldedPaymentVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetShieldedPaymentVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getShieldedPaymentVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetShieldedPaymentVerifier is a free data retrieval call binding the contract method 0x7749a31b.
//
// Solidity: function getShieldedPaymentVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetShieldedPaymentVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetShieldedPaymentVerifier(&_VerifierRouter.CallOpts)
}

// GetShieldedPaymentVerifier is a free data retrieval call binding the contract method 0x7749a31b.
//
// Solidity: function getShieldedPaymentVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetShieldedPaymentVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetShieldedPaymentVerifier(&_VerifierRouter.CallOpts)
}

// GetWithdrawVerifier is a free data retrieval call binding the contract method 0x1fb59810.
//
// Solidity: function getWithdrawVerifier(uint256 count) view returns(address verifier, uint256 maxSize)
func (_VerifierRouter *VerifierRouterCaller) GetWithdrawVerifier(opts *bind.CallOpts, count *big.Int) (struct {
	Verifier common.Address
	MaxSize  *big.Int
}, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getWithdrawVerifier", count)

	outstruct := new(struct {
		Verifier common.Address
		MaxSize  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Verifier = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.MaxSize = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetWithdrawVerifier is a free data retrieval call binding the contract method 0x1fb59810.
//
// Solidity: function getWithdrawVerifier(uint256 count) view returns(address verifier, uint256 maxSize)
func (_VerifierRouter *VerifierRouterSession) GetWithdrawVerifier(count *big.Int) (struct {
	Verifier common.Address
	MaxSize  *big.Int
}, error) {
	return _VerifierRouter.Contract.GetWithdrawVerifier(&_VerifierRouter.CallOpts, count)
}

// GetWithdrawVerifier is a free data retrieval call binding the contract method 0x1fb59810.
//
// Solidity: function getWithdrawVerifier(uint256 count) view returns(address verifier, uint256 maxSize)
func (_VerifierRouter *VerifierRouterCallerSession) GetWithdrawVerifier(count *big.Int) (struct {
	Verifier common.Address
	MaxSize  *big.Int
}, error) {
	return _VerifierRouter.Contract.GetWithdrawVerifier(&_VerifierRouter.CallOpts, count)
}

// ShieldedPaymentVerifier is a free data retrieval call binding the contract method 0x5a36aaf1.
//
// Solidity: function shieldedPaymentVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) ShieldedPaymentVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "shieldedPaymentVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ShieldedPaymentVerifier is a free data retrieval call binding the contract method 0x5a36aaf1.
//
// Solidity: function shieldedPaymentVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) ShieldedPaymentVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.ShieldedPaymentVerifier(&_VerifierRouter.CallOpts)
}

// ShieldedPaymentVerifier is a free data retrieval call binding the contract method 0x5a36aaf1.
//
// Solidity: function shieldedPaymentVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) ShieldedPaymentVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.ShieldedPaymentVerifier(&_VerifierRouter.CallOpts)
}
