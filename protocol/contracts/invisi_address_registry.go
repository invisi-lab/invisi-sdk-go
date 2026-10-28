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

// IAddressRegistryReceivingKeys is an auto generated low-level Go binding around an user-defined struct.
type IAddressRegistryReceivingKeys struct {
	ViewingPublicKey []byte
	ClueKey          [32]byte
}

// IAddressRegistryRegistration is an auto generated low-level Go binding around an user-defined struct.
type IAddressRegistryRegistration struct {
	WalletIdentity [32]byte
	Keys           IAddressRegistryReceivingKeys
	Nonce          uint64
	Deadline       uint64
}

// InvisiAddressRegistryMetaData contains all meta data concerning the InvisiAddressRegistry contract.
var InvisiAddressRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"CHAIN_ID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isRegistered\",\"inputs\":[{\"name\":\"walletIdentity\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nonceOf\",\"inputs\":[{\"name\":\"walletIdentity\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"receivingKeysOf\",\"inputs\":[{\"name\":\"walletIdentity\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"keys\",\"type\":\"tuple\",\"internalType\":\"structIAddressRegistry.ReceivingKeys\",\"components\":[{\"name\":\"viewingPublicKey\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clueKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerKeys\",\"inputs\":[{\"name\":\"registration\",\"type\":\"tuple\",\"internalType\":\"structIAddressRegistry.Registration\",\"components\":[{\"name\":\"walletIdentity\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"keys\",\"type\":\"tuple\",\"internalType\":\"structIAddressRegistry.ReceivingKeys\",\"components\":[{\"name\":\"viewingPublicKey\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clueKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"deadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"updateKeys\",\"inputs\":[{\"name\":\"registration\",\"type\":\"tuple\",\"internalType\":\"structIAddressRegistry.Registration\",\"components\":[{\"name\":\"walletIdentity\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"keys\",\"type\":\"tuple\",\"internalType\":\"structIAddressRegistry.ReceivingKeys\",\"components\":[{\"name\":\"viewingPublicKey\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clueKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"deadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReceivingKeysSet\",\"inputs\":[{\"name\":\"walletIdentity\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"viewingPublicKey\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"},{\"name\":\"clueKey\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"nextNonce\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ChainIdMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidNonce\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReceivingKeys\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRegistrationSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidWalletIdentity\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NonceExhausted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RegistrationExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]}]",
}

// InvisiAddressRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use InvisiAddressRegistryMetaData.ABI instead.
var InvisiAddressRegistryABI = InvisiAddressRegistryMetaData.ABI

// InvisiAddressRegistry is an auto generated Go binding around an Ethereum contract.
type InvisiAddressRegistry struct {
	InvisiAddressRegistryCaller     // Read-only binding to the contract
	InvisiAddressRegistryTransactor // Write-only binding to the contract
	InvisiAddressRegistryFilterer   // Log filterer for contract events
}

// InvisiAddressRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type InvisiAddressRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiAddressRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InvisiAddressRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiAddressRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InvisiAddressRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiAddressRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InvisiAddressRegistrySession struct {
	Contract     *InvisiAddressRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// InvisiAddressRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InvisiAddressRegistryCallerSession struct {
	Contract *InvisiAddressRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// InvisiAddressRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InvisiAddressRegistryTransactorSession struct {
	Contract     *InvisiAddressRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// InvisiAddressRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type InvisiAddressRegistryRaw struct {
	Contract *InvisiAddressRegistry // Generic contract binding to access the raw methods on
}

// InvisiAddressRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InvisiAddressRegistryCallerRaw struct {
	Contract *InvisiAddressRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// InvisiAddressRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InvisiAddressRegistryTransactorRaw struct {
	Contract *InvisiAddressRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInvisiAddressRegistry creates a new instance of InvisiAddressRegistry, bound to a specific deployed contract.
