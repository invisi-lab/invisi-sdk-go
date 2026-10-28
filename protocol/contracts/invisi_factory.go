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
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"initialOwner_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"globalEmergency_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"verifierRouter_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"creationCodeHash_\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"computeInstanceAddress\",\"inputs\":[{\"name\":\"creationCode\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"initialOwner_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"emergencyAdmin_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"screener_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"addressRegistry_\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"instance\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creationCodeHash\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deploy\",\"inputs\":[{\"name\":\"creationCode\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"initialOwner_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"emergencyAdmin_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"screener_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"addressRegistry_\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"instance\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"emergencyPauseInstance\",\"inputs\":[{\"name\":\"instance\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"emergencyPauseRange\",\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"end\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"emergencyUnpauseInstance\",\"inputs\":[{\"name\":\"instance\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"emergencyUnpauseRange\",\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"end\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getInstanceAddress\",\"inputs\":[{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"instance\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalEmergency\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"instanceCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isInstance\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verifierRouter\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifierRouter\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"EmergencyPauseInstance\",\"inputs\":[{\"name\":\"instance\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EmergencyPauseRange\",\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"end\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EmergencyUnpauseInstance\",\"inputs\":[{\"name\":\"instance\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EmergencyUnpauseRange\",\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"end\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"InstanceDeployed\",\"inputs\":[{\"name\":\"salt\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"instance\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InstanceAlreadyDeployed\",\"inputs\":[{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidCreationCode\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAnInstance\",\"inputs\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotGlobalEmergency\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
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

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x716a840a.
//
// Solidity: function computeInstanceAddress(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCaller) ComputeInstanceAddress(opts *bind.CallOpts, creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "computeInstanceAddress", creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x716a840a.
//
// Solidity: function computeInstanceAddress(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) view returns(address instance)
func (_InvisiFactory *InvisiFactorySession) ComputeInstanceAddress(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (common.Address, error) {
	return _InvisiFactory.Contract.ComputeInstanceAddress(&_InvisiFactory.CallOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// ComputeInstanceAddress is a free data retrieval call binding the contract method 0x716a840a.
//
// Solidity: function computeInstanceAddress(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCallerSession) ComputeInstanceAddress(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (common.Address, error) {
	return _InvisiFactory.Contract.ComputeInstanceAddress(&_InvisiFactory.CallOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
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

// GetInstanceAddress is a free data retrieval call binding the contract method 0xcdcec245.
//
// Solidity: function getInstanceAddress(bytes32 salt) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCaller) GetInstanceAddress(opts *bind.CallOpts, salt [32]byte) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "getInstanceAddress", salt)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xcdcec245.
//
// Solidity: function getInstanceAddress(bytes32 salt) view returns(address instance)
func (_InvisiFactory *InvisiFactorySession) GetInstanceAddress(salt [32]byte) (common.Address, error) {
	return _InvisiFactory.Contract.GetInstanceAddress(&_InvisiFactory.CallOpts, salt)
}

// GetInstanceAddress is a free data retrieval call binding the contract method 0xcdcec245.
//
// Solidity: function getInstanceAddress(bytes32 salt) view returns(address instance)
func (_InvisiFactory *InvisiFactoryCallerSession) GetInstanceAddress(salt [32]byte) (common.Address, error) {
	return _InvisiFactory.Contract.GetInstanceAddress(&_InvisiFactory.CallOpts, salt)
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

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiFactory *InvisiFactoryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiFactory *InvisiFactorySession) Owner() (common.Address, error) {
	return _InvisiFactory.Contract.Owner(&_InvisiFactory.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiFactory *InvisiFactoryCallerSession) Owner() (common.Address, error) {
	return _InvisiFactory.Contract.Owner(&_InvisiFactory.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiFactory *InvisiFactoryCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiFactory.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiFactory *InvisiFactorySession) PendingOwner() (common.Address, error) {
	return _InvisiFactory.Contract.PendingOwner(&_InvisiFactory.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiFactory *InvisiFactoryCallerSession) PendingOwner() (common.Address, error) {
	return _InvisiFactory.Contract.PendingOwner(&_InvisiFactory.CallOpts)
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

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiFactory *InvisiFactoryTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiFactory *InvisiFactorySession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiFactory.Contract.AcceptOwnership(&_InvisiFactory.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiFactory.Contract.AcceptOwnership(&_InvisiFactory.TransactOpts)
}

// Deploy is a paid mutator transaction binding the contract method 0x64aa77a2.
//
// Solidity: function deploy(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) returns(address instance)
func (_InvisiFactory *InvisiFactoryTransactor) Deploy(opts *bind.TransactOpts, creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "deploy", creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// Deploy is a paid mutator transaction binding the contract method 0x64aa77a2.
//
// Solidity: function deploy(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) returns(address instance)
func (_InvisiFactory *InvisiFactorySession) Deploy(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.Deploy(&_InvisiFactory.TransactOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
}

// Deploy is a paid mutator transaction binding the contract method 0x64aa77a2.
//
// Solidity: function deploy(bytes creationCode, bytes32 salt, address initialOwner_, address emergencyAdmin_, address screener_, address addressRegistry_) returns(address instance)
func (_InvisiFactory *InvisiFactoryTransactorSession) Deploy(creationCode []byte, salt [32]byte, initialOwner_ common.Address, emergencyAdmin_ common.Address, screener_ common.Address, addressRegistry_ common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.Deploy(&_InvisiFactory.TransactOpts, creationCode, salt, initialOwner_, emergencyAdmin_, screener_, addressRegistry_)
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

// EmergencyUnpauseInstance is a paid mutator transaction binding the contract method 0x3dba177e.
//
// Solidity: function emergencyUnpauseInstance(address instance) returns()
func (_InvisiFactory *InvisiFactoryTransactor) EmergencyUnpauseInstance(opts *bind.TransactOpts, instance common.Address) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "emergencyUnpauseInstance", instance)
}

// EmergencyUnpauseInstance is a paid mutator transaction binding the contract method 0x3dba177e.
//
// Solidity: function emergencyUnpauseInstance(address instance) returns()
func (_InvisiFactory *InvisiFactorySession) EmergencyUnpauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyUnpauseInstance(&_InvisiFactory.TransactOpts, instance)
}

// EmergencyUnpauseInstance is a paid mutator transaction binding the contract method 0x3dba177e.
//
// Solidity: function emergencyUnpauseInstance(address instance) returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) EmergencyUnpauseInstance(instance common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyUnpauseInstance(&_InvisiFactory.TransactOpts, instance)
}

// EmergencyUnpauseRange is a paid mutator transaction binding the contract method 0x64412fa1.
//
// Solidity: function emergencyUnpauseRange(uint256 start, uint256 end) returns()
func (_InvisiFactory *InvisiFactoryTransactor) EmergencyUnpauseRange(opts *bind.TransactOpts, start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "emergencyUnpauseRange", start, end)
}

// EmergencyUnpauseRange is a paid mutator transaction binding the contract method 0x64412fa1.
//
// Solidity: function emergencyUnpauseRange(uint256 start, uint256 end) returns()
func (_InvisiFactory *InvisiFactorySession) EmergencyUnpauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyUnpauseRange(&_InvisiFactory.TransactOpts, start, end)
}

// EmergencyUnpauseRange is a paid mutator transaction binding the contract method 0x64412fa1.
//
// Solidity: function emergencyUnpauseRange(uint256 start, uint256 end) returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) EmergencyUnpauseRange(start *big.Int, end *big.Int) (*types.Transaction, error) {
	return _InvisiFactory.Contract.EmergencyUnpauseRange(&_InvisiFactory.TransactOpts, start, end)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiFactory *InvisiFactoryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiFactory *InvisiFactorySession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiFactory.Contract.RenounceOwnership(&_InvisiFactory.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiFactory.Contract.RenounceOwnership(&_InvisiFactory.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiFactory *InvisiFactoryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _InvisiFactory.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiFactory *InvisiFactorySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.TransferOwnership(&_InvisiFactory.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiFactory *InvisiFactoryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiFactory.Contract.TransferOwnership(&_InvisiFactory.TransactOpts, newOwner)
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

// InvisiFactoryEmergencyUnpauseInstanceIterator is returned from FilterEmergencyUnpauseInstance and is used to iterate over the raw logs and unpacked data for EmergencyUnpauseInstance events raised by the InvisiFactory contract.
type InvisiFactoryEmergencyUnpauseInstanceIterator struct {
	Event *InvisiFactoryEmergencyUnpauseInstance // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryEmergencyUnpauseInstanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryEmergencyUnpauseInstance)
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
		it.Event = new(InvisiFactoryEmergencyUnpauseInstance)
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
func (it *InvisiFactoryEmergencyUnpauseInstanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryEmergencyUnpauseInstanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryEmergencyUnpauseInstance represents a EmergencyUnpauseInstance event raised by the InvisiFactory contract.
type InvisiFactoryEmergencyUnpauseInstance struct {
	Instance common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterEmergencyUnpauseInstance is a free log retrieval operation binding the contract event 0x515ef1b4563c5cfe000fbd897a1bf4191a51507c6fd4fbe8a3a1b0ce7e402e21.
//
// Solidity: event EmergencyUnpauseInstance(address indexed instance)
func (_InvisiFactory *InvisiFactoryFilterer) FilterEmergencyUnpauseInstance(opts *bind.FilterOpts, instance []common.Address) (*InvisiFactoryEmergencyUnpauseInstanceIterator, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "EmergencyUnpauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryEmergencyUnpauseInstanceIterator{contract: _InvisiFactory.contract, event: "EmergencyUnpauseInstance", logs: logs, sub: sub}, nil
}

// WatchEmergencyUnpauseInstance is a free log subscription operation binding the contract event 0x515ef1b4563c5cfe000fbd897a1bf4191a51507c6fd4fbe8a3a1b0ce7e402e21.
//
// Solidity: event EmergencyUnpauseInstance(address indexed instance)
func (_InvisiFactory *InvisiFactoryFilterer) WatchEmergencyUnpauseInstance(opts *bind.WatchOpts, sink chan<- *InvisiFactoryEmergencyUnpauseInstance, instance []common.Address) (event.Subscription, error) {

	var instanceRule []interface{}
	for _, instanceItem := range instance {
		instanceRule = append(instanceRule, instanceItem)
	}

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "EmergencyUnpauseInstance", instanceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryEmergencyUnpauseInstance)
				if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyUnpauseInstance", log); err != nil {
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
func (_InvisiFactory *InvisiFactoryFilterer) ParseEmergencyUnpauseInstance(log types.Log) (*InvisiFactoryEmergencyUnpauseInstance, error) {
	event := new(InvisiFactoryEmergencyUnpauseInstance)
	if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyUnpauseInstance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiFactoryEmergencyUnpauseRangeIterator is returned from FilterEmergencyUnpauseRange and is used to iterate over the raw logs and unpacked data for EmergencyUnpauseRange events raised by the InvisiFactory contract.
type InvisiFactoryEmergencyUnpauseRangeIterator struct {
	Event *InvisiFactoryEmergencyUnpauseRange // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryEmergencyUnpauseRangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryEmergencyUnpauseRange)
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
		it.Event = new(InvisiFactoryEmergencyUnpauseRange)
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
func (it *InvisiFactoryEmergencyUnpauseRangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryEmergencyUnpauseRangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryEmergencyUnpauseRange represents a EmergencyUnpauseRange event raised by the InvisiFactory contract.
type InvisiFactoryEmergencyUnpauseRange struct {
	Start *big.Int
	End   *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterEmergencyUnpauseRange is a free log retrieval operation binding the contract event 0xe1137af9d2e3b90b122226a33ada87fc87a62da822a8774db7beb4e77b8741be.
//
// Solidity: event EmergencyUnpauseRange(uint256 start, uint256 end)
func (_InvisiFactory *InvisiFactoryFilterer) FilterEmergencyUnpauseRange(opts *bind.FilterOpts) (*InvisiFactoryEmergencyUnpauseRangeIterator, error) {

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "EmergencyUnpauseRange")
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryEmergencyUnpauseRangeIterator{contract: _InvisiFactory.contract, event: "EmergencyUnpauseRange", logs: logs, sub: sub}, nil
}

// WatchEmergencyUnpauseRange is a free log subscription operation binding the contract event 0xe1137af9d2e3b90b122226a33ada87fc87a62da822a8774db7beb4e77b8741be.
//
// Solidity: event EmergencyUnpauseRange(uint256 start, uint256 end)
func (_InvisiFactory *InvisiFactoryFilterer) WatchEmergencyUnpauseRange(opts *bind.WatchOpts, sink chan<- *InvisiFactoryEmergencyUnpauseRange) (event.Subscription, error) {

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "EmergencyUnpauseRange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryEmergencyUnpauseRange)
				if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyUnpauseRange", log); err != nil {
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
func (_InvisiFactory *InvisiFactoryFilterer) ParseEmergencyUnpauseRange(log types.Log) (*InvisiFactoryEmergencyUnpauseRange, error) {
	event := new(InvisiFactoryEmergencyUnpauseRange)
	if err := _InvisiFactory.contract.UnpackLog(event, "EmergencyUnpauseRange", log); err != nil {
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
	Salt         [32]byte
	Instance     common.Address
	InitialOwner common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterInstanceDeployed is a free log retrieval operation binding the contract event 0xd321169aa5f5945b176ea1ec4917461694c2aafd7343099fd7e15462591404fd.
//
// Solidity: event InstanceDeployed(bytes32 indexed salt, address instance, address initialOwner)
func (_InvisiFactory *InvisiFactoryFilterer) FilterInstanceDeployed(opts *bind.FilterOpts, salt [][32]byte) (*InvisiFactoryInstanceDeployedIterator, error) {

	var saltRule []interface{}
	for _, saltItem := range salt {
		saltRule = append(saltRule, saltItem)
	}

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "InstanceDeployed", saltRule)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryInstanceDeployedIterator{contract: _InvisiFactory.contract, event: "InstanceDeployed", logs: logs, sub: sub}, nil
}

// WatchInstanceDeployed is a free log subscription operation binding the contract event 0xd321169aa5f5945b176ea1ec4917461694c2aafd7343099fd7e15462591404fd.
//
// Solidity: event InstanceDeployed(bytes32 indexed salt, address instance, address initialOwner)
func (_InvisiFactory *InvisiFactoryFilterer) WatchInstanceDeployed(opts *bind.WatchOpts, sink chan<- *InvisiFactoryInstanceDeployed, salt [][32]byte) (event.Subscription, error) {

	var saltRule []interface{}
	for _, saltItem := range salt {
		saltRule = append(saltRule, saltItem)
	}

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "InstanceDeployed", saltRule)
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

// ParseInstanceDeployed is a log parse operation binding the contract event 0xd321169aa5f5945b176ea1ec4917461694c2aafd7343099fd7e15462591404fd.
//
// Solidity: event InstanceDeployed(bytes32 indexed salt, address instance, address initialOwner)
func (_InvisiFactory *InvisiFactoryFilterer) ParseInstanceDeployed(log types.Log) (*InvisiFactoryInstanceDeployed, error) {
	event := new(InvisiFactoryInstanceDeployed)
	if err := _InvisiFactory.contract.UnpackLog(event, "InstanceDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiFactoryOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the InvisiFactory contract.
type InvisiFactoryOwnershipTransferStartedIterator struct {
	Event *InvisiFactoryOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryOwnershipTransferStarted)
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
		it.Event = new(InvisiFactoryOwnershipTransferStarted)
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
func (it *InvisiFactoryOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the InvisiFactory contract.
type InvisiFactoryOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiFactory *InvisiFactoryFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiFactoryOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryOwnershipTransferStartedIterator{contract: _InvisiFactory.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiFactory *InvisiFactoryFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *InvisiFactoryOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryOwnershipTransferStarted)
				if err := _InvisiFactory.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_InvisiFactory *InvisiFactoryFilterer) ParseOwnershipTransferStarted(log types.Log) (*InvisiFactoryOwnershipTransferStarted, error) {
	event := new(InvisiFactoryOwnershipTransferStarted)
	if err := _InvisiFactory.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiFactoryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the InvisiFactory contract.
type InvisiFactoryOwnershipTransferredIterator struct {
	Event *InvisiFactoryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *InvisiFactoryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiFactoryOwnershipTransferred)
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
		it.Event = new(InvisiFactoryOwnershipTransferred)
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
func (it *InvisiFactoryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiFactoryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiFactoryOwnershipTransferred represents a OwnershipTransferred event raised by the InvisiFactory contract.
type InvisiFactoryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiFactory *InvisiFactoryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiFactoryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiFactory.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiFactoryOwnershipTransferredIterator{contract: _InvisiFactory.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiFactory *InvisiFactoryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *InvisiFactoryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiFactory.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiFactoryOwnershipTransferred)
				if err := _InvisiFactory.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_InvisiFactory *InvisiFactoryFilterer) ParseOwnershipTransferred(log types.Log) (*InvisiFactoryOwnershipTransferred, error) {
	event := new(InvisiFactoryOwnershipTransferred)
	if err := _InvisiFactory.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
