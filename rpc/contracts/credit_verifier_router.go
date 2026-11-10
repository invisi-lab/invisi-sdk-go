// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	protocolcontracts "github.com/invisi-lab/invisi-sdk-go/protocol/contracts"

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

// CreditVerifierRouterMetaData contains all meta data concerning the CreditVerifierRouter contract.
var CreditVerifierRouterMetaData = &bind.MetaData{ABI: protocolcontracts.CreditVerifierRouterMetaData.ABI}

// CreditVerifierRouterABI is the input ABI used to generate the binding from.
// Deprecated: Use CreditVerifierRouterMetaData.ABI instead.
var CreditVerifierRouterABI = CreditVerifierRouterMetaData.ABI

// CreditVerifierRouter is an auto generated Go binding around an Ethereum contract.
type CreditVerifierRouter struct {
	CreditVerifierRouterCaller     // Read-only binding to the contract
	CreditVerifierRouterTransactor // Write-only binding to the contract
	CreditVerifierRouterFilterer   // Log filterer for contract events
}

// CreditVerifierRouterCaller is an auto generated read-only Go binding around an Ethereum contract.
type CreditVerifierRouterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CreditVerifierRouterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CreditVerifierRouterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CreditVerifierRouterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CreditVerifierRouterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CreditVerifierRouterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CreditVerifierRouterSession struct {
	Contract     *CreditVerifierRouter // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// CreditVerifierRouterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CreditVerifierRouterCallerSession struct {
	Contract *CreditVerifierRouterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// CreditVerifierRouterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CreditVerifierRouterTransactorSession struct {
	Contract     *CreditVerifierRouterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// CreditVerifierRouterRaw is an auto generated low-level Go binding around an Ethereum contract.
type CreditVerifierRouterRaw struct {
	Contract *CreditVerifierRouter // Generic contract binding to access the raw methods on
}

// CreditVerifierRouterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CreditVerifierRouterCallerRaw struct {
	Contract *CreditVerifierRouterCaller // Generic read-only contract binding to access the raw methods on
}

// CreditVerifierRouterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CreditVerifierRouterTransactorRaw struct {
	Contract *CreditVerifierRouterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCreditVerifierRouter creates a new instance of CreditVerifierRouter, bound to a specific deployed contract.
func NewCreditVerifierRouter(address common.Address, backend bind.ContractBackend) (*CreditVerifierRouter, error) {
	contract, err := bindCreditVerifierRouter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CreditVerifierRouter{CreditVerifierRouterCaller: CreditVerifierRouterCaller{contract: contract}, CreditVerifierRouterTransactor: CreditVerifierRouterTransactor{contract: contract}, CreditVerifierRouterFilterer: CreditVerifierRouterFilterer{contract: contract}}, nil
}

// NewCreditVerifierRouterCaller creates a new read-only instance of CreditVerifierRouter, bound to a specific deployed contract.
func NewCreditVerifierRouterCaller(address common.Address, caller bind.ContractCaller) (*CreditVerifierRouterCaller, error) {
	contract, err := bindCreditVerifierRouter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CreditVerifierRouterCaller{contract: contract}, nil
}

// NewCreditVerifierRouterTransactor creates a new write-only instance of CreditVerifierRouter, bound to a specific deployed contract.
func NewCreditVerifierRouterTransactor(address common.Address, transactor bind.ContractTransactor) (*CreditVerifierRouterTransactor, error) {
	contract, err := bindCreditVerifierRouter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CreditVerifierRouterTransactor{contract: contract}, nil
}

// NewCreditVerifierRouterFilterer creates a new log filterer instance of CreditVerifierRouter, bound to a specific deployed contract.
func NewCreditVerifierRouterFilterer(address common.Address, filterer bind.ContractFilterer) (*CreditVerifierRouterFilterer, error) {
	contract, err := bindCreditVerifierRouter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CreditVerifierRouterFilterer{contract: contract}, nil
}

