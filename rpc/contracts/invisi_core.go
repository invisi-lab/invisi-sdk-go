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

// IInvisiCoreCreateCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreCreateCreditParams = protocolcontracts.IInvisiCoreCreateCreditParams

// IInvisiCoreCreditDrawParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreCreditDrawParams = protocolcontracts.IInvisiCoreCreditDrawParams

// IInvisiCoreCreditGrantRef is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreCreditGrantRef = protocolcontracts.IInvisiCoreCreditGrantRef

// IInvisiCoreCreditPayer is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreCreditPayer = protocolcontracts.IInvisiCoreCreditPayer

// IInvisiCoreCrossTreeTransactParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreCrossTreeTransactParams = protocolcontracts.IInvisiCoreCrossTreeTransactParams

// IInvisiCoreDelivery is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreDelivery = protocolcontracts.IInvisiCoreDelivery

// IInvisiCoreFundingParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreFundingParams = protocolcontracts.IInvisiCoreFundingParams

// IInvisiCoreInputSource is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreInputSource = protocolcontracts.IInvisiCoreInputSource

// IInvisiCoreShieldOutput is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreShieldOutput = protocolcontracts.IInvisiCoreShieldOutput

// IInvisiCoreShieldParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreShieldParams = protocolcontracts.IInvisiCoreShieldParams

// IInvisiCoreTransactRequest is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreTransactRequest = protocolcontracts.IInvisiCoreTransactRequest

// IInvisiCoreUpdateCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreUpdateCreditParams = protocolcontracts.IInvisiCoreUpdateCreditParams

// IMetaTxTokenPermit is an auto generated low-level Go binding around an user-defined struct.
type IMetaTxTokenPermit = protocolcontracts.IMetaTxTokenPermit

// InstanceParams is an auto generated low-level Go binding around an user-defined struct.
type InstanceParams = protocolcontracts.InstanceParams

// InvisiCoreMetaData contains all meta data concerning the InvisiCore contract.
var InvisiCoreMetaData = &bind.MetaData{ABI: protocolcontracts.InvisiCoreMetaData.ABI}

// InvisiCoreABI is the input ABI used to generate the binding from.
// Deprecated: Use InvisiCoreMetaData.ABI instead.
var InvisiCoreABI = InvisiCoreMetaData.ABI

// InvisiCore is an auto generated Go binding around an Ethereum contract.
type InvisiCore struct {
	InvisiCoreCaller     // Read-only binding to the contract
	InvisiCoreTransactor // Write-only binding to the contract
	InvisiCoreFilterer   // Log filterer for contract events
}

// InvisiCoreCaller is an auto generated read-only Go binding around an Ethereum contract.
type InvisiCoreCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiCoreTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InvisiCoreTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiCoreFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InvisiCoreFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiCoreSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InvisiCoreSession struct {
	Contract     *InvisiCore       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// InvisiCoreCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InvisiCoreCallerSession struct {
	Contract *InvisiCoreCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// InvisiCoreTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InvisiCoreTransactorSession struct {
	Contract     *InvisiCoreTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// InvisiCoreRaw is an auto generated low-level Go binding around an Ethereum contract.
type InvisiCoreRaw struct {
	Contract *InvisiCore // Generic contract binding to access the raw methods on
}

// InvisiCoreCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InvisiCoreCallerRaw struct {
	Contract *InvisiCoreCaller // Generic read-only contract binding to access the raw methods on
}

// InvisiCoreTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InvisiCoreTransactorRaw struct {
	Contract *InvisiCoreTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInvisiCore creates a new instance of InvisiCore, bound to a specific deployed contract.
func NewInvisiCore(address common.Address, backend bind.ContractBackend) (*InvisiCore, error) {
	contract, err := bindInvisiCore(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InvisiCore{InvisiCoreCaller: InvisiCoreCaller{contract: contract}, InvisiCoreTransactor: InvisiCoreTransactor{contract: contract}, InvisiCoreFilterer: InvisiCoreFilterer{contract: contract}}, nil
}

// NewInvisiCoreCaller creates a new read-only instance of InvisiCore, bound to a specific deployed contract.
func NewInvisiCoreCaller(address common.Address, caller bind.ContractCaller) (*InvisiCoreCaller, error) {
	contract, err := bindInvisiCore(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCaller{contract: contract}, nil
}

// NewInvisiCoreTransactor creates a new write-only instance of InvisiCore, bound to a specific deployed contract.
func NewInvisiCoreTransactor(address common.Address, transactor bind.ContractTransactor) (*InvisiCoreTransactor, error) {
	contract, err := bindInvisiCore(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreTransactor{contract: contract}, nil
}

// NewInvisiCoreFilterer creates a new log filterer instance of InvisiCore, bound to a specific deployed contract.
func NewInvisiCoreFilterer(address common.Address, filterer bind.ContractFilterer) (*InvisiCoreFilterer, error) {
	contract, err := bindInvisiCore(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreFilterer{contract: contract}, nil
}

// bindInvisiCore binds a generic wrapper to an already deployed contract.
func bindInvisiCore(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InvisiCoreMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiCore *InvisiCoreRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiCore.Contract.InvisiCoreCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiCore *InvisiCoreRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.Contract.InvisiCoreTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiCore *InvisiCoreRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiCore.Contract.InvisiCoreTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiCore *InvisiCoreCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiCore.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiCore *InvisiCoreTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiCore *InvisiCoreTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiCore.Contract.contract.Transact(opts, method, params...)
}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) CHAINID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "CHAIN_ID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiCore *InvisiCoreSession) CHAINID() (*big.Int, error) {
	return _InvisiCore.Contract.CHAINID(&_InvisiCore.CallOpts)
}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) CHAINID() (*big.Int, error) {
	return _InvisiCore.Contract.CHAINID(&_InvisiCore.CallOpts)
}

// FUNDSHIELDTYPEHASH is a free data retrieval call binding the contract method 0xbe9f7d0e.
//
// Solidity: function FUND_SHIELD_TYPEHASH() view returns(bytes32)
func (_InvisiCore *InvisiCoreCaller) FUNDSHIELDTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "FUND_SHIELD_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// FUNDSHIELDTYPEHASH is a free data retrieval call binding the contract method 0xbe9f7d0e.
//
// Solidity: function FUND_SHIELD_TYPEHASH() view returns(bytes32)
func (_InvisiCore *InvisiCoreSession) FUNDSHIELDTYPEHASH() ([32]byte, error) {
	return _InvisiCore.Contract.FUNDSHIELDTYPEHASH(&_InvisiCore.CallOpts)
}

// FUNDSHIELDTYPEHASH is a free data retrieval call binding the contract method 0xbe9f7d0e.
//
// Solidity: function FUND_SHIELD_TYPEHASH() view returns(bytes32)
func (_InvisiCore *InvisiCoreCallerSession) FUNDSHIELDTYPEHASH() ([32]byte, error) {
	return _InvisiCore.Contract.FUNDSHIELDTYPEHASH(&_InvisiCore.CallOpts)
}

// FUNDTRANSACTTYPEHASH is a free data retrieval call binding the contract method 0x8ebb23e8.
//
// Solidity: function FUND_TRANSACT_TYPEHASH() view returns(bytes32)
func (_InvisiCore *InvisiCoreCaller) FUNDTRANSACTTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "FUND_TRANSACT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// FUNDTRANSACTTYPEHASH is a free data retrieval call binding the contract method 0x8ebb23e8.
//
// Solidity: function FUND_TRANSACT_TYPEHASH() view returns(bytes32)
func (_InvisiCore *InvisiCoreSession) FUNDTRANSACTTYPEHASH() ([32]byte, error) {
	return _InvisiCore.Contract.FUNDTRANSACTTYPEHASH(&_InvisiCore.CallOpts)
}

// FUNDTRANSACTTYPEHASH is a free data retrieval call binding the contract method 0x8ebb23e8.
//
// Solidity: function FUND_TRANSACT_TYPEHASH() view returns(bytes32)
func (_InvisiCore *InvisiCoreCallerSession) FUNDTRANSACTTYPEHASH() ([32]byte, error) {
	return _InvisiCore.Contract.FUNDTRANSACTTYPEHASH(&_InvisiCore.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) VERSION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiCore *InvisiCoreSession) VERSION() (*big.Int, error) {
	return _InvisiCore.Contract.VERSION(&_InvisiCore.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) VERSION() (*big.Int, error) {
	return _InvisiCore.Contract.VERSION(&_InvisiCore.CallOpts)
}

// ActiveNoteTreeId is a free data retrieval call binding the contract method 0x236c8dd6.
//
// Solidity: function activeNoteTreeId() view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) ActiveNoteTreeId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "activeNoteTreeId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ActiveNoteTreeId is a free data retrieval call binding the contract method 0x236c8dd6.
//
// Solidity: function activeNoteTreeId() view returns(uint256)
func (_InvisiCore *InvisiCoreSession) ActiveNoteTreeId() (*big.Int, error) {
	return _InvisiCore.Contract.ActiveNoteTreeId(&_InvisiCore.CallOpts)
}

// ActiveNoteTreeId is a free data retrieval call binding the contract method 0x236c8dd6.
//
// Solidity: function activeNoteTreeId() view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) ActiveNoteTreeId() (*big.Int, error) {
	return _InvisiCore.Contract.ActiveNoteTreeId(&_InvisiCore.CallOpts)
}

// AddressRegistry is a free data retrieval call binding the contract method 0xf3ad65f4.
//
// Solidity: function addressRegistry() view returns(address)
func (_InvisiCore *InvisiCoreCaller) AddressRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "addressRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddressRegistry is a free data retrieval call binding the contract method 0xf3ad65f4.
//
// Solidity: function addressRegistry() view returns(address)
func (_InvisiCore *InvisiCoreSession) AddressRegistry() (common.Address, error) {
	return _InvisiCore.Contract.AddressRegistry(&_InvisiCore.CallOpts)
}

// AddressRegistry is a free data retrieval call binding the contract method 0xf3ad65f4.
//
// Solidity: function addressRegistry() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) AddressRegistry() (common.Address, error) {
	return _InvisiCore.Contract.AddressRegistry(&_InvisiCore.CallOpts)
}

// BroadcasterFeeBalance is a free data retrieval call binding the contract method 0x0bc8d1bd.
//
// Solidity: function broadcasterFeeBalance(address token, address recipient) view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) BroadcasterFeeBalance(opts *bind.CallOpts, token common.Address, recipient common.Address) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "broadcasterFeeBalance", token, recipient)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BroadcasterFeeBalance is a free data retrieval call binding the contract method 0x0bc8d1bd.
//
// Solidity: function broadcasterFeeBalance(address token, address recipient) view returns(uint256)
func (_InvisiCore *InvisiCoreSession) BroadcasterFeeBalance(token common.Address, recipient common.Address) (*big.Int, error) {
	return _InvisiCore.Contract.BroadcasterFeeBalance(&_InvisiCore.CallOpts, token, recipient)
}

