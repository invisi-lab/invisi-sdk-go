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
	ABI: "[{\"type\":\"function\",\"name\":\"getCreateCreditVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditCancelVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditDrawVerifier\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditTopUpVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCrossTreeVerifier\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPrivateTransferVerifier\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"UnsupportedCreditCreation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UnsupportedCrossTreeShape\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"UnsupportedPrivateTransferShape\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
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

// GetCreateCreditVerifier is a free data retrieval call binding the contract method 0x445fc806.
//
// Solidity: function getCreateCreditVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetCreateCreditVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCreateCreditVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreateCreditVerifier is a free data retrieval call binding the contract method 0x445fc806.
//
// Solidity: function getCreateCreditVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetCreateCreditVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreateCreditVerifier(&_VerifierRouter.CallOpts)
}

// GetCreateCreditVerifier is a free data retrieval call binding the contract method 0x445fc806.
//
// Solidity: function getCreateCreditVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetCreateCreditVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreateCreditVerifier(&_VerifierRouter.CallOpts)
}

// GetCreditCancelVerifier is a free data retrieval call binding the contract method 0x7207e214.
//
// Solidity: function getCreditCancelVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetCreditCancelVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCreditCancelVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditCancelVerifier is a free data retrieval call binding the contract method 0x7207e214.
//
// Solidity: function getCreditCancelVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetCreditCancelVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditCancelVerifier(&_VerifierRouter.CallOpts)
}

// GetCreditCancelVerifier is a free data retrieval call binding the contract method 0x7207e214.
//
// Solidity: function getCreditCancelVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetCreditCancelVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditCancelVerifier(&_VerifierRouter.CallOpts)
}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 ) view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetCreditDrawVerifier(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCreditDrawVerifier", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 ) view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetCreditDrawVerifier(arg0 *big.Int) (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditDrawVerifier(&_VerifierRouter.CallOpts, arg0)
}

// GetCreditDrawVerifier is a free data retrieval call binding the contract method 0x86b5ceca.
//
// Solidity: function getCreditDrawVerifier(uint256 ) view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetCreditDrawVerifier(arg0 *big.Int) (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditDrawVerifier(&_VerifierRouter.CallOpts, arg0)
}

// GetCreditTopUpVerifier is a free data retrieval call binding the contract method 0x604ee023.
//
// Solidity: function getCreditTopUpVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetCreditTopUpVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCreditTopUpVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditTopUpVerifier is a free data retrieval call binding the contract method 0x604ee023.
//
// Solidity: function getCreditTopUpVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetCreditTopUpVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditTopUpVerifier(&_VerifierRouter.CallOpts)
}

// GetCreditTopUpVerifier is a free data retrieval call binding the contract method 0x604ee023.
//
// Solidity: function getCreditTopUpVerifier() view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetCreditTopUpVerifier() (common.Address, error) {
	return _VerifierRouter.Contract.GetCreditTopUpVerifier(&_VerifierRouter.CallOpts)
}

// GetCrossTreeVerifier is a free data retrieval call binding the contract method 0xf63ba548.
//
// Solidity: function getCrossTreeVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetCrossTreeVerifier(opts *bind.CallOpts, inputs *big.Int, outputs *big.Int) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getCrossTreeVerifier", inputs, outputs)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCrossTreeVerifier is a free data retrieval call binding the contract method 0xf63ba548.
//
// Solidity: function getCrossTreeVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetCrossTreeVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _VerifierRouter.Contract.GetCrossTreeVerifier(&_VerifierRouter.CallOpts, inputs, outputs)
}

// GetCrossTreeVerifier is a free data retrieval call binding the contract method 0xf63ba548.
//
// Solidity: function getCrossTreeVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetCrossTreeVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _VerifierRouter.Contract.GetCrossTreeVerifier(&_VerifierRouter.CallOpts, inputs, outputs)
}

// GetPrivateTransferVerifier is a free data retrieval call binding the contract method 0x25f670ac.
//
// Solidity: function getPrivateTransferVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_VerifierRouter *VerifierRouterCaller) GetPrivateTransferVerifier(opts *bind.CallOpts, inputs *big.Int, outputs *big.Int) (common.Address, error) {
	var out []interface{}
	err := _VerifierRouter.contract.Call(opts, &out, "getPrivateTransferVerifier", inputs, outputs)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPrivateTransferVerifier is a free data retrieval call binding the contract method 0x25f670ac.
//
// Solidity: function getPrivateTransferVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_VerifierRouter *VerifierRouterSession) GetPrivateTransferVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _VerifierRouter.Contract.GetPrivateTransferVerifier(&_VerifierRouter.CallOpts, inputs, outputs)
}

// GetPrivateTransferVerifier is a free data retrieval call binding the contract method 0x25f670ac.
//
// Solidity: function getPrivateTransferVerifier(uint256 inputs, uint256 outputs) view returns(address)
func (_VerifierRouter *VerifierRouterCallerSession) GetPrivateTransferVerifier(inputs *big.Int, outputs *big.Int) (common.Address, error) {
	return _VerifierRouter.Contract.GetPrivateTransferVerifier(&_VerifierRouter.CallOpts, inputs, outputs)
}