func NewInvisiAddressRegistry(address common.Address, backend bind.ContractBackend) (*InvisiAddressRegistry, error) {
	contract, err := bindInvisiAddressRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InvisiAddressRegistry{InvisiAddressRegistryCaller: InvisiAddressRegistryCaller{contract: contract}, InvisiAddressRegistryTransactor: InvisiAddressRegistryTransactor{contract: contract}, InvisiAddressRegistryFilterer: InvisiAddressRegistryFilterer{contract: contract}}, nil
}

// NewInvisiAddressRegistryCaller creates a new read-only instance of InvisiAddressRegistry, bound to a specific deployed contract.
func NewInvisiAddressRegistryCaller(address common.Address, caller bind.ContractCaller) (*InvisiAddressRegistryCaller, error) {
	contract, err := bindInvisiAddressRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiAddressRegistryCaller{contract: contract}, nil
}

// NewInvisiAddressRegistryTransactor creates a new write-only instance of InvisiAddressRegistry, bound to a specific deployed contract.
func NewInvisiAddressRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*InvisiAddressRegistryTransactor, error) {
	contract, err := bindInvisiAddressRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiAddressRegistryTransactor{contract: contract}, nil
}

// NewInvisiAddressRegistryFilterer creates a new log filterer instance of InvisiAddressRegistry, bound to a specific deployed contract.
func NewInvisiAddressRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*InvisiAddressRegistryFilterer, error) {
	contract, err := bindInvisiAddressRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InvisiAddressRegistryFilterer{contract: contract}, nil
}

