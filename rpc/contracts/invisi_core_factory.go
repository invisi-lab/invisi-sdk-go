// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	protocolcontracts "github.com/invisi-labs/invisi-sdk-go/protocol/contracts"

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

// InvisiCoreFactoryMetaData contains all meta data concerning the InvisiCoreFactory contract.
var InvisiCoreFactoryMetaData = &bind.MetaData{ABI: protocolcontracts.InvisiCoreFactoryMetaData.ABI}

// InvisiCoreFactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use InvisiCoreFactoryMetaData.ABI instead.
var InvisiCoreFactoryABI = InvisiCoreFactoryMetaData.ABI

// InvisiCoreFactory is an auto generated Go binding around an Ethereum contract.
type InvisiCoreFactory struct {
	InvisiCoreFactoryCaller     // Read-only binding to the contract
	InvisiCoreFactoryTransactor // Write-only binding to the contract
	InvisiCoreFactoryFilterer   // Log filterer for contract events
}

// InvisiCoreFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type InvisiCoreFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiCoreFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InvisiCoreFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiCoreFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InvisiCoreFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiCoreFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InvisiCoreFactorySession struct {
	Contract     *InvisiCoreFactory // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// InvisiCoreFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InvisiCoreFactoryCallerSession struct {
	Contract *InvisiCoreFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// InvisiCoreFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InvisiCoreFactoryTransactorSession struct {
	Contract     *InvisiCoreFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// InvisiCoreFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type InvisiCoreFactoryRaw struct {
	Contract *InvisiCoreFactory // Generic contract binding to access the raw methods on
}

// InvisiCoreFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InvisiCoreFactoryCallerRaw struct {
	Contract *InvisiCoreFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// InvisiCoreFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InvisiCoreFactoryTransactorRaw struct {
	Contract *InvisiCoreFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInvisiCoreFactory creates a new instance of InvisiCoreFactory, bound to a specific deployed contract.
func NewInvisiCoreFactory(address common.Address, backend bind.ContractBackend) (*InvisiCoreFactory, error) {
	contract, err := bindInvisiCoreFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactory{InvisiCoreFactoryCaller: InvisiCoreFactoryCaller{contract: contract}, InvisiCoreFactoryTransactor: InvisiCoreFactoryTransactor{contract: contract}, InvisiCoreFactoryFilterer: InvisiCoreFactoryFilterer{contract: contract}}, nil
}

// NewInvisiCoreFactoryCaller creates a new read-only instance of InvisiCoreFactory, bound to a specific deployed contract.
func NewInvisiCoreFactoryCaller(address common.Address, caller bind.ContractCaller) (*InvisiCoreFactoryCaller, error) {
	contract, err := bindInvisiCoreFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryCaller{contract: contract}, nil
}

// NewInvisiCoreFactoryTransactor creates a new write-only instance of InvisiCoreFactory, bound to a specific deployed contract.
func NewInvisiCoreFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*InvisiCoreFactoryTransactor, error) {
	contract, err := bindInvisiCoreFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryTransactor{contract: contract}, nil
}

// NewInvisiCoreFactoryFilterer creates a new log filterer instance of InvisiCoreFactory, bound to a specific deployed contract.
func NewInvisiCoreFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*InvisiCoreFactoryFilterer, error) {
	contract, err := bindInvisiCoreFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryFilterer{contract: contract}, nil
}

