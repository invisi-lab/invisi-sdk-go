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

// InvisiFactoryMetaData contains all meta data concerning the InvisiFactory contract.
var InvisiFactoryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"globalEmergency_\",\"type\":\"address\"},{\"name\":\"verifierRouter_\",\"type\":\"address\"},{\"name\":\"creationCodeHash_\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"computeInstanceAddress\",\"inputs\":[{\"name\":\"creationCode\",\"type\":\"bytes\"},{\"name\":\"operator\",\"type\":\"address\"},{\"name\":\"token_\",\"type\":\"address\"},{\"name\":\"emergencyAdmin_\",\"type\":\"address\"},{\"name\":\"timelockAdmin_\",\"type\":\"address\"},{\"name\":\"maxPerDeposit_\",\"type\":\"uint256\"},{\"name\":\"screener_\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"instance\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creationCodeHash\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deploy\",\"inputs\":[{\"name\":\"creationCode\",\"type\":\"bytes\"},{\"name\":\"token_\",\"type\":\"address\"},{\"name\":\"emergencyAdmin_\",\"type\":\"address\"},{\"name\":\"timelockAdmin_\",\"type\":\"address\"},{\"name\":\"maxPerDeposit_\",\"type\":\"uint256\"},{\"name\":\"screener_\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"instance\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"emergencyPauseInstance\",\"inputs\":[{\"name\":\"instance\",\"type\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"emergencyPauseRange\",\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\"},{\"name\":\"end\",\"type\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getInstanceAddress\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\"},{\"name\":\"token\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"instance\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalEmergency\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"instanceCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isInstance\",\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifierRouter\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"EmergencyPauseInstance\",\"inputs\":[{\"name\":\"instance\",\"type\":\"address\",\"indexed\":true}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EmergencyPauseRange\",\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"end\",\"type\":\"uint256\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InstanceDeployed\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true},{\"name\":\"instance\",\"type\":\"address\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InstanceAlreadyDeployed\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\"},{\"name\":\"token\",\"type\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidCreationCode\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAnInstance\",\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotGlobalEmergency\",\"inputs\":[]}]",
}

// InvisiFactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use InvisiFactoryMetaData.ABI instead.
var InvisiFactoryABI = InvisiFactoryMetaData.ABI

// InvisiFactory is an auto generated Go binding around an Ethereum contract.
type InvisiFactory struct {
	InvisiFactoryCaller     // Read-only binding to the contract
	InvisiFactoryTransactor // Write-only binding to the contract
	InvisiFactoryFilterer   // Log filterer for contract events
}

// InvisiFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type InvisiFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InvisiFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InvisiFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InvisiFactorySession struct {
	Contract     *InvisiFactory    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// InvisiFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InvisiFactoryCallerSession struct {
	Contract *InvisiFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// InvisiFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InvisiFactoryTransactorSession struct {
	Contract     *InvisiFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// InvisiFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type InvisiFactoryRaw struct {
	Contract *InvisiFactory // Generic contract binding to access the raw methods on
}

// InvisiFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InvisiFactoryCallerRaw struct {
	Contract *InvisiFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// InvisiFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InvisiFactoryTransactorRaw struct {
	Contract *InvisiFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInvisiFactory creates a new instance of InvisiFactory, bound to a specific deployed contract.
func NewInvisiFactory(address common.Address, backend bind.ContractBackend) (*InvisiFactory, error) {
	contract, err := bindInvisiFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InvisiFactory{InvisiFactoryCaller: InvisiFactoryCaller{contract: contract}, InvisiFactoryTransactor: InvisiFactoryTransactor{contract: contract}, InvisiFactoryFilterer: InvisiFactoryFilterer{contract: contract}}, nil
}

// NewInvisiFactoryCaller creates a new read-only instance of InvisiFactory, bound to a specific deployed contract.
func NewInvisiFactoryCaller(address common.Address, caller bind.ContractCaller) (*InvisiFactoryCaller, error) {
	contract, err := bindInvisiFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryCaller{contract: contract}, nil
}