// BroadcasterFeeBalance is a free data retrieval call binding the contract method 0x0bc8d1bd.
//
// Solidity: function broadcasterFeeBalance(address token, address recipient) view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) BroadcasterFeeBalance(token common.Address, recipient common.Address) (*big.Int, error) {
	return _InvisiCore.Contract.BroadcasterFeeBalance(&_InvisiCore.CallOpts, token, recipient)
}

// CircuitPaused is a free data retrieval call binding the contract method 0x5c12671f.
//
// Solidity: function circuitPaused() view returns(bool)
func (_InvisiCore *InvisiCoreCaller) CircuitPaused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "circuitPaused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CircuitPaused is a free data retrieval call binding the contract method 0x5c12671f.
//
// Solidity: function circuitPaused() view returns(bool)
func (_InvisiCore *InvisiCoreSession) CircuitPaused() (bool, error) {
	return _InvisiCore.Contract.CircuitPaused(&_InvisiCore.CallOpts)
}

// CircuitPaused is a free data retrieval call binding the contract method 0x5c12671f.
//
// Solidity: function circuitPaused() view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) CircuitPaused() (bool, error) {
	return _InvisiCore.Contract.CircuitPaused(&_InvisiCore.CallOpts)
}

// CreditAllocation is a free data retrieval call binding the contract method 0x70a7588f.
//
// Solidity: function creditAllocation(address owner, address token) view returns(uint64 shard, uint16 slot)
func (_InvisiCore *InvisiCoreCaller) CreditAllocation(opts *bind.CallOpts, owner common.Address, token common.Address) (struct {
	Shard uint64
	Slot  uint16
}, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "creditAllocation", owner, token)

	outstruct := new(struct {
		Shard uint64
		Slot  uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Shard = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.Slot = *abi.ConvertType(out[1], new(uint16)).(*uint16)

	return *outstruct, err

}

// CreditAllocation is a free data retrieval call binding the contract method 0x70a7588f.
//
// Solidity: function creditAllocation(address owner, address token) view returns(uint64 shard, uint16 slot)
func (_InvisiCore *InvisiCoreSession) CreditAllocation(owner common.Address, token common.Address) (struct {
	Shard uint64
	Slot  uint16
}, error) {
	return _InvisiCore.Contract.CreditAllocation(&_InvisiCore.CallOpts, owner, token)
}

// CreditAllocation is a free data retrieval call binding the contract method 0x70a7588f.
//
// Solidity: function creditAllocation(address owner, address token) view returns(uint64 shard, uint16 slot)
func (_InvisiCore *InvisiCoreCallerSession) CreditAllocation(owner common.Address, token common.Address) (struct {
	Shard uint64
	Slot  uint16
}, error) {
	return _InvisiCore.Contract.CreditAllocation(&_InvisiCore.CallOpts, owner, token)
}

// CreditRevocationBitmap is a free data retrieval call binding the contract method 0xc91ae210.
//
// Solidity: function creditRevocationBitmap(address owner, address token, uint64 shard, uint8 word) view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) CreditRevocationBitmap(opts *bind.CallOpts, owner common.Address, token common.Address, shard uint64, word uint8) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "creditRevocationBitmap", owner, token, shard, word)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CreditRevocationBitmap is a free data retrieval call binding the contract method 0xc91ae210.
//
// Solidity: function creditRevocationBitmap(address owner, address token, uint64 shard, uint8 word) view returns(uint256)
func (_InvisiCore *InvisiCoreSession) CreditRevocationBitmap(owner common.Address, token common.Address, shard uint64, word uint8) (*big.Int, error) {
	return _InvisiCore.Contract.CreditRevocationBitmap(&_InvisiCore.CallOpts, owner, token, shard, word)
}

// CreditRevocationBitmap is a free data retrieval call binding the contract method 0xc91ae210.
//
// Solidity: function creditRevocationBitmap(address owner, address token, uint64 shard, uint8 word) view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) CreditRevocationBitmap(owner common.Address, token common.Address, shard uint64, word uint8) (*big.Int, error) {
	return _InvisiCore.Contract.CreditRevocationBitmap(&_InvisiCore.CallOpts, owner, token, shard, word)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiCore *InvisiCoreCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiCore *InvisiCoreSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _InvisiCore.Contract.Eip712Domain(&_InvisiCore.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiCore *InvisiCoreCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _InvisiCore.Contract.Eip712Domain(&_InvisiCore.CallOpts)
}

// EmergencyAdmin is a free data retrieval call binding the contract method 0x70905dce.
//
// Solidity: function emergencyAdmin() view returns(address)
func (_InvisiCore *InvisiCoreCaller) EmergencyAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "emergencyAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EmergencyAdmin is a free data retrieval call binding the contract method 0x70905dce.
//
// Solidity: function emergencyAdmin() view returns(address)
func (_InvisiCore *InvisiCoreSession) EmergencyAdmin() (common.Address, error) {
	return _InvisiCore.Contract.EmergencyAdmin(&_InvisiCore.CallOpts)
}

// EmergencyAdmin is a free data retrieval call binding the contract method 0x70905dce.
//
// Solidity: function emergencyAdmin() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) EmergencyAdmin() (common.Address, error) {
	return _InvisiCore.Contract.EmergencyAdmin(&_InvisiCore.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc45a0155.
//
// Solidity: function factory() view returns(address)
func (_InvisiCore *InvisiCoreCaller) Factory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "factory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Factory is a free data retrieval call binding the contract method 0xc45a0155.
//
// Solidity: function factory() view returns(address)
func (_InvisiCore *InvisiCoreSession) Factory() (common.Address, error) {
	return _InvisiCore.Contract.Factory(&_InvisiCore.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc45a0155.
//
// Solidity: function factory() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) Factory() (common.Address, error) {
	return _InvisiCore.Contract.Factory(&_InvisiCore.CallOpts)
}

// GetNoteTreeNextLeafIndex is a free data retrieval call binding the contract method 0xb7b81f59.
//
// Solidity: function getNoteTreeNextLeafIndex(uint256 treeId) view returns(uint32)
func (_InvisiCore *InvisiCoreCaller) GetNoteTreeNextLeafIndex(opts *bind.CallOpts, treeId *big.Int) (uint32, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "getNoteTreeNextLeafIndex", treeId)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetNoteTreeNextLeafIndex is a free data retrieval call binding the contract method 0xb7b81f59.
//
// Solidity: function getNoteTreeNextLeafIndex(uint256 treeId) view returns(uint32)
func (_InvisiCore *InvisiCoreSession) GetNoteTreeNextLeafIndex(treeId *big.Int) (uint32, error) {
	return _InvisiCore.Contract.GetNoteTreeNextLeafIndex(&_InvisiCore.CallOpts, treeId)
}

// GetNoteTreeNextLeafIndex is a free data retrieval call binding the contract method 0xb7b81f59.
//
// Solidity: function getNoteTreeNextLeafIndex(uint256 treeId) view returns(uint32)
func (_InvisiCore *InvisiCoreCallerSession) GetNoteTreeNextLeafIndex(treeId *big.Int) (uint32, error) {
	return _InvisiCore.Contract.GetNoteTreeNextLeafIndex(&_InvisiCore.CallOpts, treeId)
}

// GetNoteTreeRoot is a free data retrieval call binding the contract method 0xd7793a38.
//
// Solidity: function getNoteTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiCore *InvisiCoreCaller) GetNoteTreeRoot(opts *bind.CallOpts, treeId *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "getNoteTreeRoot", treeId)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetNoteTreeRoot is a free data retrieval call binding the contract method 0xd7793a38.
//
// Solidity: function getNoteTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiCore *InvisiCoreSession) GetNoteTreeRoot(treeId *big.Int) ([32]byte, error) {
	return _InvisiCore.Contract.GetNoteTreeRoot(&_InvisiCore.CallOpts, treeId)
}

// GetNoteTreeRoot is a free data retrieval call binding the contract method 0xd7793a38.
//
// Solidity: function getNoteTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiCore *InvisiCoreCallerSession) GetNoteTreeRoot(treeId *big.Int) ([32]byte, error) {
	return _InvisiCore.Contract.GetNoteTreeRoot(&_InvisiCore.CallOpts, treeId)
}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256 count)
func (_InvisiCore *InvisiCoreCaller) GetTreeCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "getTreeCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256 count)
func (_InvisiCore *InvisiCoreSession) GetTreeCount() (*big.Int, error) {
	return _InvisiCore.Contract.GetTreeCount(&_InvisiCore.CallOpts)
}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256 count)
func (_InvisiCore *InvisiCoreCallerSession) GetTreeCount() (*big.Int, error) {
	return _InvisiCore.Contract.GetTreeCount(&_InvisiCore.CallOpts)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiCore *InvisiCoreCaller) GlobalEmergency(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "globalEmergency")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiCore *InvisiCoreSession) GlobalEmergency() (common.Address, error) {
	return _InvisiCore.Contract.GlobalEmergency(&_InvisiCore.CallOpts)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) GlobalEmergency() (common.Address, error) {
	return _InvisiCore.Contract.GlobalEmergency(&_InvisiCore.CallOpts)
}

// IsCommitmentRegistered is a free data retrieval call binding the contract method 0x620b5305.
//
// Solidity: function isCommitmentRegistered(bytes32 commitment) view returns(bool)
func (_InvisiCore *InvisiCoreCaller) IsCommitmentRegistered(opts *bind.CallOpts, commitment [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "isCommitmentRegistered", commitment)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCommitmentRegistered is a free data retrieval call binding the contract method 0x620b5305.
//
// Solidity: function isCommitmentRegistered(bytes32 commitment) view returns(bool)
func (_InvisiCore *InvisiCoreSession) IsCommitmentRegistered(commitment [32]byte) (bool, error) {
	return _InvisiCore.Contract.IsCommitmentRegistered(&_InvisiCore.CallOpts, commitment)
}

// IsCommitmentRegistered is a free data retrieval call binding the contract method 0x620b5305.
//
// Solidity: function isCommitmentRegistered(bytes32 commitment) view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) IsCommitmentRegistered(commitment [32]byte) (bool, error) {
	return _InvisiCore.Contract.IsCommitmentRegistered(&_InvisiCore.CallOpts, commitment)
}

// IsNullifierSpent is a free data retrieval call binding the contract method 0xd5a4e325.
//
// Solidity: function isNullifierSpent(bytes32 nullifier) view returns(bool)
func (_InvisiCore *InvisiCoreCaller) IsNullifierSpent(opts *bind.CallOpts, nullifier [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "isNullifierSpent", nullifier)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsNullifierSpent is a free data retrieval call binding the contract method 0xd5a4e325.
//
// Solidity: function isNullifierSpent(bytes32 nullifier) view returns(bool)
func (_InvisiCore *InvisiCoreSession) IsNullifierSpent(nullifier [32]byte) (bool, error) {
	return _InvisiCore.Contract.IsNullifierSpent(&_InvisiCore.CallOpts, nullifier)
}

// IsNullifierSpent is a free data retrieval call binding the contract method 0xd5a4e325.
//
// Solidity: function isNullifierSpent(bytes32 nullifier) view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) IsNullifierSpent(nullifier [32]byte) (bool, error) {
	return _InvisiCore.Contract.IsNullifierSpent(&_InvisiCore.CallOpts, nullifier)
}