// bindInvisiCoreFactory binds a generic wrapper to an already deployed contract.
func bindInvisiCoreFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InvisiCoreFactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiCoreFactory *InvisiCoreFactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiCoreFactory.Contract.InvisiCoreFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiCoreFactory *InvisiCoreFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.InvisiCoreFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiCoreFactory *InvisiCoreFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.InvisiCoreFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiCoreFactory *InvisiCoreFactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiCoreFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.contract.Transact(opts, method, params...)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) VERSION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiCoreFactory *InvisiCoreFactorySession) VERSION() (*big.Int, error) {
	return _InvisiCoreFactory.Contract.VERSION(&_InvisiCoreFactory.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) VERSION() (*big.Int, error) {
	return _InvisiCoreFactory.Contract.VERSION(&_InvisiCoreFactory.CallOpts)
}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x716a840a.
//
// Solidity: function computeInstanceAddress(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) view returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) ComputeInstanceAddress(opts *bind.CallOpts, creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (common.Address, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "computeInstanceAddress", creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x716a840a.
//
// Solidity: function computeInstanceAddress(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) view returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactorySession) ComputeInstanceAddress(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (common.Address, error) {
	return _InvisiCoreFactory.Contract.ComputeInstanceAddress(&_InvisiCoreFactory.CallOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x716a840a.
//
// Solidity: function computeInstanceAddress(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) view returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) ComputeInstanceAddress(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (common.Address, error) {
	return _InvisiCoreFactory.Contract.ComputeInstanceAddress(&_InvisiCoreFactory.CallOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// CreationCodeHash is a free data retrieval call binding the contract method 0x47191163.
//
// Solidity: function creationCodeHash() view returns(bytes32)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) CreationCodeHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "creationCodeHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CreationCodeHash is a free data retrieval call binding the contract method 0x47191163.
//
// Solidity: function creationCodeHash() view returns(bytes32)
func (_InvisiCoreFactory *InvisiCoreFactorySession) CreationCodeHash() ([32]byte, error) {
	return _InvisiCoreFactory.Contract.CreationCodeHash(&_InvisiCoreFactory.CallOpts)
}

// CreationCodeHash is a free data retrieval call binding the contract method 0x47191163.
//
// Solidity: function creationCodeHash() view returns(bytes32)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) CreationCodeHash() ([32]byte, error) {
	return _InvisiCoreFactory.Contract.CreationCodeHash(&_InvisiCoreFactory.CallOpts)
}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xcdcec245.
//
// Solidity: function getInstanceAddress(bytes32 salt) view returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) GetInstanceAddress(opts *bind.CallOpts, salt [32]byte) (common.Address, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "getInstanceAddress", salt)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xcdcec245.
//
// Solidity: function getInstanceAddress(bytes32 salt) view returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactorySession) GetInstanceAddress(salt [32]byte) (common.Address, error) {
	return _InvisiCoreFactory.Contract.GetInstanceAddress(&_InvisiCoreFactory.CallOpts, salt)
}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xcdcec245.
//
// Solidity: function getInstanceAddress(bytes32 salt) view returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) GetInstanceAddress(salt [32]byte) (common.Address, error) {
	return _InvisiCoreFactory.Contract.GetInstanceAddress(&_InvisiCoreFactory.CallOpts, salt)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) GlobalEmergency(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "globalEmergency")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactorySession) GlobalEmergency() (common.Address, error) {
	return _InvisiCoreFactory.Contract.GlobalEmergency(&_InvisiCoreFactory.CallOpts)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) GlobalEmergency() (common.Address, error) {
	return _InvisiCoreFactory.Contract.GlobalEmergency(&_InvisiCoreFactory.CallOpts)
}

// InstanceCount is a free data retrieval call binding the contract method 0xec56c716.
//
// Solidity: function instanceCount() view returns(uint256)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) InstanceCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "instanceCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InstanceCount is a free data retrieval call binding the contract method 0xec56c716.
//
// Solidity: function instanceCount() view returns(uint256)
func (_InvisiCoreFactory *InvisiCoreFactorySession) InstanceCount() (*big.Int, error) {
	return _InvisiCoreFactory.Contract.InstanceCount(&_InvisiCoreFactory.CallOpts)
}

// InstanceCount is a free data retrieval call binding the contract method 0xec56c716.
//
// Solidity: function instanceCount() view returns(uint256)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) InstanceCount() (*big.Int, error) {
	return _InvisiCoreFactory.Contract.InstanceCount(&_InvisiCoreFactory.CallOpts)
}