// bindCreditVerifierRouter binds a generic wrapper to an already deployed contract.
func bindCreditVerifierRouter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CreditVerifierRouterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CreditVerifierRouter *CreditVerifierRouterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CreditVerifierRouter.Contract.CreditVerifierRouterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CreditVerifierRouter *CreditVerifierRouterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CreditVerifierRouter.Contract.CreditVerifierRouterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CreditVerifierRouter *CreditVerifierRouterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CreditVerifierRouter.Contract.CreditVerifierRouterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CreditVerifierRouter *CreditVerifierRouterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CreditVerifierRouter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CreditVerifierRouter *CreditVerifierRouterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CreditVerifierRouter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CreditVerifierRouter *CreditVerifierRouterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CreditVerifierRouter.Contract.contract.Transact(opts, method, params...)
}

// CreateCreditVerifier is a free data retrieval call binding the contract method 0xf1b90628.
//
// Solidity: function createCreditVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreateCreditVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "createCreditVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreateCreditVerifier is a free data retrieval call binding the contract method 0xf1b90628.
//
// Solidity: function createCreditVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreateCreditVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreateCreditVerifier(&_CreditVerifierRouter.CallOpts)
}

// CreateCreditVerifier is a free data retrieval call binding the contract method 0xf1b90628.
//
// Solidity: function createCreditVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreateCreditVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreateCreditVerifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots10Verifier is a free data retrieval call binding the contract method 0x5d73d7f3.
//
// Solidity: function creditDrawSlots10Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots10Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots10Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots10Verifier is a free data retrieval call binding the contract method 0x5d73d7f3.
//
// Solidity: function creditDrawSlots10Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots10Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots10Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots10Verifier is a free data retrieval call binding the contract method 0x5d73d7f3.
//
// Solidity: function creditDrawSlots10Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots10Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots10Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots1Verifier is a free data retrieval call binding the contract method 0xe32cd3ea.
//
// Solidity: function creditDrawSlots1Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots1Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots1Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots1Verifier is a free data retrieval call binding the contract method 0xe32cd3ea.
//
// Solidity: function creditDrawSlots1Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots1Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots1Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots1Verifier is a free data retrieval call binding the contract method 0xe32cd3ea.
//
// Solidity: function creditDrawSlots1Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots1Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots1Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots20Verifier is a free data retrieval call binding the contract method 0xf665ed4a.
//
// Solidity: function creditDrawSlots20Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots20Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots20Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots20Verifier is a free data retrieval call binding the contract method 0xf665ed4a.
//
// Solidity: function creditDrawSlots20Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots20Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots20Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots20Verifier is a free data retrieval call binding the contract method 0xf665ed4a.
//
// Solidity: function creditDrawSlots20Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots20Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots20Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots2Verifier is a free data retrieval call binding the contract method 0xbadf1535.
//
// Solidity: function creditDrawSlots2Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots2Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots2Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots2Verifier is a free data retrieval call binding the contract method 0xbadf1535.
//
// Solidity: function creditDrawSlots2Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots2Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots2Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots2Verifier is a free data retrieval call binding the contract method 0xbadf1535.
//
// Solidity: function creditDrawSlots2Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots2Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots2Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots30Verifier is a free data retrieval call binding the contract method 0x8a350948.
//
// Solidity: function creditDrawSlots30Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots30Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots30Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots30Verifier is a free data retrieval call binding the contract method 0x8a350948.
//
// Solidity: function creditDrawSlots30Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots30Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots30Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots30Verifier is a free data retrieval call binding the contract method 0x8a350948.
//
// Solidity: function creditDrawSlots30Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots30Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots30Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots3Verifier is a free data retrieval call binding the contract method 0x4eac0686.
//
// Solidity: function creditDrawSlots3Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots3Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots3Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots3Verifier is a free data retrieval call binding the contract method 0x4eac0686.
//
// Solidity: function creditDrawSlots3Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots3Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots3Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots3Verifier is a free data retrieval call binding the contract method 0x4eac0686.
//
// Solidity: function creditDrawSlots3Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots3Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots3Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots40Verifier is a free data retrieval call binding the contract method 0x4e176696.
//
// Solidity: function creditDrawSlots40Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots40Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots40Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots40Verifier is a free data retrieval call binding the contract method 0x4e176696.
//
// Solidity: function creditDrawSlots40Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots40Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots40Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots40Verifier is a free data retrieval call binding the contract method 0x4e176696.
//
// Solidity: function creditDrawSlots40Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots40Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots40Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots4Verifier is a free data retrieval call binding the contract method 0xb12e7791.
//
// Solidity: function creditDrawSlots4Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots4Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots4Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots4Verifier is a free data retrieval call binding the contract method 0xb12e7791.
//
// Solidity: function creditDrawSlots4Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots4Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots4Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots4Verifier is a free data retrieval call binding the contract method 0xb12e7791.
//
// Solidity: function creditDrawSlots4Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots4Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots4Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots50Verifier is a free data retrieval call binding the contract method 0x6b0d9c1b.
//
// Solidity: function creditDrawSlots50Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots50Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots50Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots50Verifier is a free data retrieval call binding the contract method 0x6b0d9c1b.
//
// Solidity: function creditDrawSlots50Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots50Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots50Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots50Verifier is a free data retrieval call binding the contract method 0x6b0d9c1b.
//
// Solidity: function creditDrawSlots50Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots50Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots50Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots5Verifier is a free data retrieval call binding the contract method 0xa3c48ef1.
//
// Solidity: function creditDrawSlots5Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots5Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots5Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots5Verifier is a free data retrieval call binding the contract method 0xa3c48ef1.
//
// Solidity: function creditDrawSlots5Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots5Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots5Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots5Verifier is a free data retrieval call binding the contract method 0xa3c48ef1.
//
// Solidity: function creditDrawSlots5Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots5Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots5Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots60Verifier is a free data retrieval call binding the contract method 0x0cd316e7.
//
// Solidity: function creditDrawSlots60Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots60Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots60Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots60Verifier is a free data retrieval call binding the contract method 0x0cd316e7.
//
// Solidity: function creditDrawSlots60Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots60Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots60Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots60Verifier is a free data retrieval call binding the contract method 0x0cd316e7.
//
// Solidity: function creditDrawSlots60Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots60Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots60Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots70Verifier is a free data retrieval call binding the contract method 0xa3cf0596.
//
// Solidity: function creditDrawSlots70Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots70Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots70Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots70Verifier is a free data retrieval call binding the contract method 0xa3cf0596.
//
// Solidity: function creditDrawSlots70Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots70Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots70Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots70Verifier is a free data retrieval call binding the contract method 0xa3cf0596.
//
// Solidity: function creditDrawSlots70Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots70Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots70Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots80Verifier is a free data retrieval call binding the contract method 0x1699487e.
//
// Solidity: function creditDrawSlots80Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots80Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots80Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots80Verifier is a free data retrieval call binding the contract method 0x1699487e.
//
// Solidity: function creditDrawSlots80Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots80Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots80Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots80Verifier is a free data retrieval call binding the contract method 0x1699487e.
//
// Solidity: function creditDrawSlots80Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots80Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots80Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditUpdateVerifier is a free data retrieval call binding the contract method 0x7b8ad201.
//
// Solidity: function creditUpdateVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditUpdateVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditUpdateVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditUpdateVerifier is a free data retrieval call binding the contract method 0x7b8ad201.
//
// Solidity: function creditUpdateVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditUpdateVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditUpdateVerifier(&_CreditVerifierRouter.CallOpts)
}

