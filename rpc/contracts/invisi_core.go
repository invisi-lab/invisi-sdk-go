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

// IInvisiCoreCreditDrawState is an auto generated low-level Go binding around an user-defined struct.
type IInvisiCoreCreditDrawState = protocolcontracts.IInvisiCoreCreditDrawState

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

// CreditStatus is a free data retrieval call binding the contract method 0x1dadbb55.
//
// Solidity: function creditStatus(address owner, address token, uint128 grantId) view returns(uint256)
func (_InvisiCore *InvisiCoreCaller) CreditStatus(opts *bind.CallOpts, owner common.Address, token common.Address, grantId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "creditStatus", owner, token, grantId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CreditStatus is a free data retrieval call binding the contract method 0x1dadbb55.
//
// Solidity: function creditStatus(address owner, address token, uint128 grantId) view returns(uint256)
func (_InvisiCore *InvisiCoreSession) CreditStatus(owner common.Address, token common.Address, grantId *big.Int) (*big.Int, error) {
	return _InvisiCore.Contract.CreditStatus(&_InvisiCore.CallOpts, owner, token, grantId)
}

// CreditStatus is a free data retrieval call binding the contract method 0x1dadbb55.
//
// Solidity: function creditStatus(address owner, address token, uint128 grantId) view returns(uint256)
func (_InvisiCore *InvisiCoreCallerSession) CreditStatus(owner common.Address, token common.Address, grantId *big.Int) (*big.Int, error) {
	return _InvisiCore.Contract.CreditStatus(&_InvisiCore.CallOpts, owner, token, grantId)
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

// InstanceCircuitPaused is a free data retrieval call binding the contract method 0xbcdf9e1a.
//
// Solidity: function instanceCircuitPaused() view returns(bool)
func (_InvisiCore *InvisiCoreCaller) InstanceCircuitPaused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _InvisiCore.contract.Call(opts, &out, "instanceCircuitPaused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// InstanceCircuitPaused is a free data retrieval call binding the contract method 0xbcdf9e1a.
//
// Solidity: function instanceCircuitPaused() view returns(bool)
func (_InvisiCore *InvisiCoreSession) InstanceCircuitPaused() (bool, error) {
	return _InvisiCore.Contract.InstanceCircuitPaused(&_InvisiCore.CallOpts)
}

// InstanceCircuitPaused is a free data retrieval call binding the contract method 0xbcdf9e1a.
//
// Solidity: function instanceCircuitPaused() view returns(bool)
func (_InvisiCore *InvisiCoreCallerSession) InstanceCircuitPaused() (bool, error) {
	return _InvisiCore.Contract.InstanceCircuitPaused(&_InvisiCore.CallOpts)
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

// CreateCredit is a paid mutator transaction binding the contract method 0x4aa2aed8.
//
// Solidity: function createCredit(bytes , (address,address,uint128,bytes32,uint256,address,uint256,bool,uint256,bytes32) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactor) CreateCredit(opts *bind.TransactOpts, arg0 []byte, p IInvisiCoreCreateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "createCredit", arg0, p, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x4aa2aed8.
//
// Solidity: function createCredit(bytes , (address,address,uint128,bytes32,uint256,address,uint256,bool,uint256,bytes32) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreSession) CreateCredit(arg0 []byte, p IInvisiCoreCreateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x4aa2aed8.
//
// Solidity: function createCredit(bytes , (address,address,uint128,bytes32,uint256,address,uint256,bool,uint256,bytes32) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactorSession) CreateCredit(arg0 []byte, p IInvisiCoreCreateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa28f6594.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint128)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactor) CreditDraw(opts *bind.TransactOpts, arg0 []byte, p IInvisiCoreCreditDrawParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "creditDraw", arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa28f6594.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint128)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreSession) CreditDraw(arg0 []byte, p IInvisiCoreCreditDrawParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.CreditDraw(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa28f6594.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint128)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
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

// RevokeCredit is a paid mutator transaction binding the contract method 0x78145727.
//
// Solidity: function revokeCredit(address token, address owner, uint128 grantId, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiCore *InvisiCoreTransactor) RevokeCredit(opts *bind.TransactOpts, token common.Address, owner common.Address, grantId *big.Int, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "revokeCredit", token, owner, grantId, deadline, feeRecipient, feeAmount, signature)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x78145727.
//
// Solidity: function revokeCredit(address token, address owner, uint128 grantId, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiCore *InvisiCoreSession) RevokeCredit(token common.Address, owner common.Address, grantId *big.Int, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiCore.Contract.RevokeCredit(&_InvisiCore.TransactOpts, token, owner, grantId, deadline, feeRecipient, feeAmount, signature)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x78145727.
//
// Solidity: function revokeCredit(address token, address owner, uint128 grantId, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiCore *InvisiCoreTransactorSession) RevokeCredit(token common.Address, owner common.Address, grantId *big.Int, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiCore.Contract.RevokeCredit(&_InvisiCore.TransactOpts, token, owner, grantId, deadline, feeRecipient, feeAmount, signature)
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

// UpdateCredit is a paid mutator transaction binding the contract method 0xc8ffbcbb.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint128,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreTransactor) UpdateCredit(opts *bind.TransactOpts, arg0 []byte, p IInvisiCoreUpdateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.contract.Transact(opts, "updateCredit", arg0, p, arg2)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0xc8ffbcbb.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint128,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiCore *InvisiCoreSession) UpdateCredit(arg0 []byte, p IInvisiCoreUpdateCreditParams, arg2 []IInvisiCoreDelivery) (*types.Transaction, error) {
	return _InvisiCore.Contract.UpdateCredit(&_InvisiCore.TransactOpts, arg0, p, arg2)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0xc8ffbcbb.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint128,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
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
	GrantId          *big.Int
	CreditCommitment [32]byte
	LeafIndex        uint32
	Deliveries       []IInvisiCoreDelivery
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCreditCreated is a free log retrieval operation binding the contract event 0x36bde929d19ac9442803c1e805b2610f98e3eae09a87c4465ce65b7efeebe5fb.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint128 grantId, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
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

// WatchCreditCreated is a free log subscription operation binding the contract event 0x36bde929d19ac9442803c1e805b2610f98e3eae09a87c4465ce65b7efeebe5fb.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint128 grantId, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
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

// ParseCreditCreated is a log parse operation binding the contract event 0x36bde929d19ac9442803c1e805b2610f98e3eae09a87c4465ce65b7efeebe5fb.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint128 grantId, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
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
	Token          common.Address
	OutputTreeId   *big.Int
	NoteCommitment [32]byte
	NoteLeafIndex  uint32
	TotalAmount    *big.Int
	Deadline       *big.Int
	FeeRecipient   common.Address
	FeeAmount      *big.Int
	Delivery       IInvisiCoreDelivery
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterCreditDraw is a free log retrieval operation binding the contract event 0x2ec095e4edccddd919763ec1a653b334c331c75743d7ccda699c23522bff175e.
//
// Solidity: event CreditDraw(address indexed token, uint256 indexed outputTreeId, bytes32 indexed noteCommitment, uint32 noteLeafIndex, uint256 totalAmount, uint256 deadline, address feeRecipient, uint256 feeAmount, (bytes,bytes) delivery)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditDraw(opts *bind.FilterOpts, token []common.Address, outputTreeId []*big.Int, noteCommitment [][32]byte) (*InvisiCoreCreditDrawIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}
	var noteCommitmentRule []interface{}
	for _, noteCommitmentItem := range noteCommitment {
		noteCommitmentRule = append(noteCommitmentRule, noteCommitmentItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditDraw", tokenRule, outputTreeIdRule, noteCommitmentRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditDrawIterator{contract: _InvisiCore.contract, event: "CreditDraw", logs: logs, sub: sub}, nil
}

// WatchCreditDraw is a free log subscription operation binding the contract event 0x2ec095e4edccddd919763ec1a653b334c331c75743d7ccda699c23522bff175e.
//
// Solidity: event CreditDraw(address indexed token, uint256 indexed outputTreeId, bytes32 indexed noteCommitment, uint32 noteLeafIndex, uint256 totalAmount, uint256 deadline, address feeRecipient, uint256 feeAmount, (bytes,bytes) delivery)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditDraw(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditDraw, token []common.Address, outputTreeId []*big.Int, noteCommitment [][32]byte) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}
	var noteCommitmentRule []interface{}
	for _, noteCommitmentItem := range noteCommitment {
		noteCommitmentRule = append(noteCommitmentRule, noteCommitmentItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditDraw", tokenRule, outputTreeIdRule, noteCommitmentRule)
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

// ParseCreditDraw is a log parse operation binding the contract event 0x2ec095e4edccddd919763ec1a653b334c331c75743d7ccda699c23522bff175e.
//
// Solidity: event CreditDraw(address indexed token, uint256 indexed outputTreeId, bytes32 indexed noteCommitment, uint32 noteLeafIndex, uint256 totalAmount, uint256 deadline, address feeRecipient, uint256 feeAmount, (bytes,bytes) delivery)
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
	GrantId      *big.Int
	FeeRecipient common.Address
	FeeAmount    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditRevoked is a free log retrieval operation binding the contract event 0x90a3beab92b6605275710ce10d97108866aac6fe02d38ff1e61d0f8d40bc6ba2.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint128 indexed grantId, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditRevoked(opts *bind.FilterOpts, owner []common.Address, token []common.Address, grantId []*big.Int) (*InvisiCoreCreditRevokedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var grantIdRule []interface{}
	for _, grantIdItem := range grantId {
		grantIdRule = append(grantIdRule, grantIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditRevoked", ownerRule, tokenRule, grantIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditRevokedIterator{contract: _InvisiCore.contract, event: "CreditRevoked", logs: logs, sub: sub}, nil
}

// WatchCreditRevoked is a free log subscription operation binding the contract event 0x90a3beab92b6605275710ce10d97108866aac6fe02d38ff1e61d0f8d40bc6ba2.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint128 indexed grantId, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditRevoked(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditRevoked, owner []common.Address, token []common.Address, grantId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var grantIdRule []interface{}
	for _, grantIdItem := range grantId {
		grantIdRule = append(grantIdRule, grantIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditRevoked", ownerRule, tokenRule, grantIdRule)
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

// ParseCreditRevoked is a log parse operation binding the contract event 0x90a3beab92b6605275710ce10d97108866aac6fe02d38ff1e61d0f8d40bc6ba2.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint128 indexed grantId, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditRevoked(log types.Log) (*InvisiCoreCreditRevoked, error) {
	event := new(InvisiCoreCreditRevoked)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreCreditStateTransitionIterator is returned from FilterCreditStateTransition and is used to iterate over the raw logs and unpacked data for CreditStateTransition events raised by the InvisiCore contract.
type InvisiCoreCreditStateTransitionIterator struct {
	Event *InvisiCoreCreditStateTransition // Event containing the contract specifics and raw log

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
func (it *InvisiCoreCreditStateTransitionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreCreditStateTransition)
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
		it.Event = new(InvisiCoreCreditStateTransition)
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
func (it *InvisiCoreCreditStateTransitionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreCreditStateTransitionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreCreditStateTransition represents a CreditStateTransition event raised by the InvisiCore contract.
type InvisiCoreCreditStateTransition struct {
	OutputTreeId      *big.Int
	NoteCommitment    [32]byte
	State             IInvisiCoreCreditDrawState
	CreditLeafIndices []uint32
	CreditDeliveries  []IInvisiCoreDelivery
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCreditStateTransition is a free log retrieval operation binding the contract event 0xddfbd53f525d78d9767174b6a64bf2b6e3d5b396e440ccaf51401688b0fdf790.
//
// Solidity: event CreditStateTransition(uint256 indexed outputTreeId, bytes32 indexed noteCommitment, ((uint256,bytes32)[],(address,uint256)[],(uint256,uint128)[],bytes32[],bytes32[],bytes32[]) state, uint32[] creditLeafIndices, (bytes,bytes)[] creditDeliveries)
func (_InvisiCore *InvisiCoreFilterer) FilterCreditStateTransition(opts *bind.FilterOpts, outputTreeId []*big.Int, noteCommitment [][32]byte) (*InvisiCoreCreditStateTransitionIterator, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}
	var noteCommitmentRule []interface{}
	for _, noteCommitmentItem := range noteCommitment {
		noteCommitmentRule = append(noteCommitmentRule, noteCommitmentItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "CreditStateTransition", outputTreeIdRule, noteCommitmentRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreCreditStateTransitionIterator{contract: _InvisiCore.contract, event: "CreditStateTransition", logs: logs, sub: sub}, nil
}

// WatchCreditStateTransition is a free log subscription operation binding the contract event 0xddfbd53f525d78d9767174b6a64bf2b6e3d5b396e440ccaf51401688b0fdf790.
//
// Solidity: event CreditStateTransition(uint256 indexed outputTreeId, bytes32 indexed noteCommitment, ((uint256,bytes32)[],(address,uint256)[],(uint256,uint128)[],bytes32[],bytes32[],bytes32[]) state, uint32[] creditLeafIndices, (bytes,bytes)[] creditDeliveries)
func (_InvisiCore *InvisiCoreFilterer) WatchCreditStateTransition(opts *bind.WatchOpts, sink chan<- *InvisiCoreCreditStateTransition, outputTreeId []*big.Int, noteCommitment [][32]byte) (event.Subscription, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}
	var noteCommitmentRule []interface{}
	for _, noteCommitmentItem := range noteCommitment {
		noteCommitmentRule = append(noteCommitmentRule, noteCommitmentItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "CreditStateTransition", outputTreeIdRule, noteCommitmentRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreCreditStateTransition)
				if err := _InvisiCore.contract.UnpackLog(event, "CreditStateTransition", log); err != nil {
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

// ParseCreditStateTransition is a log parse operation binding the contract event 0xddfbd53f525d78d9767174b6a64bf2b6e3d5b396e440ccaf51401688b0fdf790.
//
// Solidity: event CreditStateTransition(uint256 indexed outputTreeId, bytes32 indexed noteCommitment, ((uint256,bytes32)[],(address,uint256)[],(uint256,uint128)[],bytes32[],bytes32[],bytes32[]) state, uint32[] creditLeafIndices, (bytes,bytes)[] creditDeliveries)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditStateTransition(log types.Log) (*InvisiCoreCreditStateTransition, error) {
	event := new(InvisiCoreCreditStateTransition)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditStateTransition", log); err != nil {
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

// FilterCreditUpdated is a free log retrieval operation binding the contract event 0xd3ad00e6e6ac1302b14dacc0a83efababfeed1472b315de9ff3a2ef8a0d146eb.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint128,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
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

// WatchCreditUpdated is a free log subscription operation binding the contract event 0xd3ad00e6e6ac1302b14dacc0a83efababfeed1472b315de9ff3a2ef8a0d146eb.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint128,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
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

// ParseCreditUpdated is a log parse operation binding the contract event 0xd3ad00e6e6ac1302b14dacc0a83efababfeed1472b315de9ff3a2ef8a0d146eb.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint128,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
func (_InvisiCore *InvisiCoreFilterer) ParseCreditUpdated(log types.Log) (*InvisiCoreCreditUpdated, error) {
	event := new(InvisiCoreCreditUpdated)
	if err := _InvisiCore.contract.UnpackLog(event, "CreditUpdated", log); err != nil {
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

// InvisiCoreNoteOperationIterator is returned from FilterNoteOperation and is used to iterate over the raw logs and unpacked data for NoteOperation events raised by the InvisiCore contract.
type InvisiCoreNoteOperationIterator struct {
	Event *InvisiCoreNoteOperation // Event containing the contract specifics and raw log

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
func (it *InvisiCoreNoteOperationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreNoteOperation)
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
		it.Event = new(InvisiCoreNoteOperation)
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
func (it *InvisiCoreNoteOperationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreNoteOperationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreNoteOperation represents a NoteOperation event raised by the InvisiCore contract.
type InvisiCoreNoteOperation struct {
	OperationId  [32]byte
	Token        common.Address
	Kind         uint8
	OutputCount  *big.Int
	Deadline     *big.Int
	FeeRecipient common.Address
	FeeAmount    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterNoteOperation is a free log retrieval operation binding the contract event 0x4e56888342f858881f3ddf2560f4c6fca8a9cf54dda777498d5fb44a8b04e8f4.
//
// Solidity: event NoteOperation(bytes32 indexed operationId, address indexed token, uint8 kind, uint256 outputCount, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) FilterNoteOperation(opts *bind.FilterOpts, operationId [][32]byte, token []common.Address) (*InvisiCoreNoteOperationIterator, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "NoteOperation", operationIdRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreNoteOperationIterator{contract: _InvisiCore.contract, event: "NoteOperation", logs: logs, sub: sub}, nil
}

// WatchNoteOperation is a free log subscription operation binding the contract event 0x4e56888342f858881f3ddf2560f4c6fca8a9cf54dda777498d5fb44a8b04e8f4.
//
// Solidity: event NoteOperation(bytes32 indexed operationId, address indexed token, uint8 kind, uint256 outputCount, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) WatchNoteOperation(opts *bind.WatchOpts, sink chan<- *InvisiCoreNoteOperation, operationId [][32]byte, token []common.Address) (event.Subscription, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "NoteOperation", operationIdRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreNoteOperation)
				if err := _InvisiCore.contract.UnpackLog(event, "NoteOperation", log); err != nil {
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

// ParseNoteOperation is a log parse operation binding the contract event 0x4e56888342f858881f3ddf2560f4c6fca8a9cf54dda777498d5fb44a8b04e8f4.
//
// Solidity: event NoteOperation(bytes32 indexed operationId, address indexed token, uint8 kind, uint256 outputCount, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiCore *InvisiCoreFilterer) ParseNoteOperation(log types.Log) (*InvisiCoreNoteOperation, error) {
	event := new(InvisiCoreNoteOperation)
	if err := _InvisiCore.contract.UnpackLog(event, "NoteOperation", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreNoteOutputsIterator is returned from FilterNoteOutputs and is used to iterate over the raw logs and unpacked data for NoteOutputs events raised by the InvisiCore contract.
type InvisiCoreNoteOutputsIterator struct {
	Event *InvisiCoreNoteOutputs // Event containing the contract specifics and raw log

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
func (it *InvisiCoreNoteOutputsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreNoteOutputs)
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
		it.Event = new(InvisiCoreNoteOutputs)
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
func (it *InvisiCoreNoteOutputsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreNoteOutputsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreNoteOutputs represents a NoteOutputs event raised by the InvisiCore contract.
type InvisiCoreNoteOutputs struct {
	OperationId      [32]byte
	TreeId           *big.Int
	StartLeafIndex   uint32
	PackedDeliveries []byte
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNoteOutputs is a free log retrieval operation binding the contract event 0x9d9d69480fdc7e222ae8781287c535f5a441fbbb50e6eb719810e9426bb5b517.
//
// Solidity: event NoteOutputs(bytes32 indexed operationId, uint256 indexed treeId, uint32 startLeafIndex, bytes packedDeliveries)
func (_InvisiCore *InvisiCoreFilterer) FilterNoteOutputs(opts *bind.FilterOpts, operationId [][32]byte, treeId []*big.Int) (*InvisiCoreNoteOutputsIterator, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "NoteOutputs", operationIdRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreNoteOutputsIterator{contract: _InvisiCore.contract, event: "NoteOutputs", logs: logs, sub: sub}, nil
}

// WatchNoteOutputs is a free log subscription operation binding the contract event 0x9d9d69480fdc7e222ae8781287c535f5a441fbbb50e6eb719810e9426bb5b517.
//
// Solidity: event NoteOutputs(bytes32 indexed operationId, uint256 indexed treeId, uint32 startLeafIndex, bytes packedDeliveries)
func (_InvisiCore *InvisiCoreFilterer) WatchNoteOutputs(opts *bind.WatchOpts, sink chan<- *InvisiCoreNoteOutputs, operationId [][32]byte, treeId []*big.Int) (event.Subscription, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "NoteOutputs", operationIdRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreNoteOutputs)
				if err := _InvisiCore.contract.UnpackLog(event, "NoteOutputs", log); err != nil {
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

// ParseNoteOutputs is a log parse operation binding the contract event 0x9d9d69480fdc7e222ae8781287c535f5a441fbbb50e6eb719810e9426bb5b517.
//
// Solidity: event NoteOutputs(bytes32 indexed operationId, uint256 indexed treeId, uint32 startLeafIndex, bytes packedDeliveries)
func (_InvisiCore *InvisiCoreFilterer) ParseNoteOutputs(log types.Log) (*InvisiCoreNoteOutputs, error) {
	event := new(InvisiCoreNoteOutputs)
	if err := _InvisiCore.contract.UnpackLog(event, "NoteOutputs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCoreNoteStateTransitionIterator is returned from FilterNoteStateTransition and is used to iterate over the raw logs and unpacked data for NoteStateTransition events raised by the InvisiCore contract.
type InvisiCoreNoteStateTransitionIterator struct {
	Event *InvisiCoreNoteStateTransition // Event containing the contract specifics and raw log

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
func (it *InvisiCoreNoteStateTransitionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCoreNoteStateTransition)
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
		it.Event = new(InvisiCoreNoteStateTransition)
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
func (it *InvisiCoreNoteStateTransitionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCoreNoteStateTransitionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCoreNoteStateTransition represents a NoteStateTransition event raised by the InvisiCore contract.
type InvisiCoreNoteStateTransition struct {
	OperationId [32]byte
	Sources     []IInvisiCoreInputSource
	Nullifiers  [][32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterNoteStateTransition is a free log retrieval operation binding the contract event 0x255d558968f0a79ff79c4073f67577f4c387247dd50a0510e5cfd3981c174473.
//
// Solidity: event NoteStateTransition(bytes32 indexed operationId, (uint256,bytes32)[] sources, bytes32[] nullifiers)
func (_InvisiCore *InvisiCoreFilterer) FilterNoteStateTransition(opts *bind.FilterOpts, operationId [][32]byte) (*InvisiCoreNoteStateTransitionIterator, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "NoteStateTransition", operationIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCoreNoteStateTransitionIterator{contract: _InvisiCore.contract, event: "NoteStateTransition", logs: logs, sub: sub}, nil
}

// WatchNoteStateTransition is a free log subscription operation binding the contract event 0x255d558968f0a79ff79c4073f67577f4c387247dd50a0510e5cfd3981c174473.
//
// Solidity: event NoteStateTransition(bytes32 indexed operationId, (uint256,bytes32)[] sources, bytes32[] nullifiers)
func (_InvisiCore *InvisiCoreFilterer) WatchNoteStateTransition(opts *bind.WatchOpts, sink chan<- *InvisiCoreNoteStateTransition, operationId [][32]byte) (event.Subscription, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "NoteStateTransition", operationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCoreNoteStateTransition)
				if err := _InvisiCore.contract.UnpackLog(event, "NoteStateTransition", log); err != nil {
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

// ParseNoteStateTransition is a log parse operation binding the contract event 0x255d558968f0a79ff79c4073f67577f4c387247dd50a0510e5cfd3981c174473.
//
// Solidity: event NoteStateTransition(bytes32 indexed operationId, (uint256,bytes32)[] sources, bytes32[] nullifiers)
func (_InvisiCore *InvisiCoreFilterer) ParseNoteStateTransition(log types.Log) (*InvisiCoreNoteStateTransition, error) {
	event := new(InvisiCoreNoteStateTransition)
	if err := _InvisiCore.contract.UnpackLog(event, "NoteStateTransition", log); err != nil {
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
	OperationId [32]byte
	Token       common.Address
	Payer       common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPublicDeposit is a free log retrieval operation binding the contract event 0x9d47b8310e9ff9fddd67b4466b94e05dccd8f877c6e789f2839befb8bad4df92.
//
// Solidity: event PublicDeposit(bytes32 indexed operationId, address indexed token, address indexed payer, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) FilterPublicDeposit(opts *bind.FilterOpts, operationId [][32]byte, token []common.Address, payer []common.Address) (*InvisiCorePublicDepositIterator, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "PublicDeposit", operationIdRule, tokenRule, payerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCorePublicDepositIterator{contract: _InvisiCore.contract, event: "PublicDeposit", logs: logs, sub: sub}, nil
}

// WatchPublicDeposit is a free log subscription operation binding the contract event 0x9d47b8310e9ff9fddd67b4466b94e05dccd8f877c6e789f2839befb8bad4df92.
//
// Solidity: event PublicDeposit(bytes32 indexed operationId, address indexed token, address indexed payer, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) WatchPublicDeposit(opts *bind.WatchOpts, sink chan<- *InvisiCorePublicDeposit, operationId [][32]byte, token []common.Address, payer []common.Address) (event.Subscription, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "PublicDeposit", operationIdRule, tokenRule, payerRule)
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

// ParsePublicDeposit is a log parse operation binding the contract event 0x9d47b8310e9ff9fddd67b4466b94e05dccd8f877c6e789f2839befb8bad4df92.
//
// Solidity: event PublicDeposit(bytes32 indexed operationId, address indexed token, address indexed payer, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) ParsePublicDeposit(log types.Log) (*InvisiCorePublicDeposit, error) {
	event := new(InvisiCorePublicDeposit)
	if err := _InvisiCore.contract.UnpackLog(event, "PublicDeposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiCorePublicWithdrawalIterator is returned from FilterPublicWithdrawal and is used to iterate over the raw logs and unpacked data for PublicWithdrawal events raised by the InvisiCore contract.
type InvisiCorePublicWithdrawalIterator struct {
	Event *InvisiCorePublicWithdrawal // Event containing the contract specifics and raw log

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
func (it *InvisiCorePublicWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiCorePublicWithdrawal)
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
		it.Event = new(InvisiCorePublicWithdrawal)
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
func (it *InvisiCorePublicWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiCorePublicWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiCorePublicWithdrawal represents a PublicWithdrawal event raised by the InvisiCore contract.
type InvisiCorePublicWithdrawal struct {
	OperationId [32]byte
	Token       common.Address
	Recipient   common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPublicWithdrawal is a free log retrieval operation binding the contract event 0x501befef3c73a0262258ef1992e9dd87244ca30eb5c7815b5f0595bfaeb3c242.
//
// Solidity: event PublicWithdrawal(bytes32 indexed operationId, address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) FilterPublicWithdrawal(opts *bind.FilterOpts, operationId [][32]byte, token []common.Address, recipient []common.Address) (*InvisiCorePublicWithdrawalIterator, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiCore.contract.FilterLogs(opts, "PublicWithdrawal", operationIdRule, tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &InvisiCorePublicWithdrawalIterator{contract: _InvisiCore.contract, event: "PublicWithdrawal", logs: logs, sub: sub}, nil
}

// WatchPublicWithdrawal is a free log subscription operation binding the contract event 0x501befef3c73a0262258ef1992e9dd87244ca30eb5c7815b5f0595bfaeb3c242.
//
// Solidity: event PublicWithdrawal(bytes32 indexed operationId, address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) WatchPublicWithdrawal(opts *bind.WatchOpts, sink chan<- *InvisiCorePublicWithdrawal, operationId [][32]byte, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var operationIdRule []interface{}
	for _, operationIdItem := range operationId {
		operationIdRule = append(operationIdRule, operationIdItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiCore.contract.WatchLogs(opts, "PublicWithdrawal", operationIdRule, tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiCorePublicWithdrawal)
				if err := _InvisiCore.contract.UnpackLog(event, "PublicWithdrawal", log); err != nil {
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

// ParsePublicWithdrawal is a log parse operation binding the contract event 0x501befef3c73a0262258ef1992e9dd87244ca30eb5c7815b5f0595bfaeb3c242.
//
// Solidity: event PublicWithdrawal(bytes32 indexed operationId, address indexed token, address indexed recipient, uint256 amount)
func (_InvisiCore *InvisiCoreFilterer) ParsePublicWithdrawal(log types.Log) (*InvisiCorePublicWithdrawal, error) {
	event := new(InvisiCorePublicWithdrawal)
	if err := _InvisiCore.contract.UnpackLog(event, "PublicWithdrawal", log); err != nil {
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
	OperationIds [][32]byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTransactAggregated is a free log retrieval operation binding the contract event 0xbba51cdcc332e4c637d0eb850d94a27af9395d3c79a417bb1615aa894b489e07.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, bytes32[] operationIds)
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

// WatchTransactAggregated is a free log subscription operation binding the contract event 0xbba51cdcc332e4c637d0eb850d94a27af9395d3c79a417bb1615aa894b489e07.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, bytes32[] operationIds)
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

// ParseTransactAggregated is a log parse operation binding the contract event 0xbba51cdcc332e4c637d0eb850d94a27af9395d3c79a417bb1615aa894b489e07.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, bytes32[] operationIds)
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