// IsInstance is a free data retrieval call binding the contract method 0x6b44e6be.
//
// Solidity: function isInstance(address ) view returns(bool)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) IsInstance(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "isInstance", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInstance is a free data retrieval call binding the contract method 0x6b44e6be.
//
// Solidity: function isInstance(address ) view returns(bool)
func (_InvisiCoreFactory *InvisiCoreFactorySession) IsInstance(arg0 common.Address) (bool, error) {
	return _InvisiCoreFactory.Contract.IsInstance(&_InvisiCoreFactory.CallOpts, arg0)
}

// IsInstance is a free data retrieval call binding the contract method 0x6b44e6be.
//
// Solidity: function isInstance(address ) view returns(bool)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) IsInstance(arg0 common.Address) (bool, error) {
	return _InvisiCoreFactory.Contract.IsInstance(&_InvisiCoreFactory.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactorySession) Owner() (common.Address, error) {
	return _InvisiCoreFactory.Contract.Owner(&_InvisiCoreFactory.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) Owner() (common.Address, error) {
	return _InvisiCoreFactory.Contract.Owner(&_InvisiCoreFactory.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactorySession) PendingOwner() (common.Address, error) {
	return _InvisiCoreFactory.Contract.PendingOwner(&_InvisiCoreFactory.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) PendingOwner() (common.Address, error) {
	return _InvisiCoreFactory.Contract.PendingOwner(&_InvisiCoreFactory.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCaller) VerifierRouter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCoreFactory.contract.Call(opts, &out, "verifierRouter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactorySession) VerifierRouter() (common.Address, error) {
	return _InvisiCoreFactory.Contract.VerifierRouter(&_InvisiCoreFactory.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiCoreFactory *InvisiCoreFactoryCallerSession) VerifierRouter() (common.Address, error) {
	return _InvisiCoreFactory.Contract.VerifierRouter(&_InvisiCoreFactory.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.AcceptOwnership(&_InvisiCoreFactory.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.AcceptOwnership(&_InvisiCoreFactory.TransactOpts)
}

// Deploy is a paid mutator transaction binding the contract method 0x64aa77a2.
//
// Solidity: function deploy(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) Deploy(opts *bind.TransactOpts, creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "deploy", creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// Deploy is a paid mutator transaction binding the contract method 0x64aa77a2.
//
// Solidity: function deploy(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactorySession) Deploy(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.Deploy(&_InvisiCoreFactory.TransactOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// Deploy is a paid mutator transaction binding the contract method 0x64aa77a2.
//
// Solidity: function deploy(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) returns(address instance)
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) Deploy(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.Deploy(&_InvisiCoreFactory.TransactOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// EmergencyPauseInstance is a paid mutator transaction binding the contract method 0xf2d5de51.
//
// Solidity: function emergencyPauseInstance(address instance) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) EmergencyPauseInstance(opts *bind.TransactOpts, instance common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "emergencyPauseInstance", instance)
}

// EmergencyPauseInstance is a paid mutator transaction binding the contract method 0xf2d5de51.
//
// Solidity: function emergencyPauseInstance(address instance) returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) EmergencyPauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyPauseInstance(&_InvisiCoreFactory.TransactOpts, instance)
}

// EmergencyPauseInstance is a paid mutator transaction binding the contract method 0xf2d5de51.
//
// Solidity: function emergencyPauseInstance(address instance) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) EmergencyPauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyPauseInstance(&_InvisiCoreFactory.TransactOpts, instance)
}

// EmergencyPauseRange is a paid mutator transaction binding the contract method 0xc7d0f69b.
//
// Solidity: function emergencyPauseRange(uint256 start, uint256 end) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) EmergencyPauseRange(opts *bind.TransactOpts, start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "emergencyPauseRange", start, end)
}

// EmergencyPauseRange is a paid mutator transaction binding the contract method 0xc7d0f69b.
//
// Solidity: function emergencyPauseRange(uint256 start, uint256 end) returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) EmergencyPauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyPauseRange(&_InvisiCoreFactory.TransactOpts, start, end)
}

// EmergencyPauseRange is a paid mutator transaction binding the contract method 0xc7d0f69b.
//
// Solidity: function emergencyPauseRange(uint256 start, uint256 end) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) EmergencyPauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyPauseRange(&_InvisiCoreFactory.TransactOpts, start, end)
}

// EmergencyUnpauseInstance is a paid mutator transaction binding the contract method 0x3dba177e.
//
// Solidity: function emergencyUnpauseInstance(address instance) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) EmergencyUnpauseInstance(opts *bind.TransactOpts, instance common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "emergencyUnpauseInstance", instance)
}

// EmergencyUnpauseInstance is a paid mutator transaction binding the contract method 0x3dba177e.
//
// Solidity: function emergencyUnpauseInstance(address instance) returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) EmergencyUnpauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyUnpauseInstance(&_InvisiCoreFactory.TransactOpts, instance)
}