// NewInvisiFactoryTransactor creates a new write-only instance of InvisiFactory, bound to a specific deployed contract.
func NewInvisiFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*InvisiFactoryTransactor, error) {
	contract, err := bindInvisiFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryTransactor{contract: contract}, nil
}

// NewInvisiFactoryFilterer creates a new log filterer instance of InvisiFactory, bound to a specific deployed contract.
func NewInvisiFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*InvisiFactoryFilterer, error) {
	contract, err := bindInvisiFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryFilterer{contract: contract}, nil
}

// bindInvisiFactory binds a generic wrapper to an already deployed contract.
func bindInvisiFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InvisiFactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiFactory *InvisiFactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiFactory.Contract.InvisiFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiFactory *InvisiFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiFactory.Contract.InvisiFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiFactory *InvisiFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiFactory.Contract.InvisiFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiFactory *InvisiFactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiFactory *InvisiFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiFactory *InvisiFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiFactory.Contract.contract.Transact(opts, method, params...)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiFactory *InvisiFactoryCaller) VERSION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiFactory *InvisiFactorySession) VERSION() (*big.Int, error) {
	return _InvisiFactory.Contract.VERSION(&_InvisiFactory.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiFactory *InvisiFactoryCallerSession) VERSION() (*big.Int, error) {
	return _InvisiFactory.Contract.VERSION(&_InvisiFactory.CallOpts)
}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x997ad877.
//
// Solidity: function computeInstanceAddress(bytes creationCode, address operator, address token_, address emergencyAdmin_, address timelockAdmin_, uint256 maxPerDeposit_, address screener_) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCaller) ComputeInstanceAddress(opts *bind.CallOpts, creationCode []byte, operator common.Address, token_ common.Address, emergencyAdmin_ common.Address, timelockAdmin_ common.Address, maxPerDeposit_ *big.Int, screener_ common.Address) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "computeInstanceAddress", creationCode, operator, token_, emergencyAdmin_, timelockAdmin_, maxPerDeposit_, screener_)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x997ad877.
//
// Solidity: function computeInstanceAddress(bytes creationCode, address operator, address token_, address emergencyAdmin_, address timelockAdmin_, uint256 maxPerDeposit_, address screener_) view returns(address instance)
func (_InvisiFactory *InvisiFactorySession) ComputeInstanceAddress(creationCode []byte, operator common.Address, token_ common.Address, emergencyAdmin_ common.Address, timelockAdmin_ common.Address, maxPerDeposit_ *big.Int, screener_ common.Address) (common.Address, error) {
	return _InvisiFactory.Contract.ComputeInstanceAddress(&_InvisiFactory.CallOpts, creationCode, operator, token_, emergencyAdmin_, timelockAdmin_, maxPerDeposit_, screener_)
}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x997ad877.
//
// Solidity: function computeInstanceAddress(bytes creationCode, address operator, address token_, address emergencyAdmin_, address timelockAdmin_, uint256 maxPerDeposit_, address screener_) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCallerSession) ComputeInstanceAddress(creationCode []byte, operator common.Address, token_ common.Address, emergencyAdmin_ common.Address, timelockAdmin_ common.Address, maxPerDeposit_ *big.Int, screener_ common.Address) (common.Address, error) {
	return _InvisiFactory.Contract.ComputeInstanceAddress(&_InvisiFactory.CallOpts, creationCode, operator, token_, emergencyAdmin_, timelockAdmin_, maxPerDeposit_, screener_)
}

// CreationCodeHash is a free data retrieval call binding the contract method 0x47191163.
//
// Solidity: function creationCodeHash() view returns(bytes32)
func (_InvisiFactory *InvisiFactoryCaller) CreationCodeHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "creationCodeHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CreationCodeHash is a free data retrieval call binding the contract method 0x47191163.
//
// Solidity: function creationCodeHash() view returns(bytes32)
func (_InvisiFactory *InvisiFactorySession) CreationCodeHash() ([32]byte, error) {
	return _InvisiFactory.Contract.CreationCodeHash(&_InvisiFactory.CallOpts)
}