// IsTreeFull is a free data retrieval call binding the contract method 0x9fc2199b.
//
// Solidity: function isTreeFull(uint256 treeId) view returns(bool)
func (_InvisiCore *InvisiCoreCaller) IsTreeFull(opts *bind.CallOpts, treeId *big.Int) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "isTreeFull", treeId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTreeFull is a free data retrieval call binding the contract method 0x9fc2199b.
//
// Solidity: function isTreeFull(uint256 treeId) view returns(bool)
func (_InvisiCore *InvisiCoreSession) IsTreeFull(treeId *big.Int) (bool, error) {
	return _InvisiCore.Contract.IsTreeFull(&_InvisiCore.CallOpts, treeId)
}

// IsTreeFull is a free data retrieval call binding the contract method 0x9fc2199b.
//
// Solidity: function isTreeFull(uint256 treeId) view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) IsTreeFull(treeId *big.Int) (bool, error) {
	return _InvisiCore.Contract.IsTreeFull(&_InvisiCore.CallOpts, treeId)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) Nonces(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "nonces", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_InvisiCore *InvisiCoreSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _InvisiCore.Contract.Nonces(&_InvisiCore.CallOpts, arg0)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _InvisiCore.Contract.Nonces(&_InvisiCore.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiCore *InvisiCoreCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiCore *InvisiCoreSession) Owner() (common.Address, error) {
	return _InvisiCore.Contract.Owner(&_InvisiCore.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) Owner() (common.Address, error) {
	return _InvisiCore.Contract.Owner(&_InvisiCore.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_InvisiCore *InvisiCoreCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_InvisiCore *InvisiCoreSession) Paused() (bool, error) {
	return _InvisiCore.Contract.Paused(&_InvisiCore.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) Paused() (bool, error) {
	return _InvisiCore.Contract.Paused(&_InvisiCore.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiCore *InvisiCoreCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiCore *InvisiCoreSession) PendingOwner() (common.Address, error) {
	return _InvisiCore.Contract.PendingOwner(&_InvisiCore.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) PendingOwner() (common.Address, error) {
	return _InvisiCore.Contract.PendingOwner(&_InvisiCore.CallOpts)
}

// RootHistory is a free data retrieval call binding the contract method 0xc718dbda.
//
// Solidity: function rootHistory(uint256 , bytes32 ) view returns(bool)
func (_InvisiCore *InvisiCoreCaller) RootHistory(opts *bind.CallOpts, arg0 *big.Int, arg1 [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "rootHistory", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RootHistory is a free data retrieval call binding the contract method 0xc718dbda.
//
// Solidity: function rootHistory(uint256 , bytes32 ) view returns(bool)
func (_InvisiCore *InvisiCoreSession) RootHistory(arg0 *big.Int, arg1 [32]byte) (bool, error) {
	return _InvisiCore.Contract.RootHistory(&_InvisiCore.CallOpts, arg0, arg1)
}

// RootHistory is a free data retrieval call binding the contract method 0xc718dbda.
//
// Solidity: function rootHistory(uint256 , bytes32 ) view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) RootHistory(arg0 *big.Int, arg1 [32]byte) (bool, error) {
	return _InvisiCore.Contract.RootHistory(&_InvisiCore.CallOpts, arg0, arg1)
}

// Screener is a free data retrieval call binding the contract method 0x6e0fec70.
//
// Solidity: function screener() view returns(address)
func (_InvisiCore *InvisiCoreCaller) Screener(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "screener")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Screener is a free data retrieval call binding the contract method 0x6e0fec70.
//
// Solidity: function screener() view returns(address)
func (_InvisiCore *InvisiCoreSession) Screener() (common.Address, error) {
	return _InvisiCore.Contract.Screener(&_InvisiCore.CallOpts)
}

// Screener is a free data retrieval call binding the contract method 0x6e0fec70.
//
// Solidity: function screener() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) Screener() (common.Address, error) {
	return _InvisiCore.Contract.Screener(&_InvisiCore.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiCore *InvisiCoreCaller) VerifierRouter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "verifierRouter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiCore *InvisiCoreSession) VerifierRouter() (common.Address, error) {
	return _InvisiCore.Contract.VerifierRouter(&_InvisiCore.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiCore *InvisiCoreCallerSession) VerifierRouter() (common.Address, error) {
	return _InvisiCore.Contract.VerifierRouter(&_InvisiCore.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiCore *InvisiCoreTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiCore *InvisiCoreSession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiCore.Contract.AcceptOwnership(&_InvisiCore.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiCore *InvisiCoreTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiCore.Contract.AcceptOwnership(&_InvisiCore.TransactOpts)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x5550f27a.
//
// Solidity: function createCredit(bytes , (address,address,uint64,uint16,bytes32,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactor) CreateCredit(opts *bind.TransactOpts, arg0 []byte, p IInvisiCoreCreateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "createCredit", arg0, p, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x5550f27a.
//
// Solidity: function createCredit(bytes , (address,address,uint64,uint16,bytes32,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreSession) CreateCredit(arg0 []byte, p IInvisiCoreCreateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x5550f27a.
//
// Solidity: function createCredit(bytes , (address,address,uint64,uint16,bytes32,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactorSession) CreateCredit(arg0 []byte, p IInvisiCoreCreateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa288961c.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactor) CreditDraw(opts *bind.TransactOpts, arg0 []byte, p IInvisiCoreCreditDrawParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "creditDraw", arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa288961c.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreSession) CreditDraw(arg0 []byte, p IInvisiCoreCreditDrawParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreditDraw(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa288961c.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactorSession) CreditDraw(arg0 []byte, p IInvisiCoreCreditDrawParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreditDraw(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// InvalidateNonce is a paid mutator transaction binding the contract method 0x5a57b46f.
//
// Solidity: function invalidateNonce() returns()
func (_InvisiCore *InvisiCoreTransactor) InvalidateNonce(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "invalidateNonce")
}

// InvalidateNonce is a paid mutator transaction binding the contract method 0x5a57b46f.
//
// Solidity: function invalidateNonce() returns()
func (_InvisiCore *InvisiCoreSession) InvalidateNonce() (*types.Transaction, error) {
	return _InvisiCore.Contract.InvalidateNonce(&_InvisiCore.TransactOpts)
}

// InvalidateNonce is a paid mutator transaction binding the contract method 0x5a57b46f.
//
// Solidity: function invalidateNonce() returns()
func (_InvisiCore *InvisiCoreTransactorSession) InvalidateNonce() (*types.Transaction, error) {
	return _InvisiCore.Contract.InvalidateNonce(&_InvisiCore.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_InvisiCore *InvisiCoreTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_InvisiCore *InvisiCoreSession) Pause() (*types.Transaction, error) {
	return _InvisiCore.Contract.Pause(&_InvisiCore.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_InvisiCore *InvisiCoreTransactorSession) Pause() (*types.Transaction, error) {
	return _InvisiCore.Contract.Pause(&_InvisiCore.TransactOpts)
}

// PauseCircuit is a paid mutator transaction binding the contract method 0x50424cd0.
//
// Solidity: function pauseCircuit() returns()
func (_InvisiCore *InvisiCoreTransactor) PauseCircuit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "pauseCircuit")
}

// PauseCircuit is a paid mutator transaction binding the contract method 0x50424cd0.
//
// Solidity: function pauseCircuit() returns()
func (_InvisiCore *InvisiCoreSession) PauseCircuit() (*types.Transaction, error) {
	return _InvisiCore.Contract.PauseCircuit(&_InvisiCore.TransactOpts)
}

// PauseCircuit is a paid mutator transaction binding the contract method 0x50424cd0.
//
// Solidity: function pauseCircuit() returns()
func (_InvisiCore *InvisiCoreTransactorSession) PauseCircuit() (*types.Transaction, error) {
	return _InvisiCore.Contract.PauseCircuit(&_InvisiCore.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiCore *InvisiCoreTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiCore *InvisiCoreSession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiCore.Contract.RenounceOwnership(&_InvisiCore.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiCore *InvisiCoreTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiCore.Contract.RenounceOwnership(&_InvisiCore.TransactOpts)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x400b3547.
//
// Solidity: function revokeCredit(address token, address owner, uint64 shard, uint16 slot, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiCore *InvisiCoreTransactor) RevokeCredit(opts *bind.TransactOpts, token common.Address, owner common.Address, shard uint64, slot uint16, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "revokeCredit", token, owner, shard, slot, deadline, feeRecipient, feeAmount, signature)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x400b3547.
//
// Solidity: function revokeCredit(address token, address owner, uint64 shard, uint16 slot, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiCore *InvisiCoreSession) RevokeCredit(token common.Address, owner common.Address, shard uint64, slot uint16, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiCore.Contract.RevokeCredit(&_InvisiCore.TransactOpts, token, owner, shard, slot, deadline, feeRecipient, feeAmount, signature)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x400b3547.
//
// Solidity: function revokeCredit(address token, address owner, uint64 shard, uint16 slot, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiCore *InvisiCoreTransactorSession) RevokeCredit(token common.Address, owner common.Address, shard uint64, slot uint16, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiCore.Contract.RevokeCredit(&_InvisiCore.TransactOpts, token, owner, shard, slot, deadline, feeRecipient, feeAmount, signature)
}

// SetScreener is a paid mutator transaction binding the contract method 0xc2ebd23d.
//
// Solidity: function setScreener(address newScreener) returns()
func (_InvisiCore *InvisiCoreTransactor) SetScreener(opts *bind.TransactOpts, newScreener common.Address) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "setScreener", newScreener)
}

// SetScreener is a paid mutator transaction binding the contract method 0xc2ebd23d.
//
// Solidity: function setScreener(address newScreener) returns()
func (_InvisiCore *InvisiCoreSession) SetScreener(newScreener common.Address) (*types.Transaction, error) {
	return _InvisiCore.Contract.SetScreener(&_InvisiCore.TransactOpts, newScreener)
}

// SetScreener is a paid mutator transaction binding the contract method 0xc2ebd23d.
//
// Solidity: function setScreener(address newScreener) returns()
func (_InvisiCore *InvisiCoreTransactorSession) SetScreener(newScreener common.Address) (*types.Transaction, error) {
	return _InvisiCore.Contract.SetScreener(&_InvisiCore.TransactOpts, newScreener)
}

// Shield is a paid mutator transaction binding the contract method 0xae897f6a.
//
// Solidity: function shield((address,bytes32[],uint256,address,uint256) p, (bytes32,uint256)[] , (bytes,bytes)[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiCore *InvisiCoreTransactor) Shield(opts *bind.TransactOpts, p IInvisiCoreShieldParams, arg1 []IInvisiCoreShieldOutput, arg2 []IInvisiCoreDelivery, arg3 IInvisiCoreFundingParams) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "shield", p, arg1, arg2, arg3)
}

// Shield is a paid mutator transaction binding the contract method 0xae897f6a.
//
// Solidity: function shield((address,bytes32[],uint256,address,uint256) p, (bytes32,uint256)[] , (bytes,bytes)[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiCore *InvisiCoreSession) Shield(p IInvisiCoreShieldParams, arg1 []IInvisiCoreShieldOutput, arg2 []IInvisiCoreDelivery, arg3 IInvisiCoreFundingParams) (*types.Transaction, error) {
	return _InvisiCore.Contract.Shield(&_InvisiCore.TransactOpts, p, arg1, arg2, arg3)
}

// Shield is a paid mutator transaction binding the contract method 0xae897f6a.
//
// Solidity: function shield((address,bytes32[],uint256,address,uint256) p, (bytes32,uint256)[] , (bytes,bytes)[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiCore *InvisiCoreTransactorSession) Shield(p IInvisiCoreShieldParams, arg1 []IInvisiCoreShieldOutput, arg2 []IInvisiCoreDelivery, arg3 IInvisiCoreFundingParams) (*types.Transaction, error) {
	return _InvisiCore.Contract.Shield(&_InvisiCore.TransactOpts, p, arg1, arg2, arg3)
}

// Transact is a paid mutator transaction binding the contract method 0x2eb90f9b.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) r) returns()
func (_InvisiCore *InvisiCoreTransactor) Transact(opts *bind.TransactOpts, arg0 []byte, r IInvisiCoreTransactRequest) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "transact", arg0, r)
}

// Transact is a paid mutator transaction binding the contract method 0x2eb90f9b.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) r) returns()
func (_InvisiCore *InvisiCoreSession) Transact(arg0 []byte, r IInvisiCoreTransactRequest) (*types.Transaction, error) {
	return _InvisiCore.Contract.Transact(&_InvisiCore.TransactOpts, arg0, r)
}

// Transact is a paid mutator transaction binding the contract method 0x2eb90f9b.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) r) returns()
func (_InvisiCore *InvisiCoreTransactorSession) Transact(arg0 []byte, r IInvisiCoreTransactRequest) (*types.Transaction, error) {
	return _InvisiCore.Contract.Transact(&_InvisiCore.TransactOpts, arg0, r)
}

// Transact0 is a paid mutator transaction binding the contract method 0x870949a9.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] requests) returns()
func (_InvisiCore *InvisiCoreTransactor) Transact0(opts *bind.TransactOpts, arg0 []byte, requests []IInvisiCoreTransactRequest) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "transact0", arg0, requests)
}

// Transact0 is a paid mutator transaction binding the contract method 0x870949a9.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] requests) returns()
func (_InvisiCore *InvisiCoreSession) Transact0(arg0 []byte, requests []IInvisiCoreTransactRequest) (*types.Transaction, error) {
	return _InvisiCore.Contract.Transact0(&_InvisiCore.TransactOpts, arg0, requests)
}

// Transact0 is a paid mutator transaction binding the contract method 0x870949a9.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] requests) returns()
func (_InvisiCore *InvisiCoreTransactorSession) Transact0(arg0 []byte, requests []IInvisiCoreTransactRequest) (*types.Transaction, error) {
	return _InvisiCore.Contract.Transact0(&_InvisiCore.TransactOpts, arg0, requests)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiCore *InvisiCoreTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiCore *InvisiCoreSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiCore.Contract.TransferOwnership(&_InvisiCore.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiCore *InvisiCoreTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiCore.Contract.TransferOwnership(&_InvisiCore.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_InvisiCore *InvisiCoreTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_InvisiCore *InvisiCoreSession) Unpause() (*types.Transaction, error) {
	return _InvisiCore.Contract.Unpause(&_InvisiCore.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_InvisiCore *InvisiCoreTransactorSession) Unpause() (*types.Transaction, error) {
	return _InvisiCore.Contract.Unpause(&_InvisiCore.TransactOpts)
}

// UnpauseCircuit is a paid mutator transaction binding the contract method 0xa3bf4880.
//
// Solidity: function unpauseCircuit() returns()
func (_InvisiCore *InvisiCoreTransactor) UnpauseCircuit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "unpauseCircuit")
}

// UnpauseCircuit is a paid mutator transaction binding the contract method 0xa3bf4880.
//
// Solidity: function unpauseCircuit() returns()
func (_InvisiCore *InvisiCoreSession) UnpauseCircuit() (*types.Transaction, error) {
	return _InvisiCore.Contract.UnpauseCircuit(&_InvisiCore.TransactOpts)
}

// UnpauseCircuit is a paid mutator transaction binding the contract method 0xa3bf4880.
//
// Solidity: function unpauseCircuit() returns()
func (_InvisiCore *InvisiCoreTransactorSession) UnpauseCircuit() (*types.Transaction, error) {
	return _InvisiCore.Contract.UnpauseCircuit(&_InvisiCore.TransactOpts)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0x2dab44e5.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactor) UpdateCredit(opts *bind.TransactOpts, arg0 []byte, p IInvisiCoreUpdateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "updateCredit", arg0, p, arg2)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0x2dab44e5.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreSession) UpdateCredit(arg0 []byte, p IInvisiCoreUpdateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.UpdateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0x2dab44e5.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactorSession) UpdateCredit(arg0 []byte, p IInvisiCoreUpdateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.UpdateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// WithdrawBroadcasterFees is a paid mutator transaction binding the contract method 0x099e7064.
//
// Solidity: function withdrawBroadcasterFees(address token, address recipient) returns(uint256)
func (_InvisiCore *InvisiCoreTransactor) WithdrawBroadcasterFees(opts *bind.TransactOpts, token common.Address, recipient common.Address) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "withdrawBroadcasterFees", token, recipient)
}

// WithdrawBroadcasterFees is a paid mutator transaction binding the contract method 0x099e7064.
//
// Solidity: function withdrawBroadcasterFees(address token, address recipient) returns(uint256)
func (_InvisiCore *InvisiCoreSession) WithdrawBroadcasterFees(token common.Address, recipient common.Address) (*types.Transaction, error) {
	return _InvisiCore.Contract.WithdrawBroadcasterFees(&_InvisiCore.TransactOpts, token, recipient)
}

// WithdrawBroadcasterFees is a paid mutator transaction binding the contract method 0x099e7064.
//
// Solidity: function withdrawBroadcasterFees(address token, address recipient) returns(uint256)
func (_InvisiCore *InvisiCoreTransactorSession) WithdrawBroadcasterFees(token common.Address, recipient common.Address) (*types.Transaction, error) {
	return _InvisiCore.Contract.WithdrawBroadcasterFees(&_InvisiCore.TransactOpts, token, recipient)
}

// InvisiCoreBatchOutputsIterator is returned from FilterBatchOutputs and is used to iterate over the raw logs and unpacked data for BatchOutputs events raised by the InvisiCore contract.
type InvisiCoreBatchOutputsIterator struct {
	Event *InvisiCoreBatchOutputs // Event containing the contract specifics and raw log

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
func (it *InvisiCoreBatchOutputsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreBatchOutputs)
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
		it.Event = new(InvisiCoreBatchOutputs)
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
func (it *InvisiCoreBatchOutputsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreBatchOutputsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreBatchOutputs represents a BatchOutputs event raised by the InvisiCore contract.
type InvisiCoreBatchOutputs struct {
	TreeId            *big.Int
	StartLeafIndex    uint32
	FirstOutputOffset *big.Int
	Tokens            []common.Address
	NoteCommitments   [][32]byte
	PackedDeliveries  []byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterBatchOutputs is a free log retrieval operation binding the contract event 0xf5c3c1e590ac45e708a69991fe84df4bb1dc4c185bf0a1ec6794c31b776a1be0.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, address[] tokens, bytes32[] noteCommitments, bytes packedDeliveries)
func (_InvisiCore *InvisiCoreFilterer) FilterBatchOutputs(opts *bind.FilterOpts, treeId []*big.Int) (*InvisiCoreBatchOutputsIterator, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "BatchOutputs", treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreBatchOutputsIterator{contract: _InvisiCore.contract, event: "BatchOutputs", logs: logs, sub: sub}, nil
}

// WatchBatchOutputs is a free log subscription operation binding the contract event 0xf5c3c1e590ac45e708a69991fe84df4bb1dc4c185bf0a1ec6794c31b776a1be0.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, address[] tokens, bytes32[] noteCommitments, bytes packedDeliveries)
func (_InvisiCore *InvisiCoreFilterer) WatchBatchOutputs(opts *bind.WatchOpts, sink chan<- *InvisiCoreBatchOutputs, treeId []*big.Int) (event.Subscription, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "BatchOutputs", treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreBatchOutputs)
				if err := _InvisiCore.contract.UnpackLog(event, "BatchOutputs", log); err != nil {
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

// ParseBatchOutputs is a log parse operation binding the contract event 0xf5c3c1e590ac45e708a69991fe84df4bb1dc4c185bf0a1ec6794c31b776a1be0.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, address[] tokens, bytes32[] noteCommitments, bytes packedDeliveries)
func (_InvisiCore *InvisiCoreFilterer) ParseBatchOutputs(log types.Log) (*InvisiCoreBatchOutputs, error) {
	event := new(InvisiCoreBatchOutputs)
	if err := _InvisiCore.contract.UnpackLog(event, "BatchOutputs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreBatchRequestIterator is returned from FilterBatchRequest and is used to iterate over the raw logs and unpacked data for BatchRequest events raised by the InvisiCore contract.
type InvisiCoreBatchRequestIterator struct {
	Event *InvisiCoreBatchRequest // Event containing the contract specifics and raw log

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
func (it *InvisiCoreBatchRequestIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreBatchRequest)
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
		it.Event = new(InvisiCoreBatchRequest)
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
func (it *InvisiCoreBatchRequestIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreBatchRequestIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreBatchRequest represents a BatchRequest event raised by the InvisiCore contract.
type InvisiCoreBatchRequest struct {
	Token             common.Address
	RequestIndex      *big.Int
	Capacity          *big.Int
	Sources           []IInvisiCoreInputSource
	Nullifiers        [][32]byte
	OutputCount       *big.Int
	Payer             common.Address
	DepositAmount     *big.Int
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	StatementHash     [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterBatchRequest is a free log retrieval operation binding the contract event 0xbceb7ab95985373d11b26c1454947414ae0b754a3b33e0f3c0d779bf540f9db5.
//
// Solidity: event BatchRequest(address token, uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiCore *InvisiCoreFilterer) FilterBatchRequest(opts *bind.FilterOpts, requestIndex []*big.Int) (*InvisiCoreBatchRequestIterator, error) {

	var requestIndexRule []interface{}
	for _, requestIndexItem := range requestIndex {
		requestIndexRule = append(requestIndexRule, requestIndexItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "BatchRequest", requestIndexRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreBatchRequestIterator{contract: _InvisiCore.contract, event: "BatchRequest", logs: logs, sub: sub}, nil
}

// WatchBatchRequest is a free log subscription operation binding the contract event 0xbceb7ab95985373d11b26c1454947414ae0b754a3b33e0f3c0d779bf540f9db5.
//
// Solidity: event BatchRequest(address token, uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiCore *InvisiCoreFilterer) WatchBatchRequest(opts *bind.WatchOpts, sink chan<- *InvisiCoreBatchRequest, requestIndex []*big.Int) (event.Subscription, error) {

	var requestIndexRule []interface{}
	for _, requestIndexItem := range requestIndex {
		requestIndexRule = append(requestIndexRule, requestIndexItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "BatchRequest", requestIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreBatchRequest)
				if err := _InvisiCore.contract.UnpackLog(event, "BatchRequest", log); err != nil {
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

// ParseBatchRequest is a log parse operation binding the contract event 0xbceb7ab95985373d11b26c1454947414ae0b754a3b33e0f3c0d779bf540f9db5.
//
// Solidity: event BatchRequest(address token, uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiCore *InvisiCoreFilterer) ParseBatchRequest(log types.Log) (*InvisiCoreBatchRequest, error) {
	event := new(InvisiCoreBatchRequest)
	if err := _InvisiCore.contract.UnpackLog(event, "BatchRequest", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreBroadcasterFeeIterator is returned from FilterBroadcasterFee and is used to iterate over the raw logs and unpacked data for BroadcasterFee events raised by the InvisiCore contract.
type InvisiCoreBroadcasterFeeIterator struct {
	Event *InvisiCoreBroadcasterFee // Event containing the contract specifics and raw log

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
func (it *InvisiCoreBroadcasterFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreBroadcasterFee)
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
		it.Event = new(InvisiCoreBroadcasterFee)
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
func (it *InvisiCoreBroadcasterFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreBroadcasterFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreBroadcasterFee represents a BroadcasterFee event raised by the InvisiCore contract.
type InvisiCoreBroadcasterFee struct {
	Token     common.Address
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBroadcasterFee is a free log retrieval operation binding the contract event 0xe5428e81365966d5479efecda1c1de8cf8624f59d57290a0501f499baeb281aa.
//
// Solidity: event BroadcasterFee(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) FilterBroadcasterFee(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*InvisiCoreBroadcasterFeeIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "BroadcasterFee", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreBroadcasterFeeIterator{contract: _InvisiCore.contract, event: "BroadcasterFee", logs: logs, sub: sub}, nil
}

// WatchBroadcasterFee is a free log subscription operation binding the contract event 0xe5428e81365966d5479efecda1c1de8cf8624f59d57290a0501f499baeb281aa.
//
// Solidity: event BroadcasterFee(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) WatchBroadcasterFee(opts *bind.WatchOpts, sink chan<- *InvisiCoreBroadcasterFee, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "BroadcasterFee", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreBroadcasterFee)
				if err := _InvisiCore.contract.UnpackLog(event, "BroadcasterFee", log); err != nil {
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

// ParseBroadcasterFee is a log parse operation binding the contract event 0xe5428e81365966d5479efecda1c1de8cf8624f59d57290a0501f499baeb281aa.
//
// Solidity: event BroadcasterFee(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) ParseBroadcasterFee(log types.Log) (*InvisiCoreBroadcasterFee, error) {
	event := new(InvisiCoreBroadcasterFee)
	if err := _InvisiCore.contract.UnpackLog(event, "BroadcasterFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreBroadcasterFeeWithdrawnIterator is returned from FilterBroadcasterFeeWithdrawn and is used to iterate over the raw logs and unpacked data for BroadcasterFeeWithdrawn events raised by the InvisiCore contract.
type InvisiCoreBroadcasterFeeWithdrawnIterator struct {
	Event *InvisiCoreBroadcasterFeeWithdrawn // Event containing the contract specifics and raw log

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
func (it *InvisiCoreBroadcasterFeeWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreBroadcasterFeeWithdrawn)
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
		it.Event = new(InvisiCoreBroadcasterFeeWithdrawn)
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
func (it *InvisiCoreBroadcasterFeeWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreBroadcasterFeeWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreBroadcasterFeeWithdrawn represents a BroadcasterFeeWithdrawn event raised by the InvisiCore contract.
type InvisiCoreBroadcasterFeeWithdrawn struct {
	Token     common.Address
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBroadcasterFeeWithdrawn is a free log retrieval operation binding the contract event 0x1b7af30e773fc72482ff43806732398e343a2e6a689255c8f9236871c4128139.
//
// Solidity: event BroadcasterFeeWithdrawn(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) FilterBroadcasterFeeWithdrawn(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*InvisiCoreBroadcasterFeeWithdrawnIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "BroadcasterFeeWithdrawn", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreBroadcasterFeeWithdrawnIterator{contract: _InvisiCore.contract, event: "BroadcasterFeeWithdrawn", logs: logs, sub: sub}, nil
}

// WatchBroadcasterFeeWithdrawn is a free log subscription operation binding the contract event 0x1b7af30e773fc72482ff43806732398e343a2e6a689255c8f9236871c4128139.
//
// Solidity: event BroadcasterFeeWithdrawn(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) WatchBroadcasterFeeWithdrawn(opts *bind.WatchOpts, sink chan<- *InvisiCoreBroadcasterFeeWithdrawn, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "BroadcasterFeeWithdrawn", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreBroadcasterFeeWithdrawn)
				if err := _InvisiCore.contract.UnpackLog(event, "BroadcasterFeeWithdrawn", log); err != nil {
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

// ParseBroadcasterFeeWithdrawn is a log parse operation binding the contract event 0x1b7af30e773fc72482ff43806732398e343a2e6a689255c8f9236871c4128139.
//
// Solidity: event BroadcasterFeeWithdrawn(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) ParseBroadcasterFeeWithdrawn(log types.Log) (*InvisiCoreBroadcasterFeeWithdrawn, error) {
	event := new(InvisiCoreBroadcasterFeeWithdrawn)
	if err := _InvisiCore.contract.UnpackLog(event, "BroadcasterFeeWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCircuitPauseUpdatedIterator is returned from FilterCircuitPauseUpdated and is used to iterate over the raw logs and unpacked data for CircuitPauseUpdated events raised by the InvisiCore contract.
type InvisiCoreCircuitPauseUpdatedIterator struct {
	Event *InvisiCoreCircuitPauseUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCircuitPauseUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCircuitPauseUpdated)
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
		it.Event = new(InvisiCoreCircuitPauseUpdated)
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
func (it *InvisiCoreCircuitPauseUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCircuitPauseUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCircuitPauseUpdated represents a CircuitPauseUpdated event raised by the InvisiCore contract.
type InvisiCoreCircuitPauseUpdated struct {
	Paused bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCircuitPauseUpdated is a free log retrieval operation binding the contract event 0xadc9431437c5558e6a2e308a3be13eebadd57c3d88323a532a827a019aecb2d8.
//
// Solidity: event CircuitPauseUpdated(bool paused)
func (_InvisiCore *InvisiCoreFilterer) FilterCircuitPauseUpdated(opts *bind.FilterOpts) (*InvisiCoreCircuitPauseUpdatedIterator, error) {

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CircuitPauseUpdated")
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCircuitPauseUpdatedIterator{contract: _InvisiCore.contract, event: "CircuitPauseUpdated", logs: logs, sub: sub}, nil
}

// WatchCircuitPauseUpdated is a free log subscription operation binding the contract event 0xadc9431437c5558e6a2e308a3be13eebadd57c3d88323a532a827a019aecb2d8.
//
// Solidity: event CircuitPauseUpdated(bool paused)
func (_InvisiCore *InvisiCoreFilterer) WatchCircuitPauseUpdated(opts *bind.WatchOpts, sink chan<- *InvisiCoreCircuitPauseUpdated) (event.Subscription, error) {

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CircuitPauseUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCircuitPauseUpdated)
				if err := _InvisiCore.contract.UnpackLog(event, "CircuitPauseUpdated", log); err != nil {
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

// ParseCircuitPauseUpdated is a log parse operation binding the contract event 0xadc9431437c5558e6a2e308a3be13eebadd57c3d88323a532a827a019aecb2d8.
//
// Solidity: event CircuitPauseUpdated(bool paused)
func (_InvisiCore *InvisiCoreFilterer) ParseCircuitPauseUpdated(log types.Log) (*InvisiCoreCircuitPauseUpdated, error) {
	event := new(InvisiCoreCircuitPauseUpdated)
	if err := _InvisiCore.contract.UnpackLog(event, "CircuitPauseUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCommitmentsInsertedIterator is returned from FilterCommitmentsInserted and is used to iterate over the raw logs and unpacked data for CommitmentsInserted events raised by the InvisiCore contract.
type InvisiCoreCommitmentsInsertedIterator struct {
	Event *InvisiCoreCommitmentsInserted // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCommitmentsInsertedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCommitmentsInserted)
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
		it.Event = new(InvisiCoreCommitmentsInserted)
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
func (it *InvisiCoreCommitmentsInsertedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCommitmentsInsertedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCommitmentsInserted represents a CommitmentsInserted event raised by the InvisiCore contract.
type InvisiCoreCommitmentsInserted struct {
	TreeId         *big.Int
	StartLeafIndex uint32
	Commitments    [][32]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterCommitmentsInserted is a free log retrieval operation binding the contract event 0x184a7bd65e05e172f493712d9cf2e9cd5fab522a20803086fec8c1502742214a.
//
// Solidity: event CommitmentsInserted(uint256 indexed treeId, uint32 startLeafIndex, bytes32[] commitments)
func (_InvisiCore *InvisiCoreFilterer) FilterCommitmentsInserted(opts *bind.FilterOpts, treeId []*big.Int) (*InvisiCoreCommitmentsInsertedIterator, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CommitmentsInserted", treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCommitmentsInsertedIterator{contract: _InvisiCore.contract, event: "CommitmentsInserted", logs: logs, sub: sub}, nil
}

// WatchCommitmentsInserted is a free log subscription operation binding the contract event 0x184a7bd65e05e172f493712d9cf2e9cd5fab522a20803086fec8c1502742214a.
//
// Solidity: event CommitmentsInserted(uint256 indexed treeId, uint32 startLeafIndex, bytes32[] commitments)
func (_InvisiCore *InvisiCoreFilterer) WatchCommitmentsInserted(opts *bind.WatchOpts, sink chan<- *InvisiCoreCommitmentsInserted, treeId []*big.Int) (event.Subscription, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CommitmentsInserted", treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCommitmentsInserted)
				if err := _InvisiCore.contract.UnpackLog(event, "CommitmentsInserted", log); err != nil {
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

// ParseCommitmentsInserted is a log parse operation binding the contract event 0x184a7bd65e05e172f493712d9cf2e9cd5fab522a20803086fec8c1502742214a.
//
// Solidity: event CommitmentsInserted(uint256 indexed treeId, uint32 startLeafIndex, bytes32[] commitments)
func (_InvisiCore *InvisiCoreFilterer) ParseCommitmentsInserted(log types.Log) (*InvisiCoreCommitmentsInserted, error) {
	event := new(InvisiCoreCommitmentsInserted)
	if err := _InvisiCore.contract.UnpackLog(event, "CommitmentsInserted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCreditCreatedIterator is returned from FilterCreditCreated and is used to iterate over the raw logs and unpacked data for CreditCreated events raised by the InvisiCore contract.
type InvisiCoreCreditCreatedIterator struct {
	Event *InvisiCoreCreditCreated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCreditCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCreditCreated)
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
		it.Event = new(InvisiCoreCreditCreated)
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
func (it *InvisiCoreCreditCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCreditCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCreditCreated represents a CreditCreated event raised by the InvisiCore contract.
type InvisiCoreCreditCreated struct {
	Token            common.Address
	Owner            common.Address
	OutputTreeId     *big.Int
	RevocationShard  uint64
	RevocationSlot   uint16
	CreditCommitment [32]byte
	LeafIndex        uint32
	Deliveries       []IInvisiCoreDelivery
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCreditCreated is a free log retrieval operation binding the contract event 0x3ff78f0284d33ce5183bc7c13bbedfddbf90722c4b3d90a5b8e59da9bbd70565.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint64 revocationShard, uint16 revocationSlot, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditCreated(opts *bind.FilterOpts, token []common.Address, owner []common.Address, outputTreeId []*big.Int) (*InvisiCoreCreditCreatedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditCreated", tokenRule, ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditCreatedIterator{contract: _InvisiCore.contract, event: "CreditCreated", logs: logs, sub: sub}, nil
}

// WatchCreditCreated is a free log subscription operation binding the contract event 0x3ff78f0284d33ce5183bc7c13bbedfddbf90722c4b3d90a5b8e59da9bbd70565.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint64 revocationShard, uint16 revocationSlot, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditCreated(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditCreated, token []common.Address, owner []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditCreated", tokenRule, ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCreditCreated)
				if err := _InvisiCore.contract.UnpackLog(event, "CreditCreated", log); err != nil {
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

// ParseCreditCreated is a log parse operation binding the contract event 0x3ff78f0284d33ce5183bc7c13bbedfddbf90722c4b3d90a5b8e59da9bbd70565.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint64 revocationShard, uint16 revocationSlot, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditCreated(log types.Log) (*InvisiCoreCreditCreated, error) {
	event := new(InvisiCoreCreditCreated)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCreditDrawIterator is returned from FilterCreditDraw and is used to iterate over the raw logs and unpacked data for CreditDraw events raised by the InvisiCore contract.
type InvisiCoreCreditDrawIterator struct {
	Event *InvisiCoreCreditDraw // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCreditDrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCreditDraw)
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
		it.Event = new(InvisiCoreCreditDraw)
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
func (it *InvisiCoreCreditDrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCreditDrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCreditDraw represents a CreditDraw event raised by the InvisiCore contract.
type InvisiCoreCreditDraw struct {
	OutputTreeId *big.Int
	Params       IInvisiCoreCreditDrawParams
	LeafIndices  []uint32
	Deliveries   []IInvisiCoreDelivery
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditDraw is a free log retrieval operation binding the contract event 0xa68be504130022fbd729573715873c6c515c318a014d5273815e20abcc08439d.
//
// Solidity: event CreditDraw(uint256 indexed outputTreeId, (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) params, uint32[] leafIndices, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditDraw(opts *bind.FilterOpts, outputTreeId []*big.Int) (*InvisiCoreCreditDrawIterator, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditDraw", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditDrawIterator{contract: _InvisiCore.contract, event: "CreditDraw", logs: logs, sub: sub}, nil
}

// WatchCreditDraw is a free log subscription operation binding the contract event 0xa68be504130022fbd729573715873c6c515c318a014d5273815e20abcc08439d.
//
// Solidity: event CreditDraw(uint256 indexed outputTreeId, (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) params, uint32[] leafIndices, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditDraw(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditDraw, outputTreeId []*big.Int) (event.Subscription, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditDraw", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCreditDraw)
				if err := _InvisiCore.contract.UnpackLog(event, "CreditDraw", log); err != nil {
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

// ParseCreditDraw is a log parse operation binding the contract event 0xa68be504130022fbd729573715873c6c515c318a014d5273815e20abcc08439d.
//
// Solidity: event CreditDraw(uint256 indexed outputTreeId, (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) params, uint32[] leafIndices, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditDraw(log types.Log) (*InvisiCoreCreditDraw, error) {
	event := new(InvisiCoreCreditDraw)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditDraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCreditRevokedIterator is returned from FilterCreditRevoked and is used to iterate over the raw logs and unpacked data for CreditRevoked events raised by the InvisiCore contract.
type InvisiCoreCreditRevokedIterator struct {
	Event *InvisiCoreCreditRevoked // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCreditRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCreditRevoked)
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
		it.Event = new(InvisiCoreCreditRevoked)
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
func (it *InvisiCoreCreditRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCreditRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCreditRevoked represents a CreditRevoked event raised by the InvisiCore contract.
type InvisiCoreCreditRevoked struct {
	Owner        common.Address
	Token        common.Address
	Shard        uint64
	Slot         uint16
	FeeRecipient common.Address
	FeeAmount    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditRevoked is a free log retrieval operation binding the contract event 0x417df449bda975b802268d8343b7ebb855bdc3f24c8b5c75d8772f56642c9710.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint64 indexed shard, uint16 slot, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditRevoked(opts *bind.FilterOpts, owner []common.Address, token []common.Address, shard []uint64) (*InvisiCoreCreditRevokedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var shardRule []interface{}
	for _, shardItem := range shard {
		shardRule = append(shardRule, shardItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditRevoked", ownerRule, tokenRule, shardRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditRevokedIterator{contract: _InvisiCore.contract, event: "CreditRevoked", logs: logs, sub: sub}, nil
}

// WatchCreditRevoked is a free log subscription operation binding the contract event 0x417df449bda975b802268d8343b7ebb855bdc3f24c8b5c75d8772f56642c9710.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint64 indexed shard, uint16 slot, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditRevoked(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditRevoked, owner []common.Address, token []common.Address, shard []uint64) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var shardRule []interface{}
	for _, shardItem := range shard {
		shardRule = append(shardRule, shardItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditRevoked", ownerRule, tokenRule, shardRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCreditRevoked)
				if err := _InvisiCore.contract.UnpackLog(event, "CreditRevoked", log); err != nil {
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

// ParseCreditRevoked is a log parse operation binding the contract event 0x417df449bda975b802268d8343b7ebb855bdc3f24c8b5c75d8772f56642c9710.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint64 indexed shard, uint16 slot, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditRevoked(log types.Log) (*InvisiCoreCreditRevoked, error) {
	event := new(InvisiCoreCreditRevoked)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCreditUpdatedIterator is returned from FilterCreditUpdated and is used to iterate over the raw logs and unpacked data for CreditUpdated events raised by the InvisiCore contract.
type InvisiCoreCreditUpdatedIterator struct {
	Event *InvisiCoreCreditUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCreditUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCreditUpdated)
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
		it.Event = new(InvisiCoreCreditUpdated)
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
func (it *InvisiCoreCreditUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCreditUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCreditUpdated represents a CreditUpdated event raised by the InvisiCore contract.
type InvisiCoreCreditUpdated struct {
	Owner        common.Address
	OutputTreeId *big.Int
	Params       IInvisiCoreUpdateCreditParams
	LeafIndex    uint32
	Deliveries   []IInvisiCoreDelivery
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditUpdated is a free log retrieval operation binding the contract event 0x4a3fdcaf61cef0446996557873d07793a373211d25c66949a80fe711b29b8d15.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditUpdated(opts *bind.FilterOpts, owner []common.Address, outputTreeId []*big.Int) (*InvisiCoreCreditUpdatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditUpdated", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditUpdatedIterator{contract: _InvisiCore.contract, event: "CreditUpdated", logs: logs, sub: sub}, nil
}

// WatchCreditUpdated is a free log subscription operation binding the contract event 0x4a3fdcaf61cef0446996557873d07793a373211d25c66949a80fe711b29b8d15.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditUpdated(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditUpdated, owner []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditUpdated", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCreditUpdated)
				if err := _InvisiCore.contract.UnpackLog(event, "CreditUpdated", log); err != nil {
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

// ParseCreditUpdated is a log parse operation binding the contract event 0x4a3fdcaf61cef0446996557873d07793a373211d25c66949a80fe711b29b8d15.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditUpdated(log types.Log) (*InvisiCoreCreditUpdated, error) {
	event := new(InvisiCoreCreditUpdated)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCrossTreeTransactIterator is returned from FilterCrossTreeTransact and is used to iterate over the raw logs and unpacked data for CrossTreeTransact events raised by the InvisiCore contract.
type InvisiCoreCrossTreeTransactIterator struct {
	Event *InvisiCoreCrossTreeTransact // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCrossTreeTransactIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCrossTreeTransact)
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
		it.Event = new(InvisiCoreCrossTreeTransact)
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
func (it *InvisiCoreCrossTreeTransactIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCrossTreeTransactIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCrossTreeTransact represents a CrossTreeTransact event raised by the InvisiCore contract.
type InvisiCoreCrossTreeTransact struct {
	Token             common.Address
	OutputTreeId      *big.Int
	Sources           []IInvisiCoreInputSource
	Nullifiers        [][32]byte
	NoteCommitments   [][32]byte
	LeafIndices       []uint32
	Deliveries        []IInvisiCoreDelivery
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCrossTreeTransact is a free log retrieval operation binding the contract event 0x4f0edefe49cf4f3b4346ad9b94338d41abb8bdd8fdcdd7fdb9debeb27b80113d.
//
// Solidity: event CrossTreeTransact(address token, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] noteCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) FilterCrossTreeTransact(opts *bind.FilterOpts, outputTreeId []*big.Int) (*InvisiCoreCrossTreeTransactIterator, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CrossTreeTransact", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCrossTreeTransactIterator{contract: _InvisiCore.contract, event: "CrossTreeTransact", logs: logs, sub: sub}, nil
}

// WatchCrossTreeTransact is a free log subscription operation binding the contract event 0x4f0edefe49cf4f3b4346ad9b94338d41abb8bdd8fdcdd7fdb9debeb27b80113d.
//
// Solidity: event CrossTreeTransact(address token, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] noteCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) WatchCrossTreeTransact(opts *bind.WatchOpts, sink chan<- *InvisiCoreCrossTreeTransact, outputTreeId []*big.Int) (event.Subscription, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CrossTreeTransact", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCrossTreeTransact)
				if err := _InvisiCore.contract.UnpackLog(event, "CrossTreeTransact", log); err != nil {
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

// ParseCrossTreeTransact is a log parse operation binding the contract event 0x4f0edefe49cf4f3b4346ad9b94338d41abb8bdd8fdcdd7fdb9debeb27b80113d.
//
// Solidity: event CrossTreeTransact(address token, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] noteCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) ParseCrossTreeTransact(log types.Log) (*InvisiCoreCrossTreeTransact, error) {
	event := new(InvisiCoreCrossTreeTransact)
	if err := _InvisiCore.contract.UnpackLog(event, "CrossTreeTransact", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the InvisiCore contract.
type InvisiCoreEIP712DomainChangedIterator struct {
	Event *InvisiCoreEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *InvisiCoreEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreEIP712DomainChanged)
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
		it.Event = new(InvisiCoreEIP712DomainChanged)
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
func (it *InvisiCoreEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreEIP712DomainChanged represents a EIP712DomainChanged event raised by the InvisiCore contract.
type InvisiCoreEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiCore *InvisiCoreFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*InvisiCoreEIP712DomainChangedIterator, error) {

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &InvisiCoreEIP712DomainChangedIterator{contract: _InvisiCore.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiCore *InvisiCoreFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *InvisiCoreEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreEIP712DomainChanged)
				if err := _InvisiCore.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiCore *InvisiCoreFilterer) ParseEIP712DomainChanged(log types.Log) (*InvisiCoreEIP712DomainChanged, error) {
	event := new(InvisiCoreEIP712DomainChanged)
	if err := _InvisiCore.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreNonceInvalidatedIterator is returned from FilterNonceInvalidated and is used to iterate over the raw logs and unpacked data for NonceInvalidated events raised by the InvisiCore contract.
type InvisiCoreNonceInvalidatedIterator struct {
	Event *InvisiCoreNonceInvalidated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreNonceInvalidatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreNonceInvalidated)
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
		it.Event = new(InvisiCoreNonceInvalidated)
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
func (it *InvisiCoreNonceInvalidatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreNonceInvalidatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreNonceInvalidated represents a NonceInvalidated event raised by the InvisiCore contract.
type InvisiCoreNonceInvalidated struct {
	Signer   common.Address
	NewNonce *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNonceInvalidated is a free log retrieval operation binding the contract event 0x1800cd2301fbc20790ed94f3d55a28ef2306a9c31cd3c72b5b71b6e4cf5c6241.
//
// Solidity: event NonceInvalidated(address indexed signer, uint256 newNonce)
func (_InvisiCore *InvisiCoreFilterer) FilterNonceInvalidated(opts *bind.FilterOpts, signer []common.Address) (*InvisiCoreNonceInvalidatedIterator, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "NonceInvalidated", signerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreNonceInvalidatedIterator{contract: _InvisiCore.contract, event: "NonceInvalidated", logs: logs, sub: sub}, nil
}

// WatchNonceInvalidated is a free log subscription operation binding the contract event 0x1800cd2301fbc20790ed94f3d55a28ef2306a9c31cd3c72b5b71b6e4cf5c6241.
//
// Solidity: event NonceInvalidated(address indexed signer, uint256 newNonce)
func (_InvisiCore *InvisiCoreFilterer) WatchNonceInvalidated(opts *bind.WatchOpts, sink chan<- *InvisiCoreNonceInvalidated, signer []common.Address) (event.Subscription, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "NonceInvalidated", signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreNonceInvalidated)
				if err := _InvisiCore.contract.UnpackLog(event, "NonceInvalidated", log); err != nil {
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

// ParseNonceInvalidated is a log parse operation binding the contract event 0x1800cd2301fbc20790ed94f3d55a28ef2306a9c31cd3c72b5b71b6e4cf5c6241.
//
// Solidity: event NonceInvalidated(address indexed signer, uint256 newNonce)
func (_InvisiCore *InvisiCoreFilterer) ParseNonceInvalidated(log types.Log) (*InvisiCoreNonceInvalidated, error) {
	event := new(InvisiCoreNonceInvalidated)
	if err := _InvisiCore.contract.UnpackLog(event, "NonceInvalidated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the InvisiCore contract.
type InvisiCoreOwnershipTransferStartedIterator struct {
	Event *InvisiCoreOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *InvisiCoreOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreOwnershipTransferStarted)
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
		it.Event = new(InvisiCoreOwnershipTransferStarted)
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
func (it *InvisiCoreOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the InvisiCore contract.
type InvisiCoreOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiCore *InvisiCoreFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiCoreOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreOwnershipTransferStartedIterator{contract: _InvisiCore.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiCore *InvisiCoreFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *InvisiCoreOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreOwnershipTransferStarted)
				if err := _InvisiCore.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_InvisiCore *InvisiCoreFilterer) ParseOwnershipTransferStarted(log types.Log) (*InvisiCoreOwnershipTransferStarted, error) {
	event := new(InvisiCoreOwnershipTransferStarted)
	if err := _InvisiCore.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the InvisiCore contract.
type InvisiCoreOwnershipTransferredIterator struct {
	Event *InvisiCoreOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *InvisiCoreOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreOwnershipTransferred)
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
		it.Event = new(InvisiCoreOwnershipTransferred)
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
func (it *InvisiCoreOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreOwnershipTransferred represents a OwnershipTransferred event raised by the InvisiCore contract.
type InvisiCoreOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiCore *InvisiCoreFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiCoreOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreOwnershipTransferredIterator{contract: _InvisiCore.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiCore *InvisiCoreFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *InvisiCoreOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreOwnershipTransferred)
				if err := _InvisiCore.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_InvisiCore *InvisiCoreFilterer) ParseOwnershipTransferred(log types.Log) (*InvisiCoreOwnershipTransferred, error) {
	event := new(InvisiCoreOwnershipTransferred)
	if err := _InvisiCore.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCorePausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the InvisiCore contract.
type InvisiCorePausedIterator struct {
	Event *InvisiCorePaused // Event containing the contract specifics and raw log

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
func (it *InvisiCorePausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCorePaused)
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
		it.Event = new(InvisiCorePaused)
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
func (it *InvisiCorePausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCorePausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCorePaused represents a Paused event raised by the InvisiCore contract.
type InvisiCorePaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_InvisiCore *InvisiCoreFilterer) FilterPaused(opts *bind.FilterOpts) (*InvisiCorePausedIterator, error) {

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &InvisiCorePausedIterator{contract: _InvisiCore.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_InvisiCore *InvisiCoreFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *InvisiCorePaused) (event.Subscription, error) {

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCorePaused)
				if err := _InvisiCore.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_InvisiCore *InvisiCoreFilterer) ParsePaused(log types.Log) (*InvisiCorePaused, error) {
	event := new(InvisiCorePaused)
	if err := _InvisiCore.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCorePublicDepositIterator is returned from FilterPublicDeposit and is used to iterate over the raw logs and unpacked data for PublicDeposit events raised by the InvisiCore contract.
type InvisiCorePublicDepositIterator struct {
	Event *InvisiCorePublicDeposit // Event containing the contract specifics and raw log

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
func (it *InvisiCorePublicDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCorePublicDeposit)
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
		it.Event = new(InvisiCorePublicDeposit)
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
func (it *InvisiCorePublicDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCorePublicDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCorePublicDeposit represents a PublicDeposit event raised by the InvisiCore contract.
type InvisiCorePublicDeposit struct {
	Token     common.Address
	Payer     common.Address
	Amount    *big.Int
	ProofHash [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPublicDeposit is a free log retrieval operation binding the contract event 0x1626a1d6f055052717d0a40e4bb7e02f2b2b379432abf8fcd2ef80d2374e6661.
//
// Solidity: event PublicDeposit(address indexed token, address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiCore *InvisiCoreFilterer) FilterPublicDeposit(opts *bind.FilterOpts, token []common.Address, payer []common.Address) (*InvisiCorePublicDepositIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "PublicDeposit", tokenRule, payerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCorePublicDepositIterator{contract: _InvisiCore.contract, event: "PublicDeposit", logs: logs, sub: sub}, nil
}

// WatchPublicDeposit is a free log subscription operation binding the contract event 0x1626a1d6f055052717d0a40e4bb7e02f2b2b379432abf8fcd2ef80d2374e6661.
//
// Solidity: event PublicDeposit(address indexed token, address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiCore *InvisiCoreFilterer) WatchPublicDeposit(opts *bind.WatchOpts, sink chan<- *InvisiCorePublicDeposit, token []common.Address, payer []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "PublicDeposit", tokenRule, payerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCorePublicDeposit)
				if err := _InvisiCore.contract.UnpackLog(event, "PublicDeposit", log); err != nil {
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

// ParsePublicDeposit is a log parse operation binding the contract event 0x1626a1d6f055052717d0a40e4bb7e02f2b2b379432abf8fcd2ef80d2374e6661.
//
// Solidity: event PublicDeposit(address indexed token, address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiCore *InvisiCoreFilterer) ParsePublicDeposit(log types.Log) (*InvisiCorePublicDeposit, error) {
	event := new(InvisiCorePublicDeposit)
	if err := _InvisiCore.contract.UnpackLog(event, "PublicDeposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreScreenerUpdatedIterator is returned from FilterScreenerUpdated and is used to iterate over the raw logs and unpacked data for ScreenerUpdated events raised by the InvisiCore contract.
type InvisiCoreScreenerUpdatedIterator struct {
	Event *InvisiCoreScreenerUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreScreenerUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreScreenerUpdated)
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
		it.Event = new(InvisiCoreScreenerUpdated)
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
func (it *InvisiCoreScreenerUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreScreenerUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreScreenerUpdated represents a ScreenerUpdated event raised by the InvisiCore contract.
type InvisiCoreScreenerUpdated struct {
	NewScreener common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterScreenerUpdated is a free log retrieval operation binding the contract event 0xb52b505537cdccb66a697cc92b4e51933a403c89ac139bf2297b918de3791dbe.
//
// Solidity: event ScreenerUpdated(address indexed newScreener)
func (_InvisiCore *InvisiCoreFilterer) FilterScreenerUpdated(opts *bind.FilterOpts, newScreener []common.Address) (*InvisiCoreScreenerUpdatedIterator, error) {

	var newScreenerRule []interface{}
	for _, newScreenerItem := range newScreener {
		newScreenerRule = append(newScreenerRule, newScreenerItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "ScreenerUpdated", newScreenerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreScreenerUpdatedIterator{contract: _InvisiCore.contract, event: "ScreenerUpdated", logs: logs, sub: sub}, nil
}

// WatchScreenerUpdated is a free log subscription operation binding the contract event 0xb52b505537cdccb66a697cc92b4e51933a403c89ac139bf2297b918de3791dbe.
//
// Solidity: event ScreenerUpdated(address indexed newScreener)
func (_InvisiCore *InvisiCoreFilterer) WatchScreenerUpdated(opts *bind.WatchOpts, sink chan<- *InvisiCoreScreenerUpdated, newScreener []common.Address) (event.Subscription, error) {

	var newScreenerRule []interface{}
	for _, newScreenerItem := range newScreener {
		newScreenerRule = append(newScreenerRule, newScreenerItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "ScreenerUpdated", newScreenerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreScreenerUpdated)
				if err := _InvisiCore.contract.UnpackLog(event, "ScreenerUpdated", log); err != nil {
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

// ParseScreenerUpdated is a log parse operation binding the contract event 0xb52b505537cdccb66a697cc92b4e51933a403c89ac139bf2297b918de3791dbe.
//
// Solidity: event ScreenerUpdated(address indexed newScreener)
func (_InvisiCore *InvisiCoreFilterer) ParseScreenerUpdated(log types.Log) (*InvisiCoreScreenerUpdated, error) {
	event := new(InvisiCoreScreenerUpdated)
	if err := _InvisiCore.contract.UnpackLog(event, "ScreenerUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreShieldIterator is returned from FilterShield and is used to iterate over the raw logs and unpacked data for Shield events raised by the InvisiCore contract.
type InvisiCoreShieldIterator struct {
	Event *InvisiCoreShield // Event containing the contract specifics and raw log

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
func (it *InvisiCoreShieldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreShield)
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
		it.Event = new(InvisiCoreShield)
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
func (it *InvisiCoreShieldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreShieldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreShield represents a Shield event raised by the InvisiCore contract.
type InvisiCoreShield struct {
	Token          common.Address
	Payer          common.Address
	Amount         *big.Int
	TreeId         *big.Int
	StartLeafIndex uint32
	Deliveries     []IInvisiCoreDelivery
	Deadline       *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterShield is a free log retrieval operation binding the contract event 0xefe7bab38d42993e5331d7dd43bbf9a0e1fa493cdd18f5d6d0b980f5b324f0d1.
//
// Solidity: event Shield(address indexed token, address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, (bytes,bytes)[] deliveries, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) FilterShield(opts *bind.FilterOpts, token []common.Address, payer []common.Address, treeId []*big.Int) (*InvisiCoreShieldIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "Shield", tokenRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreShieldIterator{contract: _InvisiCore.contract, event: "Shield", logs: logs, sub: sub}, nil
}

// WatchShield is a free log subscription operation binding the contract event 0xefe7bab38d42993e5331d7dd43bbf9a0e1fa493cdd18f5d6d0b980f5b324f0d1.
//
// Solidity: event Shield(address indexed token, address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, (bytes,bytes)[] deliveries, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) WatchShield(opts *bind.WatchOpts, sink chan<- *InvisiCoreShield, token []common.Address, payer []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "Shield", tokenRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreShield)
				if err := _InvisiCore.contract.UnpackLog(event, "Shield", log); err != nil {
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

// ParseShield is a log parse operation binding the contract event 0xefe7bab38d42993e5331d7dd43bbf9a0e1fa493cdd18f5d6d0b980f5b324f0d1.
//
// Solidity: event Shield(address indexed token, address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, (bytes,bytes)[] deliveries, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) ParseShield(log types.Log) (*InvisiCoreShield, error) {
	event := new(InvisiCoreShield)
	if err := _InvisiCore.contract.UnpackLog(event, "Shield", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreTransactIterator is returned from FilterTransact and is used to iterate over the raw logs and unpacked data for Transact events raised by the InvisiCore contract.
type InvisiCoreTransactIterator struct {
	Event *InvisiCoreTransact // Event containing the contract specifics and raw log

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
func (it *InvisiCoreTransactIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreTransact)
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
		it.Event = new(InvisiCoreTransact)
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
func (it *InvisiCoreTransactIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreTransactIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreTransact represents a Transact event raised by the InvisiCore contract.
type InvisiCoreTransact struct {
	Token             common.Address
	InputTreeId       *big.Int
	OutputTreeId      *big.Int
	Nullifiers        [][32]byte
	NoteCommitments   [][32]byte
	LeafIndices       []uint32
	Deliveries        []IInvisiCoreDelivery
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterTransact is a free log retrieval operation binding the contract event 0xac5599eb2d39ff304a3cfae313226db689176cd11a081b088a08b2ade0c88382.
//
// Solidity: event Transact(address token, uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] noteCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) FilterTransact(opts *bind.FilterOpts, inputTreeId []*big.Int, outputTreeId []*big.Int) (*InvisiCoreTransactIterator, error) {

	var inputTreeIdRule []interface{}
	for _, inputTreeIdItem := range inputTreeId {
		inputTreeIdRule = append(inputTreeIdRule, inputTreeIdItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "Transact", inputTreeIdRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreTransactIterator{contract: _InvisiCore.contract, event: "Transact", logs: logs, sub: sub}, nil
}

// WatchTransact is a free log subscription operation binding the contract event 0xac5599eb2d39ff304a3cfae313226db689176cd11a081b088a08b2ade0c88382.
//
// Solidity: event Transact(address token, uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] noteCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) WatchTransact(opts *bind.WatchOpts, sink chan<- *InvisiCoreTransact, inputTreeId []*big.Int, outputTreeId []*big.Int) (event.Subscription, error) {

	var inputTreeIdRule []interface{}
	for _, inputTreeIdItem := range inputTreeId {
		inputTreeIdRule = append(inputTreeIdRule, inputTreeIdItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "Transact", inputTreeIdRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreTransact)
				if err := _InvisiCore.contract.UnpackLog(event, "Transact", log); err != nil {
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

// ParseTransact is a log parse operation binding the contract event 0xac5599eb2d39ff304a3cfae313226db689176cd11a081b088a08b2ade0c88382.
//
// Solidity: event Transact(address token, uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] noteCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiCore *InvisiCoreFilterer) ParseTransact(log types.Log) (*InvisiCoreTransact, error) {
	event := new(InvisiCoreTransact)
	if err := _InvisiCore.contract.UnpackLog(event, "Transact", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreTransactAggregatedIterator is returned from FilterTransactAggregated and is used to iterate over the raw logs and unpacked data for TransactAggregated events raised by the InvisiCore contract.
type InvisiCoreTransactAggregatedIterator struct {
	Event *InvisiCoreTransactAggregated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreTransactAggregatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreTransactAggregated)
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
		it.Event = new(InvisiCoreTransactAggregated)
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
func (it *InvisiCoreTransactAggregatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreTransactAggregatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreTransactAggregated represents a TransactAggregated event raised by the InvisiCore contract.
type InvisiCoreTransactAggregated struct {
	ProofHash    [32]byte
	RequestCount *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTransactAggregated is a free log retrieval operation binding the contract event 0xb550ba00978bffde806b7356916d552dea8794ebdc978bc7d205be9d2ce1a8df.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, uint256 requestCount)
func (_InvisiCore *InvisiCoreFilterer) FilterTransactAggregated(opts *bind.FilterOpts, proofHash [][32]byte) (*InvisiCoreTransactAggregatedIterator, error) {

	var proofHashRule []interface{}
	for _, proofHashItem := range proofHash {
		proofHashRule = append(proofHashRule, proofHashItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "TransactAggregated", proofHashRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreTransactAggregatedIterator{contract: _InvisiCore.contract, event: "TransactAggregated", logs: logs, sub: sub}, nil
}

// WatchTransactAggregated is a free log subscription operation binding the contract event 0xb550ba00978bffde806b7356916d552dea8794ebdc978bc7d205be9d2ce1a8df.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, uint256 requestCount)
func (_InvisiCore *InvisiCoreFilterer) WatchTransactAggregated(opts *bind.WatchOpts, sink chan<- *InvisiCoreTransactAggregated, proofHash [][32]byte) (event.Subscription, error) {

	var proofHashRule []interface{}
	for _, proofHashItem := range proofHash {
		proofHashRule = append(proofHashRule, proofHashItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "TransactAggregated", proofHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreTransactAggregated)
				if err := _InvisiCore.contract.UnpackLog(event, "TransactAggregated", log); err != nil {
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

// ParseTransactAggregated is a log parse operation binding the contract event 0xb550ba00978bffde806b7356916d552dea8794ebdc978bc7d205be9d2ce1a8df.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, uint256 requestCount)
func (_InvisiCore *InvisiCoreFilterer) ParseTransactAggregated(log types.Log) (*InvisiCoreTransactAggregated, error) {
	event := new(InvisiCoreTransactAggregated)
	if err := _InvisiCore.contract.UnpackLog(event, "TransactAggregated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreTreeCreatedIterator is returned from FilterTreeCreated and is used to iterate over the raw logs and unpacked data for TreeCreated events raised by the InvisiCore contract.
type InvisiCoreTreeCreatedIterator struct {
	Event *InvisiCoreTreeCreated // Event containing the contract specifics and raw log

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
func (it *InvisiCoreTreeCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreTreeCreated)
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
		it.Event = new(InvisiCoreTreeCreated)
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
func (it *InvisiCoreTreeCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreTreeCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreTreeCreated represents a TreeCreated event raised by the InvisiCore contract.
type InvisiCoreTreeCreated struct {
	TreeId *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTreeCreated is a free log retrieval operation binding the contract event 0x8403033038c2f099c559e73d79138265904c614f5925083ccb2cef7bdeaba28a.
//
// Solidity: event TreeCreated(uint256 indexed treeId)
func (_InvisiCore *InvisiCoreFilterer) FilterTreeCreated(opts *bind.FilterOpts, treeId []*big.Int) (*InvisiCoreTreeCreatedIterator, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "TreeCreated", treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreTreeCreatedIterator{contract: _InvisiCore.contract, event: "TreeCreated", logs: logs, sub: sub}, nil
}

// WatchTreeCreated is a free log subscription operation binding the contract event 0x8403033038c2f099c559e73d79138265904c614f5925083ccb2cef7bdeaba28a.
//
// Solidity: event TreeCreated(uint256 indexed treeId)
func (_InvisiCore *InvisiCoreFilterer) WatchTreeCreated(opts *bind.WatchOpts, sink chan<- *InvisiCoreTreeCreated, treeId []*big.Int) (event.Subscription, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "TreeCreated", treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreTreeCreated)
				if err := _InvisiCore.contract.UnpackLog(event, "TreeCreated", log); err != nil {
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

// ParseTreeCreated is a log parse operation binding the contract event 0x8403033038c2f099c559e73d79138265904c614f5925083ccb2cef7bdeaba28a.
//
// Solidity: event TreeCreated(uint256 indexed treeId)
func (_InvisiCore *InvisiCoreFilterer) ParseTreeCreated(log types.Log) (*InvisiCoreTreeCreated, error) {
	event := new(InvisiCoreTreeCreated)
	if err := _InvisiCore.contract.UnpackLog(event, "TreeCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the InvisiCore contract.
type InvisiCoreUnpausedIterator struct {
	Event *InvisiCoreUnpaused // Event containing the contract specifics and raw log

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
func (it *InvisiCoreUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreUnpaused)
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
		it.Event = new(InvisiCoreUnpaused)
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
func (it *InvisiCoreUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreUnpaused represents a Unpaused event raised by the InvisiCore contract.
type InvisiCoreUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_InvisiCore *InvisiCoreFilterer) FilterUnpaused(opts *bind.FilterOpts) (*InvisiCoreUnpausedIterator, error) {

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &InvisiCoreUnpausedIterator{contract: _InvisiCore.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_InvisiCore *InvisiCoreFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *InvisiCoreUnpaused) (event.Subscription, error) {

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreUnpaused)
				if err := _InvisiCore.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_InvisiCore *InvisiCoreFilterer) ParseUnpaused(log types.Log) (*InvisiCoreUnpaused, error) {
	event := new(InvisiCoreUnpaused)
	if err := _InvisiCore.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