// CreditUpdateVerifier is a free data retrieval call binding the contract method 0x7b8ad201.
//
// Solidity: function creditUpdateVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditUpdateVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditUpdateVerifier(&_CreditVerifierRouter.CallOpts)
}

// CrossTreeRegistry is a free data retrieval call binding the contract method 0x196b73d4.
//
// Solidity: function crossTreeRegistry() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CrossTreeRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "crossTreeRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CrossTreeRegistry is a free data retrieval call binding the contract method 0x196b73d4.
//
// Solidity: function crossTreeRegistry() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CrossTreeRegistry() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CrossTreeRegistry(&_CreditVerifierRouter.CallOpts)
}

// CrossTreeRegistry is a free data retrieval call binding the contract method 0x196b73d4.
//
// Solidity: function crossTreeRegistry() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CrossTreeRegistry() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CrossTreeRegistry(&_CreditVerifierRouter.CallOpts)
}

// GetCreateCreditVerifier is a free data retrieval call binding the contract method 0x445fc806.
//
// Solidity: function getCreateCreditVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetCreateCreditVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getCreateCreditVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreateCreditVerifier is a free data retrieval call binding the contract method 0x445fc806.
//
// Solidity: function getCreateCreditVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetCreateCreditVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreateCreditVerifier(&_CreditVerifierRouter.CallOpts)
}