// EmergencyUnpauseInstance is a paid mutator transaction binding the contract method 0x3dba177e.
//
// Solidity: function emergencyUnpauseInstance(address instance) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) EmergencyUnpauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyUnpauseInstance(&_InvisiCoreFactory.TransactOpts, instance)
}

// EmergencyUnpauseRange is a paid mutator transaction binding the contract method 0x64412fa1.
//
// Solidity: function emergencyUnpauseRange(uint256 start, uint256 end) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) EmergencyUnpauseRange(opts *bind.TransactOpts, start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "emergencyUnpauseRange", start, end)
}

// EmergencyUnpauseRange is a paid mutator transaction binding the contract method 0x64412fa1.
//
// Solidity: function emergencyUnpauseRange(uint256 start, uint256 end) returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) EmergencyUnpauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyUnpauseRange(&_InvisiCoreFactory.TransactOpts, start, end)
}

// EmergencyUnpauseRange is a paid mutator transaction binding the contract method 0x64412fa1.
//
// Solidity: function emergencyUnpauseRange(uint256 start, uint256 end) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) EmergencyUnpauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.EmergencyUnpauseRange(&_InvisiCoreFactory.TransactOpts, start, end)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.RenounceOwnership(&_InvisiCoreFactory.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.RenounceOwnership(&_InvisiCoreFactory.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiCoreFactory *InvisiCoreFactorySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.TransferOwnership(&_InvisiCoreFactory.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiCoreFactory *InvisiCoreFactoryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiCoreFactory.Contract.TransferOwnership(&_InvisiCoreFactory.TransactOpts, newOwner)
}