// bindInvisiAddressRegistry binds a generic wrapper to an already deployed contract.
func bindInvisiAddressRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiAddressRegistry *InvisiAddressRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiAddressRegistry.Contract.InvisiAddressRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiAddressRegistry *InvisiAddressRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.InvisiAddressRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiAddressRegistry *InvisiAddressRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.InvisiAddressRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiAddressRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiAddressRegistry *InvisiAddressRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiAddressRegistry *InvisiAddressRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.contract.Transact(opts, method, params...)
}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiAddressRegistry *InvisiAddressRegistryCaller) CHAINID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiAddressRegistry.contract.Call(opts, &out, "CHAIN_ID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) CHAINID() (*big.Int, error) {
	return _InvisiAddressRegistry.Contract.CHAINID(&_InvisiAddressRegistry.CallOpts)
}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerSession) CHAINID() (*big.Int, error) {
	return _InvisiAddressRegistry.Contract.CHAINID(&_InvisiAddressRegistry.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiAddressRegistry *InvisiAddressRegistryCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _InvisiAddressRegistry.contract.Call(opts, &out, "eip712Domain")

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
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _InvisiAddressRegistry.Contract.Eip712Domain(&_InvisiAddressRegistry.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _InvisiAddressRegistry.Contract.Eip712Domain(&_InvisiAddressRegistry.CallOpts)
}

// IsRegistered is a free data retrieval call binding the contract method 0x27258b22.
//
// Solidity: function isRegistered(bytes32 walletIdentity) view returns(bool)
func (_InvisiAddressRegistry *InvisiAddressRegistryCaller) IsRegistered(opts *bind.CallOpts, walletIdentity [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiAddressRegistry.contract.Call(opts, &out, "isRegistered", walletIdentity)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRegistered is a free data retrieval call binding the contract method 0x27258b22.
//
// Solidity: function isRegistered(bytes32 walletIdentity) view returns(bool)
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) IsRegistered(walletIdentity [32]byte) (bool, error) {
	return _InvisiAddressRegistry.Contract.IsRegistered(&_InvisiAddressRegistry.CallOpts, walletIdentity)
}

// IsRegistered is a free data retrieval call binding the contract method 0x27258b22.
//
// Solidity: function isRegistered(bytes32 walletIdentity) view returns(bool)
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerSession) IsRegistered(walletIdentity [32]byte) (bool, error) {
	return _InvisiAddressRegistry.Contract.IsRegistered(&_InvisiAddressRegistry.CallOpts, walletIdentity)
}

// NonceOf is a free data retrieval call binding the contract method 0x59f3ec16.
//
// Solidity: function nonceOf(bytes32 walletIdentity) view returns(uint64)
func (_InvisiAddressRegistry *InvisiAddressRegistryCaller) NonceOf(opts *bind.CallOpts, walletIdentity [32]byte) (uint64, error) {
	var out []interface{}
	err := _InvisiAddressRegistry.contract.Call(opts, &out, "nonceOf", walletIdentity)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// NonceOf is a free data retrieval call binding the contract method 0x59f3ec16.
//
// Solidity: function nonceOf(bytes32 walletIdentity) view returns(uint64)
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) NonceOf(walletIdentity [32]byte) (uint64, error) {
	return _InvisiAddressRegistry.Contract.NonceOf(&_InvisiAddressRegistry.CallOpts, walletIdentity)
}

// NonceOf is a free data retrieval call binding the contract method 0x59f3ec16.
//
// Solidity: function nonceOf(bytes32 walletIdentity) view returns(uint64)
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerSession) NonceOf(walletIdentity [32]byte) (uint64, error) {
	return _InvisiAddressRegistry.Contract.NonceOf(&_InvisiAddressRegistry.CallOpts, walletIdentity)
}

// ReceivingKeysOf is a free data retrieval call binding the contract method 0xac822250.
//
// Solidity: function receivingKeysOf(bytes32 walletIdentity) view returns((bytes,bytes32) keys, uint64 nonce)
func (_InvisiAddressRegistry *InvisiAddressRegistryCaller) ReceivingKeysOf(opts *bind.CallOpts, walletIdentity [32]byte) (struct {
	Keys  IAddressRegistryReceivingKeys
	Nonce uint64
}, error) {
	var out []interface{}
	err := _InvisiAddressRegistry.contract.Call(opts, &out, "receivingKeysOf", walletIdentity)

	outstruct := new(struct {
		Keys  IAddressRegistryReceivingKeys
		Nonce uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Keys = *abi.ConvertType(out[0], new(IAddressRegistryReceivingKeys)).(*IAddressRegistryReceivingKeys)
	outstruct.Nonce = *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return *outstruct, err

}

// ReceivingKeysOf is a free data retrieval call binding the contract method 0xac822250.
//
// Solidity: function receivingKeysOf(bytes32 walletIdentity) view returns((bytes,bytes32) keys, uint64 nonce)
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) ReceivingKeysOf(walletIdentity [32]byte) (struct {
	Keys  IAddressRegistryReceivingKeys
	Nonce uint64
}, error) {
	return _InvisiAddressRegistry.Contract.ReceivingKeysOf(&_InvisiAddressRegistry.CallOpts, walletIdentity)
}

// ReceivingKeysOf is a free data retrieval call binding the contract method 0xac822250.
//
// Solidity: function receivingKeysOf(bytes32 walletIdentity) view returns((bytes,bytes32) keys, uint64 nonce)
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerSession) ReceivingKeysOf(walletIdentity [32]byte) (struct {
	Keys  IAddressRegistryReceivingKeys
	Nonce uint64
}, error) {
	return _InvisiAddressRegistry.Contract.ReceivingKeysOf(&_InvisiAddressRegistry.CallOpts, walletIdentity)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_InvisiAddressRegistry *InvisiAddressRegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _InvisiAddressRegistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _InvisiAddressRegistry.Contract.SupportsInterface(&_InvisiAddressRegistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_InvisiAddressRegistry *InvisiAddressRegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _InvisiAddressRegistry.Contract.SupportsInterface(&_InvisiAddressRegistry.CallOpts, interfaceId)
}

// RegisterKeys is a paid mutator transaction binding the contract method 0x0d116606.
//
// Solidity: function registerKeys((bytes32,(bytes,bytes32),uint64,uint64) registration, bytes signature) returns()
func (_InvisiAddressRegistry *InvisiAddressRegistryTransactor) RegisterKeys(opts *bind.TransactOpts, registration IAddressRegistryRegistration, signature []byte) (*types.Transaction, error) {
	return _InvisiAddressRegistry.contract.Transact(opts, "registerKeys", registration, signature)
}

// RegisterKeys is a paid mutator transaction binding the contract method 0x0d116606.
//
// Solidity: function registerKeys((bytes32,(bytes,bytes32),uint64,uint64) registration, bytes signature) returns()
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) RegisterKeys(registration IAddressRegistryRegistration, signature []byte) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.RegisterKeys(&_InvisiAddressRegistry.TransactOpts, registration, signature)
}