// CreationCodeHash is a free data retrieval call binding the contract method 0x47191163.
//
// Solidity: function creationCodeHash() view returns(bytes32)
func (_InvisiFactory *InvisiFactoryCallerSession) CreationCodeHash() ([32]byte, error) {
	return _InvisiFactory.Contract.CreationCodeHash(&_InvisiFactory.CallOpts)
}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xdfb17f48.
//
// Solidity: function getInstanceAddress(address operator, address token) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCaller) GetInstanceAddress(opts *bind.CallOpts, operator common.Address, token common.Address) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "getInstanceAddress", operator, token)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xdfb17f48.
//
// Solidity: function getInstanceAddress(address operator, address token) view returns(address instance)
func (_InvisiFactory *InvisiFactorySession) GetInstanceAddress(operator common.Address, token common.Address) (common.Address, error) {
	return _InvisiFactory.Contract.GetInstanceAddress(&_InvisiFactory.CallOpts, operator, token)
}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xdfb17f48.
//
// Solidity: function getInstanceAddress(address operator, address token) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCallerSession) GetInstanceAddress(operator common.Address, token common.Address) (common.Address, error) {
	return _InvisiFactory.Contract.GetInstanceAddress(&_InvisiFactory.CallOpts, operator, token)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiFactory *InvisiFactoryCaller) GlobalEmergency(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "globalEmergency")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiFactory *InvisiFactorySession) GlobalEmergency() (common.Address, error) {
	return _InvisiFactory.Contract.GlobalEmergency(&_InvisiFactory.CallOpts)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiFactory *InvisiFactoryCallerSession) GlobalEmergency() (common.Address, error) {
	return _InvisiFactory.Contract.GlobalEmergency(&_InvisiFactory.CallOpts)
}

// InstanceCount is a free data retrieval call binding the contract method 0xec56c716.
//
// Solidity: function instanceCount() view returns(uint256)
func (_InvisiFactory *InvisiFactoryCaller) InstanceCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "instanceCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InstanceCount is a free data retrieval call binding the contract method 0xec56c716.
//
// Solidity: function instanceCount() view returns(uint256)
func (_InvisiFactory *InvisiFactorySession) InstanceCount() (*big.Int, error) {
	return _InvisiFactory.Contract.InstanceCount(&_InvisiFactory.CallOpts)
}

// InstanceCount is a free data retrieval call binding the contract method 0xec56c716.
//
// Solidity: function instanceCount() view returns(uint256)
func (_InvisiFactory *InvisiFactoryCallerSession) InstanceCount() (*big.Int, error) {
	return _InvisiFactory.Contract.InstanceCount(&_InvisiFactory.CallOpts)
}

// IsInstance is a free data retrieval call binding the contract method 0x6b44e6be.
//
// Solidity: function isInstance(address ) view returns(bool)
func (_InvisiFactory *InvisiFactoryCaller) IsInstance(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "isInstance", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInstance is a free data retrieval call binding the contract method 0x6b44e6be.
//
// Solidity: function isInstance(address ) view returns(bool)
func (_InvisiFactory *InvisiFactorySession) IsInstance(arg0 common.Address) (bool, error) {
	return _InvisiFactory.Contract.IsInstance(&_InvisiFactory.CallOpts, arg0)
}

// IsInstance is a free data retrieval call binding the contract method 0x6b44e6be.
//
// Solidity: function isInstance(address ) view returns(bool)
func (_InvisiFactory *InvisiFactoryCallerSession) IsInstance(arg0 common.Address) (bool, error) {
	return _InvisiFactory.Contract.IsInstance(&_InvisiFactory.CallOpts, arg0)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiFactory *InvisiFactoryCaller) VerifierRouter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "verifierRouter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiFactory *InvisiFactorySession) VerifierRouter() (common.Address, error) {
	return _InvisiFactory.Contract.VerifierRouter(&_InvisiFactory.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiFactory *InvisiFactoryCallerSession) VerifierRouter() (common.Address, error) {
	return _InvisiFactory.Contract.VerifierRouter(&_InvisiFactory.CallOpts)
}

// Deploy is a paid mutator transaction binding the contract method 0x261f19b6.
//
// Solidity: function deploy(bytes creationCode, address token_, address emergencyAdmin_, address timelockAdmin_, uint256 maxPerDeposit_, address screener_) returns(address instance)
func (_InvisiFactory *InvisiFactoryTransactor) Deploy(opts *bind.TransactOpts, creationCode []byte, token_ common.Address, emergencyAdmin_ common.Address, timelockAdmin_ common.Address, maxPerDeposit_ *big.Int, screener_ common.Address) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "deploy", creationCode, token_, emergencyAdmin_, timelockAdmin_, maxPerDeposit_, screener_)
}