// GetCreateCreditVerifier is a free data retrieval call binding the contract method 0x445fc806.
//
// Solidity: function getCreateCreditVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetCreateCreditVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreateCreditVerifier(&_CreditVerifierRouter.CallOpts)
}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 inputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetCreditDrawVerifier(opts *bind.CallOpts, inputs *big.Int) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getCreditDrawVerifier", inputs)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 inputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetCreditDrawVerifier(inputs *big.Int) (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditDrawVerifier(&_CreditVerifierRouter.CallOpts, inputs)
}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 inputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetCreditDrawVerifier(inputs *big.Int) (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditDrawVerifier(&_CreditVerifierRouter.CallOpts, inputs)
}

// GetCreditUpdateVerifier is a free data retrieval call binding the contract method 0xb6b64562.
//
// Solidity: function getCreditUpdateVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetCreditUpdateVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getCreditUpdateVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditUpdateVerifier is a free data retrieval call binding the contract method 0xb6b64562.
//
// Solidity: function getCreditUpdateVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetCreditUpdateVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditUpdateVerifier(&_CreditVerifierRouter.CallOpts)
}

// GetCreditUpdateVerifier is a free data retrieval call binding the contract method 0xb6b64562.
//
// Solidity: function getCreditUpdateVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetCreditUpdateVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditUpdateVerifier(&_CreditVerifierRouter.CallOpts)
}

// GetCrossTreeVerifier is a free data retrieval call binding the contract method 0xf63ba548.
//
// Solidity: function getCrossTreeVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetCrossTreeVerifier(opts *bind.CallOpts, inputs *big.Int, outputs *big.Int) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getCrossTreeVerifier", inputs, outputs)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCrossTreeVerifier is a free data retrieval call binding the contract method 0xf63ba548.
//
// Solidity: function getCrossTreeVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetCrossTreeVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCrossTreeVerifier(&_CreditVerifierRouter.CallOpts, inputs, outputs)
}

// GetCrossTreeVerifier is a free data retrieval call binding the contract method 0xf63ba548.
//
// Solidity: function getCrossTreeVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetCrossTreeVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCrossTreeVerifier(&_CreditVerifierRouter.CallOpts, inputs, outputs)
}

// GetPrivateTransferVerifier is a free data retrieval call binding the contract method 0x25f670ac.
//
// Solidity: function getPrivateTransferVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetPrivateTransferVerifier(opts *bind.CallOpts, inputs *big.Int, outputs *big.Int) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getPrivateTransferVerifier", inputs, outputs)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPrivateTransferVerifier is a free data retrieval call binding the contract method 0x25f670ac.
//
// Solidity: function getPrivateTransferVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetPrivateTransferVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetPrivateTransferVerifier(&_CreditVerifierRouter.CallOpts, inputs, outputs)
}

// GetPrivateTransferVerifier is a free data retrieval call binding the contract method 0x25f670ac.
//
// Solidity: function getPrivateTransferVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetPrivateTransferVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetPrivateTransferVerifier(&_CreditVerifierRouter.CallOpts, inputs, outputs)
}

// PrivateTransferRegistry is a free data retrieval call binding the contract method 0x947bbafd.
//
// Solidity: function privateTransferRegistry() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) PrivateTransferRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "privateTransferRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PrivateTransferRegistry is a free data retrieval call binding the contract method 0x947bbafd.
//
// Solidity: function privateTransferRegistry() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) PrivateTransferRegistry() (common.Address, error) {
	return _CreditVerifierRouter.Contract.PrivateTransferRegistry(&_CreditVerifierRouter.CallOpts)
}

// PrivateTransferRegistry is a free data retrieval call binding the contract method 0x947bbafd.
//
// Solidity: function privateTransferRegistry() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) PrivateTransferRegistry() (common.Address, error) {
	return _CreditVerifierRouter.Contract.PrivateTransferRegistry(&_CreditVerifierRouter.CallOpts)
}