// InvisiCoreFactoryEmergencyPauseInstanceIterator is returned from FilterEmergencyPauseInstance and is used to iterate over the raw logs and unpacked data for EmergencyPauseInstance events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyPauseInstanceIterator struct {
	Event *InvisiCoreFactoryEmergencyPauseInstance // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryEmergencyPauseInstanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryEmergencyPauseInstance)
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
		it.Event = new(InvisiCoreFactoryEmergencyPauseInstance)
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
func (it *InvisiCoreFactoryEmergencyPauseInstanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryEmergencyPauseInstanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryEmergencyPauseInstance represents a EmergencyPauseInstance event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyPauseInstance struct {
	Instance common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterEmergencyPauseInstance is a free log retrieval operation binding the contract event 0x24b92279accff6ed4dff2406bf1a1425ea9047c24b5393bfe017b1273782038f.
//
// Solidity: event EmergencyPauseInstance(address indexed instance)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterEmergencyPauseInstance(opts *bind.FilterOpts, instance []common.Address) (*InvisiCoreFactoryEmergencyPauseInstanceIterator, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "EmergencyPauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryEmergencyPauseInstanceIterator{contract: _InvisiCoreFactory.contract, event: "EmergencyPauseInstance", logs: logs, sub: sub}, nil
}

// WatchEmergencyPauseInstance is a free log subscription operation binding the contract event 0x24b92279accff6ed4dff2406bf1a1425ea9047c24b5393bfe017b1273782038f.
//
// Solidity: event EmergencyPauseInstance(address indexed instance)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchEmergencyPauseInstance(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryEmergencyPauseInstance, instance []common.Address) (event.Subscription, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "EmergencyPauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryEmergencyPauseInstance)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyPauseInstance", log); err != nil {
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
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseEmergencyPauseInstance(log types.Log) (*InvisiCoreFactoryEmergencyPauseInstance, error) {
	event := new(InvisiCoreFactoryEmergencyPauseInstance)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyPauseInstance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreFactoryEmergencyPauseRangeIterator is returned from FilterEmergencyPauseRange and is used to iterate over the raw logs and unpacked data for EmergencyPauseRange events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyPauseRangeIterator struct {
	Event *InvisiCoreFactoryEmergencyPauseRange // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryEmergencyPauseRangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryEmergencyPauseRange)
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
		it.Event = new(InvisiCoreFactoryEmergencyPauseRange)
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
func (it *InvisiCoreFactoryEmergencyPauseRangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryEmergencyPauseRangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryEmergencyPauseRange represents a EmergencyPauseRange event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyPauseRange struct {
	Start *big.Int
	End   *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterEmergencyPauseRange is a free log retrieval operation binding the contract event 0xd18e4026cc47782e9c90504841b795c9bf9b836bb5a0d76a8619dd12cc44be70.
//
// Solidity: event EmergencyPauseRange(uint256 start, uint256 end)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterEmergencyPauseRange(opts *bind.FilterOpts) (*InvisiCoreFactoryEmergencyPauseRangeIterator, error) {

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "EmergencyPauseRange")
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryEmergencyPauseRangeIterator{contract: _InvisiCoreFactory.contract, event: "EmergencyPauseRange", logs: logs, sub: sub}, nil
}

// WatchEmergencyPauseRange is a free log subscription operation binding the contract event 0xd18e4026cc47782e9c90504841b795c9bf9b836bb5a0d76a8619dd12cc44be70.
//
// Solidity: event EmergencyPauseRange(uint256 start, uint256 end)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchEmergencyPauseRange(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryEmergencyPauseRange) (event.Subscription, error) {

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "EmergencyPauseRange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryEmergencyPauseRange)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyPauseRange", log); err != nil {
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
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseEmergencyPauseRange(log types.Log) (*InvisiCoreFactoryEmergencyPauseRange, error) {
	event := new(InvisiCoreFactoryEmergencyPauseRange)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyPauseRange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreFactoryEmergencyUnpauseInstanceIterator is returned from FilterEmergencyUnpauseInstance and is used to iterate over the raw logs and unpacked data for EmergencyUnpauseInstance events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyUnpauseInstanceIterator struct {
	Event *InvisiCoreFactoryEmergencyUnpauseInstance // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryEmergencyUnpauseInstanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryEmergencyUnpauseInstance)
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
		it.Event = new(InvisiCoreFactoryEmergencyUnpauseInstance)
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
func (it *InvisiCoreFactoryEmergencyUnpauseInstanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryEmergencyUnpauseInstanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryEmergencyUnpauseInstance represents a EmergencyUnpauseInstance event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyUnpauseInstance struct {
	Instance common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterEmergencyUnpauseInstance is a free log retrieval operation binding the contract event 0x515ef1b4563c5cfe000fbd897a1bf4191a51507c6fd4fbe8a3a1b0ce7e402e21.
//
// Solidity: event EmergencyUnpauseInstance(address indexed instance)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterEmergencyUnpauseInstance(opts *bind.FilterOpts, instance []common.Address) (*InvisiCoreFactoryEmergencyUnpauseInstanceIterator, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "EmergencyUnpauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryEmergencyUnpauseInstanceIterator{contract: _InvisiCoreFactory.contract, event: "EmergencyUnpauseInstance", logs: logs, sub: sub}, nil
}

// WatchEmergencyUnpauseInstance is a free log subscription operation binding the contract event 0x515ef1b4563c5cfe000fbd897a1bf4191a51507c6fd4fbe8a3a1b0ce7e402e21.
//
// Solidity: event EmergencyUnpauseInstance(address indexed instance)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchEmergencyUnpauseInstance(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryEmergencyUnpauseInstance, instance []common.Address) (event.Subscription, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "EmergencyUnpauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryEmergencyUnpauseInstance)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyUnpauseInstance", log); err != nil {
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

// ParseEmergencyUnpauseInstance is a log parse operation binding the contract event 0x515ef1b4563c5cfe000fbd897a1bf4191a51507c6fd4fbe8a3a1b0ce7e402e21.
//
// Solidity: event EmergencyUnpauseInstance(address indexed instance)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseEmergencyUnpauseInstance(log types.Log) (*InvisiCoreFactoryEmergencyUnpauseInstance, error) {
	event := new(InvisiCoreFactoryEmergencyUnpauseInstance)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyUnpauseInstance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreFactoryEmergencyUnpauseRangeIterator is returned from FilterEmergencyUnpauseRange and is used to iterate over the raw logs and unpacked data for EmergencyUnpauseRange events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyUnpauseRangeIterator struct {
	Event *InvisiCoreFactoryEmergencyUnpauseRange // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryEmergencyUnpauseRangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryEmergencyUnpauseRange)
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
		it.Event = new(InvisiCoreFactoryEmergencyUnpauseRange)
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
func (it *InvisiCoreFactoryEmergencyUnpauseRangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryEmergencyUnpauseRangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryEmergencyUnpauseRange represents a EmergencyUnpauseRange event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryEmergencyUnpauseRange struct {
	Start *big.Int
	End   *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterEmergencyUnpauseRange is a free log retrieval operation binding the contract event 0xe1137af9d2e3b90b122226a33ada87fc87a62da822a8774db7beb4e77b8741be.
//
// Solidity: event EmergencyUnpauseRange(uint256 start, uint256 end)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterEmergencyUnpauseRange(opts *bind.FilterOpts) (*InvisiCoreFactoryEmergencyUnpauseRangeIterator, error) {

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "EmergencyUnpauseRange")
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryEmergencyUnpauseRangeIterator{contract: _InvisiCoreFactory.contract, event: "EmergencyUnpauseRange", logs: logs, sub: sub}, nil
}

// WatchEmergencyUnpauseRange is a free log subscription operation binding the contract event 0xe1137af9d2e3b90b122226a33ada87fc87a62da822a8774db7beb4e77b8741be.
//
// Solidity: event EmergencyUnpauseRange(uint256 start, uint256 end)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchEmergencyUnpauseRange(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryEmergencyUnpauseRange) (event.Subscription, error) {

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "EmergencyUnpauseRange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryEmergencyUnpauseRange)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyUnpauseRange", log); err != nil {
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

// ParseEmergencyUnpauseRange is a log parse operation binding the contract event 0xe1137af9d2e3b90b122226a33ada87fc87a62da822a8774db7beb4e77b8741be.
//
// Solidity: event EmergencyUnpauseRange(uint256 start, uint256 end)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseEmergencyUnpauseRange(log types.Log) (*InvisiCoreFactoryEmergencyUnpauseRange, error) {
	event := new(InvisiCoreFactoryEmergencyUnpauseRange)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "EmergencyUnpauseRange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreFactoryInstanceDeployedIterator is returned from FilterInstanceDeployed and is used to iterate over the raw logs and unpacked data for InstanceDeployed events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryInstanceDeployedIterator struct {
	Event *InvisiCoreFactoryInstanceDeployed // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryInstanceDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryInstanceDeployed)
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
		it.Event = new(InvisiCoreFactoryInstanceDeployed)
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
func (it *InvisiCoreFactoryInstanceDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryInstanceDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryInstanceDeployed represents a InstanceDeployed event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryInstanceDeployed struct {
	Salt         [32]byte
	Instance     common.Address
	InitialOwner common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterInstanceDeployed is a free log retrieval operation binding the contract event 0xd321169aa5f5945b176ea1ec4917461694c2aafd7343099fd7e15462591404fd.
//
// Solidity: event InstanceDeployed(bytes32 indexed salt, address instance, address initialOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterInstanceDeployed(opts *bind.FilterOpts, salt [][32]byte) (*InvisiCoreFactoryInstanceDeployedIterator, error) {

	var saltRule []interface{}
	for _, saltItem := range salt {
		saltRule = append(saltRule, saltItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "InstanceDeployed", saltRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryInstanceDeployedIterator{contract: _InvisiCoreFactory.contract, event: "InstanceDeployed", logs: logs, sub: sub}, nil
}

// WatchInstanceDeployed is a free log subscription operation binding the contract event 0xd321169aa5f5945b176ea1ec4917461694c2aafd7343099fd7e15462591404fd.
//
// Solidity: event InstanceDeployed(bytes32 indexed salt, address instance, address initialOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchInstanceDeployed(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryInstanceDeployed, salt [][32]byte) (event.Subscription, error) {

	var saltRule []interface{}
	for _, saltItem := range salt {
		saltRule = append(saltRule, saltItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "InstanceDeployed", saltRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryInstanceDeployed)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "InstanceDeployed", log); err != nil {
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

// ParseInstanceDeployed is a log parse operation binding the contract event 0xd321169aa5f5945b176ea1ec4917461694c2aafd7343099fd7e15462591404fd.
//
// Solidity: event InstanceDeployed(bytes32 indexed salt, address instance, address initialOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseInstanceDeployed(log types.Log) (*InvisiCoreFactoryInstanceDeployed, error) {
	event := new(InvisiCoreFactoryInstanceDeployed)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "InstanceDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreFactoryOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryOwnershipTransferStartedIterator struct {
	Event *InvisiCoreFactoryOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryOwnershipTransferStarted)
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
		it.Event = new(InvisiCoreFactoryOwnershipTransferStarted)
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
func (it *InvisiCoreFactoryOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiCoreFactoryOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryOwnershipTransferStartedIterator{contract: _InvisiCoreFactory.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryOwnershipTransferStarted)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseOwnershipTransferStarted(log types.Log) (*InvisiCoreFactoryOwnershipTransferStarted, error) {
	event := new(InvisiCoreFactoryOwnershipTransferStarted)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreFactoryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryOwnershipTransferredIterator struct {
	Event *InvisiCoreFactoryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *InvisiCoreFactoryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreFactoryOwnershipTransferred)
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
		it.Event = new(InvisiCoreFactoryOwnershipTransferred)
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
func (it *InvisiCoreFactoryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreFactoryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreFactoryOwnershipTransferred represents a OwnershipTransferred event raised by the InvisiCoreFactory contract.
type InvisiCoreFactoryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiCoreFactoryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFactoryOwnershipTransferredIterator{contract: _InvisiCoreFactory.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *InvisiCoreFactoryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCoreFactory.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreFactoryOwnershipTransferred)
				if err := _InvisiCoreFactory.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiCoreFactory *InvisiCoreFactoryFilterer) ParseOwnershipTransferred(log types.Log) (*InvisiCoreFactoryOwnershipTransferred, error) {
	event := new(InvisiCoreFactoryOwnershipTransferred)
	if err := _InvisiCoreFactory.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