// Deploy is a paid mutator transaction binding the contract method 0x261f19b6.
//
// Solidity: function deploy(bytes creationCode, address token_, address emergencyAdmin_, address timelockAdmin_, uint256 maxPerDeposit_, address screener_) returns(address instance)
func (_InvisiFactory *InvisiFactorySession) Deploy(creationCode []byte, token_ common.Address, emergencyAdmin_ common.Address, timelockAdmin_ common.Address, maxPerDeposit_ *big.Int, screener_ common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.Deploy(&_InvisiFactory.TransactOpts, creationCode, token_, emergencyAdmin_, timelockAdmin_, maxPerDeposit_, screener_)
}

// Deploy is a paid mutator transaction binding the contract method 0x261f19b6.
//
// Solidity: function deploy(bytes creationCode, address token_, address emergencyAdmin_, address timelockAdmin_, uint256 maxPerDeposit_, address screener_) returns(address instance)
func (_InvisiFactory *InvisiFactoryTransactorSession) Deploy(creationCode []byte, token_ common.Address, emergencyAdmin_ common.Address, timelockAdmin_ common.Address, maxPerDeposit_ *big.Int, screener_ common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.Deploy(&_InvisiFactory.TransactOpts, creationCode, token_, emergencyAdmin_, timelockAdmin_, maxPerDeposit_, screener_)
}

// EmergencyPauseInstance is a paid mutator transaction binding the contract method 0xf2d5de51.
//
// Solidity: function emergencyPauseInstance(address instance) returns()
func (_InvisiFactory *InvisiFactoryTransactor) EmergencyPauseInstance(opts *bind.TransactOpts, instance common.Address) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "emergencyPauseInstance", instance)
}

// EmergencyPauseInstance is a paid mutator transaction binding the contract method 0xf2d5de51.
//
// Solidity: function emergencyPauseInstance(address instance) returns()
func (_InvisiFactory *InvisiFactorySession) EmergencyPauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyPauseInstance(&_InvisiFactory.TransactOpts, instance)
}

// EmergencyPauseInstance is a paid mutator transaction binding the contract method 0xf2d5de51.
//
// Solidity: function emergencyPauseInstance(address instance) returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) EmergencyPauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyPauseInstance(&_InvisiFactory.TransactOpts, instance)
}

// EmergencyPauseRange is a paid mutator transaction binding the contract method 0xc7d0f69b.
//
// Solidity: function emergencyPauseRange(uint256 start, uint256 end) returns()
func (_InvisiFactory *InvisiFactoryTransactor) EmergencyPauseRange(opts *bind.TransactOpts, start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "emergencyPauseRange", start, end)
}

// EmergencyPauseRange is a paid mutator transaction binding the contract method 0xc7d0f69b.
//
// Solidity: function emergencyPauseRange(uint256 start, uint256 end) returns()
func (_InvisiFactory *InvisiFactorySession) EmergencyPauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyPauseRange(&_InvisiFactory.TransactOpts, start, end)
}

