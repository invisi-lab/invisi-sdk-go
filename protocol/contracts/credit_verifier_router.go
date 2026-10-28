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

// CreditVerifierRouterMetaData contains all meta data concerning the CreditVerifierRouter contract.
var CreditVerifierRouterMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"singleRegistry\",\"type\":\"address\",\"internalType\":\"contractIPrivateTransferVerifierRegistry\"},{\"name\":\"crossRegistry\",\"type\":\"address\",\"internalType\":\"contractCrossTreeVerifierRegistry\"},{\"name\":\"creditVerifiers\",\"type\":\"address[7]\",\"internalType\":\"contractIVerifier[7]\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createCreditVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditCancelVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDrawSlots120Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDrawSlots12Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDrawSlots1Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDrawSlots50Verifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditTopUpVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"crossTreeRegistry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractCrossTreeVerifierRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreateCreditVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditCancelVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditDrawVerifier\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCreditTopUpVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCrossTreeVerifier\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPrivateTransferVerifier\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"privateTransferRegistry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPrivateTransferVerifierRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"InvalidBatchSize\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UnsupportedCreditCreation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UnsupportedCrossTreeShape\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"UnsupportedPrivateTransferShape\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
}

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

// CreditCancelVerifier is a free data retrieval call binding the contract method 0x967c7d10.
//
// Solidity: function creditCancelVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditCancelVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditCancelVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditCancelVerifier is a free data retrieval call binding the contract method 0x967c7d10.
//
// Solidity: function creditCancelVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditCancelVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditCancelVerifier(&_CreditVerifierRouter.CallOpts)
}

// CreditCancelVerifier is a free data retrieval call binding the contract method 0x967c7d10.
//
// Solidity: function creditCancelVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditCancelVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditCancelVerifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots120Verifier is a free data retrieval call binding the contract method 0x7c7f67e4.
//
// Solidity: function creditDrawSlots120Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots120Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots120Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots120Verifier is a free data retrieval call binding the contract method 0x7c7f67e4.
//
// Solidity: function creditDrawSlots120Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots120Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots120Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots120Verifier is a free data retrieval call binding the contract method 0x7c7f67e4.
//
// Solidity: function creditDrawSlots120Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots120Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots120Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots12Verifier is a free data retrieval call binding the contract method 0x9c58a03f.
//
// Solidity: function creditDrawSlots12Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditDrawSlots12Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditDrawSlots12Verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditDrawSlots12Verifier is a free data retrieval call binding the contract method 0x9c58a03f.
//
// Solidity: function creditDrawSlots12Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditDrawSlots12Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots12Verifier(&_CreditVerifierRouter.CallOpts)
}

// CreditDrawSlots12Verifier is a free data retrieval call binding the contract method 0x9c58a03f.
//
// Solidity: function creditDrawSlots12Verifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditDrawSlots12Verifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditDrawSlots12Verifier(&_CreditVerifierRouter.CallOpts)
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

// CreditTopUpVerifier is a free data retrieval call binding the contract method 0x1e99874f.
//
// Solidity: function creditTopUpVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) CreditTopUpVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "creditTopUpVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CreditTopUpVerifier is a free data retrieval call binding the contract method 0x1e99874f.
//
// Solidity: function creditTopUpVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) CreditTopUpVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditTopUpVerifier(&_CreditVerifierRouter.CallOpts)
}

// CreditTopUpVerifier is a free data retrieval call binding the contract method 0x1e99874f.
//
// Solidity: function creditTopUpVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) CreditTopUpVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.CreditTopUpVerifier(&_CreditVerifierRouter.CallOpts)
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

// GetCreditCancelVerifier is a free data retrieval call binding the contract method 0x7207e214.
//
// Solidity: function getCreditCancelVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetCreditCancelVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getCreditCancelVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditCancelVerifier is a free data retrieval call binding the contract method 0x7207e214.
//
// Solidity: function getCreditCancelVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetCreditCancelVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditCancelVerifier(&_CreditVerifierRouter.CallOpts)
}

// GetCreditCancelVerifier is a free data retrieval call binding the contract method 0x7207e214.
//
// Solidity: function getCreditCancelVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetCreditCancelVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditCancelVerifier(&_CreditVerifierRouter.CallOpts)
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

// GetCreditTopUpVerifier is a free data retrieval call binding the contract method 0x604ee023.
//
// Solidity: function getCreditTopUpVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCaller) GetCreditTopUpVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CreditVerifierRouter.contract.Call(opts, &out, "getCreditTopUpVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCreditTopUpVerifier is a free data retrieval call binding the contract method 0x604ee023.
//
// Solidity: function getCreditTopUpVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterSession) GetCreditTopUpVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditTopUpVerifier(&_CreditVerifierRouter.CallOpts)
}

// GetCreditTopUpVerifier is a free data retrieval call binding the contract method 0x604ee023.
//
// Solidity: function getCreditTopUpVerifier() view returns(address)
func (_CreditVerifierRouter *CreditVerifierRouterCallerSession) GetCreditTopUpVerifier() (common.Address, error) {
	return _CreditVerifierRouter.Contract.GetCreditTopUpVerifier(&_CreditVerifierRouter.CallOpts)
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