// RegisterKeys is a paid mutator transaction binding the contract method 0x0d116606.
//
// Solidity: function registerKeys((bytes32,(bytes,bytes32),uint64,uint64) registration, bytes signature) returns()
func (_InvisiAddressRegistry *InvisiAddressRegistryTransactorSession) RegisterKeys(registration IAddressRegistryRegistration, signature []byte) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.RegisterKeys(&_InvisiAddressRegistry.TransactOpts, registration, signature)
}

// UpdateKeys is a paid mutator transaction binding the contract method 0x75971d55.
//
// Solidity: function updateKeys((bytes32,(bytes,bytes32),uint64,uint64) registration, bytes signature) returns()
func (_InvisiAddressRegistry *InvisiAddressRegistryTransactor) UpdateKeys(opts *bind.TransactOpts, registration IAddressRegistryRegistration, signature []byte) (*types.Transaction, error) {
	return _InvisiAddressRegistry.contract.Transact(opts, "updateKeys", registration, signature)
}

// UpdateKeys is a paid mutator transaction binding the contract method 0x75971d55.
//
// Solidity: function updateKeys((bytes32,(bytes,bytes32),uint64,uint64) registration, bytes signature) returns()
func (_InvisiAddressRegistry *InvisiAddressRegistrySession) UpdateKeys(registration IAddressRegistryRegistration, signature []byte) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.UpdateKeys(&_InvisiAddressRegistry.TransactOpts, registration, signature)
}

// UpdateKeys is a paid mutator transaction binding the contract method 0x75971d55.
//
// Solidity: function updateKeys((bytes32,(bytes,bytes32),uint64,uint64) registration, bytes signature) returns()
func (_InvisiAddressRegistry *InvisiAddressRegistryTransactorSession) UpdateKeys(registration IAddressRegistryRegistration, signature []byte) (*types.Transaction, error) {
	return _InvisiAddressRegistry.Contract.UpdateKeys(&_InvisiAddressRegistry.TransactOpts, registration, signature)
}

// InvisiAddressRegistryEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the InvisiAddressRegistry contract.
type InvisiAddressRegistryEIP712DomainChangedIterator struct {
	Event *InvisiAddressRegistryEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *InvisiAddressRegistryEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiAddressRegistryEIP712DomainChanged)
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
		it.Event = new(InvisiAddressRegistryEIP712DomainChanged)
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
func (it *InvisiAddressRegistryEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiAddressRegistryEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiAddressRegistryEIP712DomainChanged represents a EIP712DomainChanged event raised by the InvisiAddressRegistry contract.
type InvisiAddressRegistryEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiAddressRegistry *InvisiAddressRegistryFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*InvisiAddressRegistryEIP712DomainChangedIterator, error) {

	logs, sub, err := _InvisiAddressRegistry.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &InvisiAddressRegistryEIP712DomainChangedIterator{contract: _InvisiAddressRegistry.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiAddressRegistry *InvisiAddressRegistryFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *InvisiAddressRegistryEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _InvisiAddressRegistry.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiAddressRegistryEIP712DomainChanged)
				if err := _InvisiAddressRegistry.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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
func (_InvisiAddressRegistry *InvisiAddressRegistryFilterer) ParseEIP712DomainChanged(log types.Log) (*InvisiAddressRegistryEIP712DomainChanged, error) {
	event := new(InvisiAddressRegistryEIP712DomainChanged)
	if err := _InvisiAddressRegistry.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiAddressRegistryReceivingKeysSetIterator is returned from FilterReceivingKeysSet and is used to iterate over the raw logs and unpacked data for ReceivingKeysSet events raised by the InvisiAddressRegistry contract.
type InvisiAddressRegistryReceivingKeysSetIterator struct {
	Event *InvisiAddressRegistryReceivingKeysSet // Event containing the contract specifics and raw log

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
func (it *InvisiAddressRegistryReceivingKeysSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiAddressRegistryReceivingKeysSet)
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
		it.Event = new(InvisiAddressRegistryReceivingKeysSet)
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
func (it *InvisiAddressRegistryReceivingKeysSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiAddressRegistryReceivingKeysSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiAddressRegistryReceivingKeysSet represents a ReceivingKeysSet event raised by the InvisiAddressRegistry contract.
type InvisiAddressRegistryReceivingKeysSet struct {
	WalletIdentity   [32]byte
	ViewingPublicKey []byte
	ClueKey          [32]byte
	NextNonce        uint64
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterReceivingKeysSet is a free log retrieval operation binding the contract event 0xa723f8c7b539ea283743c52ddd42ad3397b62c55bde6d956ee62637815231147.
//
// Solidity: event ReceivingKeysSet(bytes32 indexed walletIdentity, bytes viewingPublicKey, bytes32 clueKey, uint64 nextNonce)
func (_InvisiAddressRegistry *InvisiAddressRegistryFilterer) FilterReceivingKeysSet(opts *bind.FilterOpts, walletIdentity [][32]byte) (*InvisiAddressRegistryReceivingKeysSetIterator, error) {

	var walletIdentityRule []interface{}
	for _, walletIdentityItem := range walletIdentity {
		walletIdentityRule = append(walletIdentityRule, walletIdentityItem)
	}

	logs, sub, err := _InvisiAddressRegistry.contract.FilterLogs(opts, "ReceivingKeysSet", walletIdentityRule)
	if err != nil {
		return nil, err
	}
	return &InvisiAddressRegistryReceivingKeysSetIterator{contract: _InvisiAddressRegistry.contract, event: "ReceivingKeysSet", logs: logs, sub: sub}, nil
}

// WatchReceivingKeysSet is a free log subscription operation binding the contract event 0xa723f8c7b539ea283743c52ddd42ad3397b62c55bde6d956ee62637815231147.
//
// Solidity: event ReceivingKeysSet(bytes32 indexed walletIdentity, bytes viewingPublicKey, bytes32 clueKey, uint64 nextNonce)
func (_InvisiAddressRegistry *InvisiAddressRegistryFilterer) WatchReceivingKeysSet(opts *bind.WatchOpts, sink chan<- *InvisiAddressRegistryReceivingKeysSet, walletIdentity [][32]byte) (event.Subscription, error) {

	var walletIdentityRule []interface{}
	for _, walletIdentityItem := range walletIdentity {
		walletIdentityRule = append(walletIdentityRule, walletIdentityItem)
	}

	logs, sub, err := _InvisiAddressRegistry.contract.WatchLogs(opts, "ReceivingKeysSet", walletIdentityRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiAddressRegistryReceivingKeysSet)
				if err := _InvisiAddressRegistry.contract.UnpackLog(event, "ReceivingKeysSet", log); err != nil {
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

// ParseReceivingKeysSet is a log parse operation binding the contract event 0xa723f8c7b539ea283743c52ddd42ad3397b62c55bde6d956ee62637815231147.
//
// Solidity: event ReceivingKeysSet(bytes32 indexed walletIdentity, bytes viewingPublicKey, bytes32 clueKey, uint64 nextNonce)
func (_InvisiAddressRegistry *InvisiAddressRegistryFilterer) ParseReceivingKeysSet(log types.Log) (*InvisiAddressRegistryReceivingKeysSet, error) {
	event := new(InvisiAddressRegistryReceivingKeysSet)
	if err := _InvisiAddressRegistry.contract.UnpackLog(event, "ReceivingKeysSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