// EmergencyPauseRange is a paid mutator transaction binding the contract method 0xc7d0f69b.
//
// Solidity: function emergencyPauseRange(uint256 start, uint256 end) returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) EmergencyPauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyPauseRange(&_InvisiFactory.TransactOpts, start, end)
}

// InvisiFactoryEmergencyPauseInstanceIterator is returned from FilterEmergencyPauseInstance and is used to iterate over the raw logs and unpacked data for EmergencyPauseInstance events raised by the InvisiFactory contract.
type InvisiFactoryEmergencyPauseInstanceIterator struct {
	Event *InvisiFactoryEmergencyPauseInstance // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryEmergencyPauseInstanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryEmergencyPauseInstance)
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
		it.Event = new(InvisiFactoryEmergencyPauseInstance)
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
func (it *InvisiFactoryEmergencyPauseInstanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryEmergencyPauseInstanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryEmergencyPauseInstance represents a EmergencyPauseInstance event raised by the InvisiFactory contract.
type InvisiFactoryEmergencyPauseInstance struct {
	Instance common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterEmergencyPauseInstance is a free log retrieval operation binding the contract event 0x24b92279accff6ed4dff2406bf1a1425ea9047c24b5393bfe017b1273782038f.
//
// Solidity: event EmergencyPauseInstance(address indexed instance)
func (_InvisiFactory *InvisiFactoryFilterer) FilterEmergencyPauseInstance(opts *bind.FilterOpts, instance []common.Address) (*InvisiFactoryEmergencyPauseInstanceIterator, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "EmergencyPauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryEmergencyPauseInstanceIterator{contract: _InvisiFactory.contract, event: "EmergencyPauseInstance", logs: logs, sub: sub}, nil
}

// WatchEmergencyPauseInstance is a free log subscription operation binding the contract event 0x24b92279accff6ed4dff2406bf1a1425ea9047c24b5393bfe017b1273782038f.
//
// Solidity: event EmergencyPauseInstance(address indexed instance)
func (_InvisiFactory *InvisiFactoryFilterer) WatchEmergencyPauseInstance(opts *bind.WatchOpts, sink chan<- *InvisiFactoryEmergencyPauseInstance, instance []common.Address) (event.Subscription, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "EmergencyPauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryEmergencyPauseInstance)
				if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyPauseInstance", log); err != nil {
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

// ParseEmergencyPauseInstance is a log parse operation binding the contract event 0x24b92279accff6ed4dff2406bf1a1425ea9047c24b5393bfe017b1273782038f.
//
// Solidity: event EmergencyPauseInstance(address indexed instance)
func (_InvisiFactory *InvisiFactoryFilterer) ParseEmergencyPauseInstance(log types.Log) (*InvisiFactoryEmergencyPauseInstance, error) {
	event := new(InvisiFactoryEmergencyPauseInstance)
	if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyPauseInstance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiFactoryEmergencyPauseRangeIterator is returned from FilterEmergencyPauseRange and is used to iterate over the raw logs and unpacked data for EmergencyPauseRange events raised by the InvisiFactory contract.
type InvisiFactoryEmergencyPauseRangeIterator struct {
	Event *InvisiFactoryEmergencyPauseRange // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryEmergencyPauseRangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryEmergencyPauseRange)
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
		it.Event = new(InvisiFactoryEmergencyPauseRange)
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
func (it *InvisiFactoryEmergencyPauseRangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryEmergencyPauseRangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryEmergencyPauseRange represents a EmergencyPauseRange event raised by the InvisiFactory contract.
type InvisiFactoryEmergencyPauseRange struct {
	Start *big.Int
	End   *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterEmergencyPauseRange is a free log retrieval operation binding the contract event 0xd18e4026cc47782e9c90504841b795c9bf9b836bb5a0d76a8619dd12cc44be70.
//
// Solidity: event EmergencyPauseRange(uint256 start, uint256 end)
func (_InvisiFactory *InvisiFactoryFilterer) FilterEmergencyPauseRange(opts *bind.FilterOpts) (*InvisiFactoryEmergencyPauseRangeIterator, error) {

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "EmergencyPauseRange")
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryEmergencyPauseRangeIterator{contract: _InvisiFactory.contract, event: "EmergencyPauseRange", logs: logs, sub: sub}, nil
}

// WatchEmergencyPauseRange is a free log subscription operation binding the contract event 0xd18e4026cc47782e9c90504841b795c9bf9b836bb5a0d76a8619dd12cc44be70.
//
// Solidity: event EmergencyPauseRange(uint256 start, uint256 end)
func (_InvisiFactory *InvisiFactoryFilterer) WatchEmergencyPauseRange(opts *bind.WatchOpts, sink chan<- *InvisiFactoryEmergencyPauseRange) (event.Subscription, error) {

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "EmergencyPauseRange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryEmergencyPauseRange)
				if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyPauseRange", log); err != nil {
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

// ParseEmergencyPauseRange is a log parse operation binding the contract event 0xd18e4026cc47782e9c90504841b795c9bf9b836bb5a0d76a8619dd12cc44be70.
//
// Solidity: event EmergencyPauseRange(uint256 start, uint256 end)
func (_InvisiFactory *InvisiFactoryFilterer) ParseEmergencyPauseRange(log types.Log) (*InvisiFactoryEmergencyPauseRange, error) {
	event := new(InvisiFactoryEmergencyPauseRange)
	if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyPauseRange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiFactoryInstanceDeployedIterator is returned from FilterInstanceDeployed and is used to iterate over the raw logs and unpacked data for InstanceDeployed events raised by the InvisiFactory contract.
type InvisiFactoryInstanceDeployedIterator struct {
	Event *InvisiFactoryInstanceDeployed // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryInstanceDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryInstanceDeployed)
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
		it.Event = new(InvisiFactoryInstanceDeployed)
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
func (it *InvisiFactoryInstanceDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryInstanceDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryInstanceDeployed represents a InstanceDeployed event raised by the InvisiFactory contract.
type InvisiFactoryInstanceDeployed struct {
	Operator common.Address
	Token    common.Address
	Instance common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterInstanceDeployed is a free log retrieval operation binding the contract event 0x36766c358955a6abb308cd2a6dccd5a0bd5162ea88313b29c34781a7988186a4.
//
// Solidity: event InstanceDeployed(address indexed operator, address indexed token, address instance)
func (_InvisiFactory *InvisiFactoryFilterer) FilterInstanceDeployed(opts *bind.FilterOpts, operator []common.Address, token []common.Address) (*InvisiFactoryInstanceDeployedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "InstanceDeployed", operatorRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryInstanceDeployedIterator{contract: _InvisiFactory.contract, event: "InstanceDeployed", logs: logs, sub: sub}, nil
}

// WatchInstanceDeployed is a free log subscription operation binding the contract event 0x36766c358955a6abb308cd2a6dccd5a0bd5162ea88313b29c34781a7988186a4.
//
// Solidity: event InstanceDeployed(address indexed operator, address indexed token, address instance)
func (_InvisiFactory *InvisiFactoryFilterer) WatchInstanceDeployed(opts *bind.WatchOpts, sink chan<- *InvisiFactoryInstanceDeployed, operator []common.Address, token []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "InstanceDeployed", operatorRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryInstanceDeployed)
				if err := _InvisiFactory.contract.UnpackLog(event, "InstanceDeployed", log); err != nil {
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

// ParseInstanceDeployed is a log parse operation binding the contract event 0x36766c358955a6abb308cd2a6dccd5a0bd5162ea88313b29c34781a7988186a4.
//
// Solidity: event InstanceDeployed(address indexed operator, address indexed token, address instance)
func (_InvisiFactory *InvisiFactoryFilterer) ParseInstanceDeployed(log types.Log) (*InvisiFactoryInstanceDeployed, error) {
	event := new(InvisiFactoryInstanceDeployed)
	if err := _InvisiFactory.contract.UnpackLog(event, "InstanceDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
