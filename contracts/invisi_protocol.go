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

// Struct0 is an auto generated low-level Go binding around an user-defined struct.
type Struct0 struct {
	Factory         common.Address
	Token           common.Address
	GlobalEmergency common.Address
	EmergencyAdmin  common.Address
	TimelockAdmin   common.Address
	MaxPerDeposit   *big.Int
	VerifierRouter  common.Address
}

// Struct2 is an auto generated low-level Go binding around an user-defined struct.
type Struct2 struct {
	Signer      common.Address
	SigNonce    *big.Int
	Deadline    *big.Int
	Signature   []byte
	PermitToken common.Address
	Permit      Struct1
}

// Struct3 is an auto generated low-level Go binding around an user-defined struct.
type Struct3 struct {
	MerkleRoot           [32]byte
	Nullifiers           [][32]byte
	TotalAmount          *big.Int
	ReceivingAddress     common.Address
	TreeId               *big.Int
	RemainderCommitments [][32]byte
	Expiry               *big.Int
}

// Struct1 is an auto generated low-level Go binding around an user-defined struct.
type Struct1 struct {
	Deadline *big.Int
	V        uint8
	R        [32]byte
	S        [32]byte
}

// InvisiProtocolMetaData contains all meta data concerning the InvisiProtocol contract.
var InvisiProtocolMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"params\",\"type\":\"tuple\",\"components\":[{\"name\":\"factory\",\"type\":\"address\"},{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"globalEmergency\",\"type\":\"address\"},{\"name\":\"emergencyAdmin\",\"type\":\"address\"},{\"name\":\"timelockAdmin\",\"type\":\"address\"},{\"name\":\"maxPerDeposit\",\"type\":\"uint256\"},{\"name\":\"verifierRouter\",\"type\":\"address\"}]},{\"name\":\"_screener\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"CHAIN_ID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"COOLDOWN_PERIOD\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEPOSIT_CREDIT_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PAY_BLIND_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PAY_OPEN_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDraw\",\"inputs\":[{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"params\",\"type\":\"tuple\",\"components\":[{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\"},{\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"name\":\"receivingAddress\",\"type\":\"address\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"remainderCommitments\",\"type\":\"bytes32[]\"},{\"name\":\"expiry\",\"type\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"depositCredit\",\"inputs\":[{\"name\":\"identityCommitment\",\"type\":\"bytes32\"},{\"name\":\"payerCommitment\",\"type\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonce\",\"type\":\"bytes32\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"nullifierKeyHash\",\"type\":\"bytes32\"},{\"name\":\"treeId\",\"type\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"depositCreditWithSig\",\"inputs\":[{\"name\":\"identityCommitment\",\"type\":\"bytes32\"},{\"name\":\"payerCommitment\",\"type\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonce\",\"type\":\"bytes32\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"nullifierKeyHash\",\"type\":\"bytes32\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"meta\",\"type\":\"tuple\",\"components\":[{\"name\":\"signer\",\"type\":\"address\"},{\"name\":\"sigNonce\",\"type\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\"},{\"name\":\"permitToken\",\"type\":\"address\"},{\"name\":\"permit\",\"type\":\"tuple\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"version\",\"type\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"emergencyAdmin\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"factory\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPaymentTreeRoot\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTreeCount\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalEmergency\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalPause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"invalidateNonce\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isCommitmentPaid\",\"inputs\":[{\"name\":\"paymentCommitment\",\"type\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isNullifierSpent\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isTreeFull\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxPerDeposit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nonces\",\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"payBlind\",\"inputs\":[{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"paymentCommitment\",\"type\":\"bytes32\"},{\"name\":\"changeCommitment\",\"type\":\"bytes32\"},{\"name\":\"inputAmount\",\"type\":\"uint256\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"refundLink\",\"type\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"payBlindWithSig\",\"inputs\":[{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"paymentCommitment\",\"type\":\"bytes32\"},{\"name\":\"changeCommitment\",\"type\":\"bytes32\"},{\"name\":\"inputAmount\",\"type\":\"uint256\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"refundLink\",\"type\":\"bytes32\"},{\"name\":\"meta\",\"type\":\"tuple\",\"components\":[{\"name\":\"signer\",\"type\":\"address\"},{\"name\":\"sigNonce\",\"type\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\"},{\"name\":\"permitToken\",\"type\":\"address\"},{\"name\":\"permit\",\"type\":\"tuple\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"payOpen\",\"inputs\":[{\"name\":\"identityCommitment\",\"type\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonce\",\"type\":\"bytes32\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"refundLink\",\"type\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"payOpenWithSig\",\"inputs\":[{\"name\":\"identityCommitment\",\"type\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonce\",\"type\":\"bytes32\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"refundLink\",\"type\":\"bytes32\"},{\"name\":\"meta\",\"type\":\"tuple\",\"components\":[{\"name\":\"signer\",\"type\":\"address\"},{\"name\":\"sigNonce\",\"type\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\"},{\"name\":\"permitToken\",\"type\":\"address\"},{\"name\":\"permit\",\"type\":\"tuple\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reclaim\",\"inputs\":[{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"nullifier\",\"type\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"receivingAddress\",\"type\":\"address\"},{\"name\":\"treeId\",\"type\":\"uint256\"},{\"name\":\"expiry\",\"type\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"screener\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setDepositLimit\",\"inputs\":[{\"name\":\"newMaxPerDeposit\",\"type\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setScreener\",\"inputs\":[{\"name\":\"newScreener\",\"type\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"timelockAdmin\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verifierRouter\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdraw\",\"inputs\":[{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\"},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\"},{\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"name\":\"receivingAddress\",\"type\":\"address\"},{\"name\":\"treeId\",\"type\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"CreditDeposit\",\"inputs\":[{\"name\":\"commitment\",\"type\":\"bytes32\",\"indexed\":true},{\"name\":\"depositor\",\"type\":\"address\",\"indexed\":true},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true},{\"name\":\"leafIndex\",\"type\":\"uint32\",\"indexed\":false},{\"name\":\"identityCommitment\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"payerCommitment\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":false},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"nonce\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"expiry\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"nullifierKeyHash\",\"type\":\"bytes32\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditDraw\",\"inputs\":[{\"name\":\"receivingAddress\",\"type\":\"address\",\"indexed\":true},{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":true},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"remainderCommitments\",\"type\":\"bytes32[]\",\"indexed\":false},{\"name\":\"remainderLeafIndices\",\"type\":\"uint32[]\",\"indexed\":false},{\"name\":\"expiry\",\"type\":\"uint256\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditReclaim\",\"inputs\":[{\"name\":\"receivingAddress\",\"type\":\"address\",\"indexed\":true},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true},{\"name\":\"nullifier\",\"type\":\"bytes32\",\"indexed\":true},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":false},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"expiry\",\"type\":\"uint256\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DepositLimitUpdated\",\"inputs\":[{\"name\":\"maxPerDeposit\",\"type\":\"uint256\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NonceInvalidated\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"indexed\":true},{\"name\":\"newNonce\",\"type\":\"uint256\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PayBlind\",\"inputs\":[{\"name\":\"paymentCommitment\",\"type\":\"bytes32\",\"indexed\":true},{\"name\":\"payer\",\"type\":\"address\",\"indexed\":true},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true},{\"name\":\"changeCommitment\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"paymentLeafIndex\",\"type\":\"uint32\",\"indexed\":false},{\"name\":\"changeLeafIndex\",\"type\":\"uint32\",\"indexed\":false},{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":false},{\"name\":\"inputAmount\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"refundLink\",\"type\":\"bytes32\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PayOpen\",\"inputs\":[{\"name\":\"paymentCommitment\",\"type\":\"bytes32\",\"indexed\":true},{\"name\":\"payer\",\"type\":\"address\",\"indexed\":true},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true},{\"name\":\"leafIndex\",\"type\":\"uint32\",\"indexed\":false},{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":false},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false},{\"name\":\"identityCommitment\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"nonce\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"refundLink\",\"type\":\"bytes32\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ScreenerUpdated\",\"inputs\":[{\"name\":\"newScreener\",\"type\":\"address\",\"indexed\":true}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreeCreated\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdraw\",\"inputs\":[{\"name\":\"receivingAddress\",\"type\":\"address\",\"indexed\":true},{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":true},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"indexed\":false},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"indexed\":false}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressScreeningFailed\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChainIdMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CommitmentAlreadyPaid\",\"inputs\":[{\"name\":\"commitment\",\"type\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"CooldownNotPassed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CreditDrawWindowClosed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DoubleSpend\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExceedsDepositLimit\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"maxPerDeposit\",\"type\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpiryInPast\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBatchSize\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidIdentityCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidMerkleRoot\",\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidNonce\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidNullifierKeyHash\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPayerCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidProof\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReceivingAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTree\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"LengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MerkleTreeFull\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAdmin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotEmergencyAdmin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotGlobalEmergency\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NullifierAlreadySpent\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"NullifiersEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}]},{\"type\":\"error\",\"name\":\"SignatureExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StaleRoot\",\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"TransferFailed\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"TreeAlreadyInitialized\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TreeFull\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"TreeNotInitialized\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"WrongPermitToken\",\"inputs\":[{\"name\":\"provided\",\"type\":\"address\"},{\"name\":\"expected\",\"type\":\"address\"}]},{\"type\":\"error\",\"name\":\"ZeroNullifier\",\"inputs\":[]}]",
}

// InvisiProtocolABI is the input ABI used to generate the binding from.
// Deprecated: Use InvisiProtocolMetaData.ABI instead.
var InvisiProtocolABI = InvisiProtocolMetaData.ABI

// InvisiProtocol is an auto generated Go binding around an Ethereum contract.
type InvisiProtocol struct {
	InvisiProtocolCaller     // Read-only binding to the contract
	InvisiProtocolTransactor // Write-only binding to the contract
	InvisiProtocolFilterer   // Log filterer for contract events
}

// InvisiProtocolCaller is an auto generated read-only Go binding around an Ethereum contract.
type InvisiProtocolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiProtocolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InvisiProtocolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiProtocolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InvisiProtocolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InvisiProtocolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InvisiProtocolSession struct {
	Contract     *InvisiProtocol   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// InvisiProtocolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InvisiProtocolCallerSession struct {
	Contract *InvisiProtocolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// InvisiProtocolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InvisiProtocolTransactorSession struct {
	Contract     *InvisiProtocolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// InvisiProtocolRaw is an auto generated low-level Go binding around an Ethereum contract.
type InvisiProtocolRaw struct {
	Contract *InvisiProtocol // Generic contract binding to access the raw methods on
}

// InvisiProtocolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InvisiProtocolCallerRaw struct {
	Contract *InvisiProtocolCaller // Generic read-only contract binding to access the raw methods on
}

// InvisiProtocolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InvisiProtocolTransactorRaw struct {
	Contract *InvisiProtocolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInvisiProtocol creates a new instance of InvisiProtocol, bound to a specific deployed contract.
func NewInvisiProtocol(address common.Address, backend bind.ContractBackend) (*InvisiProtocol, error) {
	contract, err := bindInvisiProtocol(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocol{InvisiProtocolCaller: InvisiProtocolCaller{contract: contract}, InvisiProtocolTransactor: InvisiProtocolTransactor{contract: contract}, InvisiProtocolFilterer: InvisiProtocolFilterer{contract: contract}}, nil
}

// NewInvisiProtocolCaller creates a new read-only instance of InvisiProtocol, bound to a specific deployed contract.
func NewInvisiProtocolCaller(address common.Address, caller bind.ContractCaller) (*InvisiProtocolCaller, error) {
	contract, err := bindInvisiProtocol(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCaller{contract: contract}, nil
}

// NewInvisiProtocolTransactor creates a new write-only instance of InvisiProtocol, bound to a specific deployed contract.
func NewInvisiProtocolTransactor(address common.Address, transactor bind.ContractTransactor) (*InvisiProtocolTransactor, error) {
	contract, err := bindInvisiProtocol(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolTransactor{contract: contract}, nil
}

// NewInvisiProtocolFilterer creates a new log filterer instance of InvisiProtocol, bound to a specific deployed contract.
func NewInvisiProtocolFilterer(address common.Address, filterer bind.ContractFilterer) (*InvisiProtocolFilterer, error) {
	contract, err := bindInvisiProtocol(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolFilterer{contract: contract}, nil
}

// bindInvisiProtocol binds a generic wrapper to an already deployed contract.
func bindInvisiProtocol(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InvisiProtocolMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiProtocol *InvisiProtocolRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiProtocol.Contract.InvisiProtocolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiProtocol *InvisiProtocolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.InvisiProtocolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiProtocol *InvisiProtocolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.InvisiProtocolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_InvisiProtocol *InvisiProtocolCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _InvisiProtocol.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_InvisiProtocol *InvisiProtocolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_InvisiProtocol *InvisiProtocolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.contract.Transact(opts, method, params...)
}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) CHAINID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "CHAIN_ID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) CHAINID() (*big.Int, error) {
	return _InvisiProtocol.Contract.CHAINID(&_InvisiProtocol.CallOpts)
}

// CHAINID is a free data retrieval call binding the contract method 0x85e1f4d0.
//
// Solidity: function CHAIN_ID() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) CHAINID() (*big.Int, error) {
	return _InvisiProtocol.Contract.CHAINID(&_InvisiProtocol.CallOpts)
}

// COOLDOWNPERIOD is a free data retrieval call binding the contract method 0x6e99d52f.
//
// Solidity: function COOLDOWN_PERIOD() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) COOLDOWNPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "COOLDOWN_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// COOLDOWNPERIOD is a free data retrieval call binding the contract method 0x6e99d52f.
//
// Solidity: function COOLDOWN_PERIOD() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) COOLDOWNPERIOD() (*big.Int, error) {
	return _InvisiProtocol.Contract.COOLDOWNPERIOD(&_InvisiProtocol.CallOpts)
}

// COOLDOWNPERIOD is a free data retrieval call binding the contract method 0x6e99d52f.
//
// Solidity: function COOLDOWN_PERIOD() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) COOLDOWNPERIOD() (*big.Int, error) {
	return _InvisiProtocol.Contract.COOLDOWNPERIOD(&_InvisiProtocol.CallOpts)
}

// DEPOSITCREDITTYPEHASH is a free data retrieval call binding the contract method 0xe0c77c2d.
//
// Solidity: function DEPOSIT_CREDIT_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) DEPOSITCREDITTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "DEPOSIT_CREDIT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEPOSITCREDITTYPEHASH is a free data retrieval call binding the contract method 0xe0c77c2d.
//
// Solidity: function DEPOSIT_CREDIT_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) DEPOSITCREDITTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.DEPOSITCREDITTYPEHASH(&_InvisiProtocol.CallOpts)
}

// DEPOSITCREDITTYPEHASH is a free data retrieval call binding the contract method 0xe0c77c2d.
//
// Solidity: function DEPOSIT_CREDIT_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) DEPOSITCREDITTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.DEPOSITCREDITTYPEHASH(&_InvisiProtocol.CallOpts)
}

// PAYBLINDTYPEHASH is a free data retrieval call binding the contract method 0xd597ba47.
//
// Solidity: function PAY_BLIND_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) PAYBLINDTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "PAY_BLIND_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PAYBLINDTYPEHASH is a free data retrieval call binding the contract method 0xd597ba47.
//
// Solidity: function PAY_BLIND_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) PAYBLINDTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.PAYBLINDTYPEHASH(&_InvisiProtocol.CallOpts)
}

// PAYBLINDTYPEHASH is a free data retrieval call binding the contract method 0xd597ba47.
//
// Solidity: function PAY_BLIND_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) PAYBLINDTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.PAYBLINDTYPEHASH(&_InvisiProtocol.CallOpts)
}

// PAYOPENTYPEHASH is a free data retrieval call binding the contract method 0x730337db.
//
// Solidity: function PAY_OPEN_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) PAYOPENTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "PAY_OPEN_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PAYOPENTYPEHASH is a free data retrieval call binding the contract method 0x730337db.
//
// Solidity: function PAY_OPEN_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) PAYOPENTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.PAYOPENTYPEHASH(&_InvisiProtocol.CallOpts)
}

// PAYOPENTYPEHASH is a free data retrieval call binding the contract method 0x730337db.
//
// Solidity: function PAY_OPEN_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) PAYOPENTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.PAYOPENTYPEHASH(&_InvisiProtocol.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) VERSION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) VERSION() (*big.Int, error) {
	return _InvisiProtocol.Contract.VERSION(&_InvisiProtocol.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) VERSION() (*big.Int, error) {
	return _InvisiProtocol.Contract.VERSION(&_InvisiProtocol.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiProtocol *InvisiProtocolCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "eip712Domain")

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
func (_InvisiProtocol *InvisiProtocolSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _InvisiProtocol.Contract.Eip712Domain(&_InvisiProtocol.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_InvisiProtocol *InvisiProtocolCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _InvisiProtocol.Contract.Eip712Domain(&_InvisiProtocol.CallOpts)
}

// EmergencyAdmin is a free data retrieval call binding the contract method 0x70905dce.
//
// Solidity: function emergencyAdmin() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) EmergencyAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "emergencyAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EmergencyAdmin is a free data retrieval call binding the contract method 0x70905dce.
//
// Solidity: function emergencyAdmin() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) EmergencyAdmin() (common.Address, error) {
	return _InvisiProtocol.Contract.EmergencyAdmin(&_InvisiProtocol.CallOpts)
}

// EmergencyAdmin is a free data retrieval call binding the contract method 0x70905dce.
//
// Solidity: function emergencyAdmin() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) EmergencyAdmin() (common.Address, error) {
	return _InvisiProtocol.Contract.EmergencyAdmin(&_InvisiProtocol.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc45a0155.
//
// Solidity: function factory() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) Factory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "factory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Factory is a free data retrieval call binding the contract method 0xc45a0155.
//
// Solidity: function factory() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) Factory() (common.Address, error) {
	return _InvisiProtocol.Contract.Factory(&_InvisiProtocol.CallOpts)
}

// Factory is a free data retrieval call binding the contract method 0xc45a0155.
//
// Solidity: function factory() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) Factory() (common.Address, error) {
	return _InvisiProtocol.Contract.Factory(&_InvisiProtocol.CallOpts)
}

// GetPaymentTreeRoot is a free data retrieval call binding the contract method 0x8588c186.
//
// Solidity: function getPaymentTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) GetPaymentTreeRoot(opts *bind.CallOpts, treeId *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "getPaymentTreeRoot", treeId)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetPaymentTreeRoot is a free data retrieval call binding the contract method 0x8588c186.
//
// Solidity: function getPaymentTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) GetPaymentTreeRoot(treeId *big.Int) ([32]byte, error) {
	return _InvisiProtocol.Contract.GetPaymentTreeRoot(&_InvisiProtocol.CallOpts, treeId)
}

// GetPaymentTreeRoot is a free data retrieval call binding the contract method 0x8588c186.
//
// Solidity: function getPaymentTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) GetPaymentTreeRoot(treeId *big.Int) ([32]byte, error) {
	return _InvisiProtocol.Contract.GetPaymentTreeRoot(&_InvisiProtocol.CallOpts, treeId)
}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256 count)
func (_InvisiProtocol *InvisiProtocolCaller) GetTreeCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "getTreeCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256 count)
func (_InvisiProtocol *InvisiProtocolSession) GetTreeCount() (*big.Int, error) {
	return _InvisiProtocol.Contract.GetTreeCount(&_InvisiProtocol.CallOpts)
}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256 count)
func (_InvisiProtocol *InvisiProtocolCallerSession) GetTreeCount() (*big.Int, error) {
	return _InvisiProtocol.Contract.GetTreeCount(&_InvisiProtocol.CallOpts)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) GlobalEmergency(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "globalEmergency")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) GlobalEmergency() (common.Address, error) {
	return _InvisiProtocol.Contract.GlobalEmergency(&_InvisiProtocol.CallOpts)
}

// GlobalEmergency is a free data retrieval call binding the contract method 0x5befabc2.
//
// Solidity: function globalEmergency() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) GlobalEmergency() (common.Address, error) {
	return _InvisiProtocol.Contract.GlobalEmergency(&_InvisiProtocol.CallOpts)
}

// IsCommitmentPaid is a free data retrieval call binding the contract method 0xe7c3ffe7.
//
// Solidity: function isCommitmentPaid(bytes32 paymentCommitment) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) IsCommitmentPaid(opts *bind.CallOpts, paymentCommitment [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "isCommitmentPaid", paymentCommitment)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCommitmentPaid is a free data retrieval call binding the contract method 0xe7c3ffe7.
//
// Solidity: function isCommitmentPaid(bytes32 paymentCommitment) view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) IsCommitmentPaid(paymentCommitment [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.IsCommitmentPaid(&_InvisiProtocol.CallOpts, paymentCommitment)
}

// IsCommitmentPaid is a free data retrieval call binding the contract method 0xe7c3ffe7.
//
// Solidity: function isCommitmentPaid(bytes32 paymentCommitment) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) IsCommitmentPaid(paymentCommitment [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.IsCommitmentPaid(&_InvisiProtocol.CallOpts, paymentCommitment)
}

// IsNullifierSpent is a free data retrieval call binding the contract method 0xd5a4e325.
//
// Solidity: function isNullifierSpent(bytes32 nullifier) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) IsNullifierSpent(opts *bind.CallOpts, nullifier [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "isNullifierSpent", nullifier)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsNullifierSpent is a free data retrieval call binding the contract method 0xd5a4e325.
//
// Solidity: function isNullifierSpent(bytes32 nullifier) view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) IsNullifierSpent(nullifier [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.IsNullifierSpent(&_InvisiProtocol.CallOpts, nullifier)
}

// IsNullifierSpent is a free data retrieval call binding the contract method 0xd5a4e325.
//
// Solidity: function isNullifierSpent(bytes32 nullifier) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) IsNullifierSpent(nullifier [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.IsNullifierSpent(&_InvisiProtocol.CallOpts, nullifier)
}

// IsTreeFull is a free data retrieval call binding the contract method 0x9fc2199b.
//
// Solidity: function isTreeFull(uint256 treeId) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) IsTreeFull(opts *bind.CallOpts, treeId *big.Int) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "isTreeFull", treeId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTreeFull is a free data retrieval call binding the contract method 0x9fc2199b.
//
// Solidity: function isTreeFull(uint256 treeId) view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) IsTreeFull(treeId *big.Int) (bool, error) {
	return _InvisiProtocol.Contract.IsTreeFull(&_InvisiProtocol.CallOpts, treeId)
}

// IsTreeFull is a free data retrieval call binding the contract method 0x9fc2199b.
//
// Solidity: function isTreeFull(uint256 treeId) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) IsTreeFull(treeId *big.Int) (bool, error) {
	return _InvisiProtocol.Contract.IsTreeFull(&_InvisiProtocol.CallOpts, treeId)
}

// MaxPerDeposit is a free data retrieval call binding the contract method 0x43fe08b0.
//
// Solidity: function maxPerDeposit() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) MaxPerDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "maxPerDeposit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxPerDeposit is a free data retrieval call binding the contract method 0x43fe08b0.
//
// Solidity: function maxPerDeposit() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) MaxPerDeposit() (*big.Int, error) {
	return _InvisiProtocol.Contract.MaxPerDeposit(&_InvisiProtocol.CallOpts)
}

// MaxPerDeposit is a free data retrieval call binding the contract method 0x43fe08b0.
//
// Solidity: function maxPerDeposit() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) MaxPerDeposit() (*big.Int, error) {
	return _InvisiProtocol.Contract.MaxPerDeposit(&_InvisiProtocol.CallOpts)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) Nonces(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "nonces", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _InvisiProtocol.Contract.Nonces(&_InvisiProtocol.CallOpts, arg0)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _InvisiProtocol.Contract.Nonces(&_InvisiProtocol.CallOpts, arg0)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) Paused() (bool, error) {
	return _InvisiProtocol.Contract.Paused(&_InvisiProtocol.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) Paused() (bool, error) {
	return _InvisiProtocol.Contract.Paused(&_InvisiProtocol.CallOpts)
}

// Screener is a free data retrieval call binding the contract method 0x6e0fec70.
//
// Solidity: function screener() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) Screener(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "screener")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Screener is a free data retrieval call binding the contract method 0x6e0fec70.
//
// Solidity: function screener() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) Screener() (common.Address, error) {
	return _InvisiProtocol.Contract.Screener(&_InvisiProtocol.CallOpts)
}

// Screener is a free data retrieval call binding the contract method 0x6e0fec70.
//
// Solidity: function screener() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) Screener() (common.Address, error) {
	return _InvisiProtocol.Contract.Screener(&_InvisiProtocol.CallOpts)
}

// TimelockAdmin is a free data retrieval call binding the contract method 0x64ceb97a.
//
// Solidity: function timelockAdmin() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) TimelockAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "timelockAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TimelockAdmin is a free data retrieval call binding the contract method 0x64ceb97a.
//
// Solidity: function timelockAdmin() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) TimelockAdmin() (common.Address, error) {
	return _InvisiProtocol.Contract.TimelockAdmin(&_InvisiProtocol.CallOpts)
}

// TimelockAdmin is a free data retrieval call binding the contract method 0x64ceb97a.
//
// Solidity: function timelockAdmin() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) TimelockAdmin() (common.Address, error) {
	return _InvisiProtocol.Contract.TimelockAdmin(&_InvisiProtocol.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) Token() (common.Address, error) {
	return _InvisiProtocol.Contract.Token(&_InvisiProtocol.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) Token() (common.Address, error) {
	return _InvisiProtocol.Contract.Token(&_InvisiProtocol.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) VerifierRouter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "verifierRouter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) VerifierRouter() (common.Address, error) {
	return _InvisiProtocol.Contract.VerifierRouter(&_InvisiProtocol.CallOpts)
}

// VerifierRouter is a free data retrieval call binding the contract method 0x849a1586.
//
// Solidity: function verifierRouter() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) VerifierRouter() (common.Address, error) {
	return _InvisiProtocol.Contract.VerifierRouter(&_InvisiProtocol.CallOpts)
}

// CreditDraw is a paid mutator transaction binding the contract method 0x014a29e6.
//
// Solidity: function creditDraw(bytes proof, (bytes32,bytes32[],uint256,address,uint256,bytes32[],uint256) params) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) CreditDraw(opts *bind.TransactOpts, proof []byte, params Struct3) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "creditDraw", proof, params)
}

// CreditDraw is a paid mutator transaction binding the contract method 0x014a29e6.
//
// Solidity: function creditDraw(bytes proof, (bytes32,bytes32[],uint256,address,uint256,bytes32[],uint256) params) returns()
func (_InvisiProtocol *InvisiProtocolSession) CreditDraw(proof []byte, params Struct3) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreditDraw(&_InvisiProtocol.TransactOpts, proof, params)
}

// CreditDraw is a paid mutator transaction binding the contract method 0x014a29e6.
//
// Solidity: function creditDraw(bytes proof, (bytes32,bytes32[],uint256,address,uint256,bytes32[],uint256) params) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) CreditDraw(proof []byte, params Struct3) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreditDraw(&_InvisiProtocol.TransactOpts, proof, params)
}

// DepositCredit is a paid mutator transaction binding the contract method 0x83847d64.
//
// Solidity: function depositCredit(bytes32 identityCommitment, bytes32 payerCommitment, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash, uint256 treeId) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) DepositCredit(opts *bind.TransactOpts, identityCommitment [32]byte, payerCommitment [32]byte, amount *big.Int, nonce [32]byte, expiry *big.Int, nullifierKeyHash [32]byte, treeId *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "depositCredit", identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId)
}

// DepositCredit is a paid mutator transaction binding the contract method 0x83847d64.
//
// Solidity: function depositCredit(bytes32 identityCommitment, bytes32 payerCommitment, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash, uint256 treeId) returns()
func (_InvisiProtocol *InvisiProtocolSession) DepositCredit(identityCommitment [32]byte, payerCommitment [32]byte, amount *big.Int, nonce [32]byte, expiry *big.Int, nullifierKeyHash [32]byte, treeId *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.DepositCredit(&_InvisiProtocol.TransactOpts, identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId)
}

// DepositCredit is a paid mutator transaction binding the contract method 0x83847d64.
//
// Solidity: function depositCredit(bytes32 identityCommitment, bytes32 payerCommitment, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash, uint256 treeId) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) DepositCredit(identityCommitment [32]byte, payerCommitment [32]byte, amount *big.Int, nonce [32]byte, expiry *big.Int, nullifierKeyHash [32]byte, treeId *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.DepositCredit(&_InvisiProtocol.TransactOpts, identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId)
}

// DepositCreditWithSig is a paid mutator transaction binding the contract method 0xedb30e24.
//
// Solidity: function depositCreditWithSig(bytes32 identityCommitment, bytes32 payerCommitment, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash, uint256 treeId, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) DepositCreditWithSig(opts *bind.TransactOpts, identityCommitment [32]byte, payerCommitment [32]byte, amount *big.Int, nonce [32]byte, expiry *big.Int, nullifierKeyHash [32]byte, treeId *big.Int, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "depositCreditWithSig", identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId, meta)
}

// DepositCreditWithSig is a paid mutator transaction binding the contract method 0xedb30e24.
//
// Solidity: function depositCreditWithSig(bytes32 identityCommitment, bytes32 payerCommitment, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash, uint256 treeId, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolSession) DepositCreditWithSig(identityCommitment [32]byte, payerCommitment [32]byte, amount *big.Int, nonce [32]byte, expiry *big.Int, nullifierKeyHash [32]byte, treeId *big.Int, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.DepositCreditWithSig(&_InvisiProtocol.TransactOpts, identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId, meta)
}

// DepositCreditWithSig is a paid mutator transaction binding the contract method 0xedb30e24.
//
// Solidity: function depositCreditWithSig(bytes32 identityCommitment, bytes32 payerCommitment, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash, uint256 treeId, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) DepositCreditWithSig(identityCommitment [32]byte, payerCommitment [32]byte, amount *big.Int, nonce [32]byte, expiry *big.Int, nullifierKeyHash [32]byte, treeId *big.Int, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.DepositCreditWithSig(&_InvisiProtocol.TransactOpts, identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId, meta)
}

// GlobalPause is a paid mutator transaction binding the contract method 0xf12d54d8.
//
// Solidity: function globalPause() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) GlobalPause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "globalPause")
}

// GlobalPause is a paid mutator transaction binding the contract method 0xf12d54d8.
//
// Solidity: function globalPause() returns()
func (_InvisiProtocol *InvisiProtocolSession) GlobalPause() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.GlobalPause(&_InvisiProtocol.TransactOpts)
}

// GlobalPause is a paid mutator transaction binding the contract method 0xf12d54d8.
//
// Solidity: function globalPause() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) GlobalPause() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.GlobalPause(&_InvisiProtocol.TransactOpts)
}

// InvalidateNonce is a paid mutator transaction binding the contract method 0x5a57b46f.
//
// Solidity: function invalidateNonce() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) InvalidateNonce(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "invalidateNonce")
}

// InvalidateNonce is a paid mutator transaction binding the contract method 0x5a57b46f.
//
// Solidity: function invalidateNonce() returns()
func (_InvisiProtocol *InvisiProtocolSession) InvalidateNonce() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.InvalidateNonce(&_InvisiProtocol.TransactOpts)
}

// InvalidateNonce is a paid mutator transaction binding the contract method 0x5a57b46f.
//
// Solidity: function invalidateNonce() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) InvalidateNonce() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.InvalidateNonce(&_InvisiProtocol.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_InvisiProtocol *InvisiProtocolSession) Pause() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Pause(&_InvisiProtocol.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Pause() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Pause(&_InvisiProtocol.TransactOpts)
}

// PayBlind is a paid mutator transaction binding the contract method 0x82d33b9b.
//
// Solidity: function payBlind(bytes proof, bytes32 paymentCommitment, bytes32 changeCommitment, uint256 inputAmount, uint256 treeId, bytes32 refundLink) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) PayBlind(opts *bind.TransactOpts, proof []byte, paymentCommitment [32]byte, changeCommitment [32]byte, inputAmount *big.Int, treeId *big.Int, refundLink [32]byte) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "payBlind", proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink)
}

// PayBlind is a paid mutator transaction binding the contract method 0x82d33b9b.
//
// Solidity: function payBlind(bytes proof, bytes32 paymentCommitment, bytes32 changeCommitment, uint256 inputAmount, uint256 treeId, bytes32 refundLink) returns()
func (_InvisiProtocol *InvisiProtocolSession) PayBlind(proof []byte, paymentCommitment [32]byte, changeCommitment [32]byte, inputAmount *big.Int, treeId *big.Int, refundLink [32]byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayBlind(&_InvisiProtocol.TransactOpts, proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink)
}

// PayBlind is a paid mutator transaction binding the contract method 0x82d33b9b.
//
// Solidity: function payBlind(bytes proof, bytes32 paymentCommitment, bytes32 changeCommitment, uint256 inputAmount, uint256 treeId, bytes32 refundLink) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) PayBlind(proof []byte, paymentCommitment [32]byte, changeCommitment [32]byte, inputAmount *big.Int, treeId *big.Int, refundLink [32]byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayBlind(&_InvisiProtocol.TransactOpts, proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink)
}

// PayBlindWithSig is a paid mutator transaction binding the contract method 0xc172f2db.
//
// Solidity: function payBlindWithSig(bytes proof, bytes32 paymentCommitment, bytes32 changeCommitment, uint256 inputAmount, uint256 treeId, bytes32 refundLink, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) PayBlindWithSig(opts *bind.TransactOpts, proof []byte, paymentCommitment [32]byte, changeCommitment [32]byte, inputAmount *big.Int, treeId *big.Int, refundLink [32]byte, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "payBlindWithSig", proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink, meta)
}

// PayBlindWithSig is a paid mutator transaction binding the contract method 0xc172f2db.
//
// Solidity: function payBlindWithSig(bytes proof, bytes32 paymentCommitment, bytes32 changeCommitment, uint256 inputAmount, uint256 treeId, bytes32 refundLink, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolSession) PayBlindWithSig(proof []byte, paymentCommitment [32]byte, changeCommitment [32]byte, inputAmount *big.Int, treeId *big.Int, refundLink [32]byte, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayBlindWithSig(&_InvisiProtocol.TransactOpts, proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink, meta)
}

// PayBlindWithSig is a paid mutator transaction binding the contract method 0xc172f2db.
//
// Solidity: function payBlindWithSig(bytes proof, bytes32 paymentCommitment, bytes32 changeCommitment, uint256 inputAmount, uint256 treeId, bytes32 refundLink, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) PayBlindWithSig(proof []byte, paymentCommitment [32]byte, changeCommitment [32]byte, inputAmount *big.Int, treeId *big.Int, refundLink [32]byte, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayBlindWithSig(&_InvisiProtocol.TransactOpts, proof, paymentCommitment, changeCommitment, inputAmount, treeId, refundLink, meta)
}

// PayOpen is a paid mutator transaction binding the contract method 0x8cf55984.
//
// Solidity: function payOpen(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) PayOpen(opts *bind.TransactOpts, identityCommitment [32]byte, amount *big.Int, nonce [32]byte, treeId *big.Int, refundLink [32]byte) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "payOpen", identityCommitment, amount, nonce, treeId, refundLink)
}

// PayOpen is a paid mutator transaction binding the contract method 0x8cf55984.
//
// Solidity: function payOpen(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink) returns()
func (_InvisiProtocol *InvisiProtocolSession) PayOpen(identityCommitment [32]byte, amount *big.Int, nonce [32]byte, treeId *big.Int, refundLink [32]byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayOpen(&_InvisiProtocol.TransactOpts, identityCommitment, amount, nonce, treeId, refundLink)
}

// PayOpen is a paid mutator transaction binding the contract method 0x8cf55984.
//
// Solidity: function payOpen(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) PayOpen(identityCommitment [32]byte, amount *big.Int, nonce [32]byte, treeId *big.Int, refundLink [32]byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayOpen(&_InvisiProtocol.TransactOpts, identityCommitment, amount, nonce, treeId, refundLink)
}

// PayOpenWithSig is a paid mutator transaction binding the contract method 0x1c18781e.
//
// Solidity: function payOpenWithSig(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) PayOpenWithSig(opts *bind.TransactOpts, identityCommitment [32]byte, amount *big.Int, nonce [32]byte, treeId *big.Int, refundLink [32]byte, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "payOpenWithSig", identityCommitment, amount, nonce, treeId, refundLink, meta)
}

// PayOpenWithSig is a paid mutator transaction binding the contract method 0x1c18781e.
//
// Solidity: function payOpenWithSig(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolSession) PayOpenWithSig(identityCommitment [32]byte, amount *big.Int, nonce [32]byte, treeId *big.Int, refundLink [32]byte, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayOpenWithSig(&_InvisiProtocol.TransactOpts, identityCommitment, amount, nonce, treeId, refundLink, meta)
}

// PayOpenWithSig is a paid mutator transaction binding the contract method 0x1c18781e.
//
// Solidity: function payOpenWithSig(bytes32 identityCommitment, uint256 amount, bytes32 nonce, uint256 treeId, bytes32 refundLink, (address,uint256,uint256,bytes,address,(uint256,uint8,bytes32,bytes32)) meta) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) PayOpenWithSig(identityCommitment [32]byte, amount *big.Int, nonce [32]byte, treeId *big.Int, refundLink [32]byte, meta Struct2) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PayOpenWithSig(&_InvisiProtocol.TransactOpts, identityCommitment, amount, nonce, treeId, refundLink, meta)
}

// Reclaim is a paid mutator transaction binding the contract method 0xb9ed9a68.
//
// Solidity: function reclaim(bytes proof, bytes32 merkleRoot, bytes32 nullifier, uint256 amount, address receivingAddress, uint256 treeId, uint256 expiry) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Reclaim(opts *bind.TransactOpts, proof []byte, merkleRoot [32]byte, nullifier [32]byte, amount *big.Int, receivingAddress common.Address, treeId *big.Int, expiry *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "reclaim", proof, merkleRoot, nullifier, amount, receivingAddress, treeId, expiry)
}

// Reclaim is a paid mutator transaction binding the contract method 0xb9ed9a68.
//
// Solidity: function reclaim(bytes proof, bytes32 merkleRoot, bytes32 nullifier, uint256 amount, address receivingAddress, uint256 treeId, uint256 expiry) returns()
func (_InvisiProtocol *InvisiProtocolSession) Reclaim(proof []byte, merkleRoot [32]byte, nullifier [32]byte, amount *big.Int, receivingAddress common.Address, treeId *big.Int, expiry *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Reclaim(&_InvisiProtocol.TransactOpts, proof, merkleRoot, nullifier, amount, receivingAddress, treeId, expiry)
}

// Reclaim is a paid mutator transaction binding the contract method 0xb9ed9a68.
//
// Solidity: function reclaim(bytes proof, bytes32 merkleRoot, bytes32 nullifier, uint256 amount, address receivingAddress, uint256 treeId, uint256 expiry) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Reclaim(proof []byte, merkleRoot [32]byte, nullifier [32]byte, amount *big.Int, receivingAddress common.Address, treeId *big.Int, expiry *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Reclaim(&_InvisiProtocol.TransactOpts, proof, merkleRoot, nullifier, amount, receivingAddress, treeId, expiry)
}

// SetDepositLimit is a paid mutator transaction binding the contract method 0xbdc8144b.
//
// Solidity: function setDepositLimit(uint256 newMaxPerDeposit) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) SetDepositLimit(opts *bind.TransactOpts, newMaxPerDeposit *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "setDepositLimit", newMaxPerDeposit)
}

// SetDepositLimit is a paid mutator transaction binding the contract method 0xbdc8144b.
//
// Solidity: function setDepositLimit(uint256 newMaxPerDeposit) returns()
func (_InvisiProtocol *InvisiProtocolSession) SetDepositLimit(newMaxPerDeposit *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.SetDepositLimit(&_InvisiProtocol.TransactOpts, newMaxPerDeposit)
}

// SetDepositLimit is a paid mutator transaction binding the contract method 0xbdc8144b.
//
// Solidity: function setDepositLimit(uint256 newMaxPerDeposit) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) SetDepositLimit(newMaxPerDeposit *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.SetDepositLimit(&_InvisiProtocol.TransactOpts, newMaxPerDeposit)
}

// SetScreener is a paid mutator transaction binding the contract method 0xc2ebd23d.
//
// Solidity: function setScreener(address newScreener) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) SetScreener(opts *bind.TransactOpts, newScreener common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "setScreener", newScreener)
}

// SetScreener is a paid mutator transaction binding the contract method 0xc2ebd23d.
//
// Solidity: function setScreener(address newScreener) returns()
func (_InvisiProtocol *InvisiProtocolSession) SetScreener(newScreener common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.SetScreener(&_InvisiProtocol.TransactOpts, newScreener)
}

// SetScreener is a paid mutator transaction binding the contract method 0xc2ebd23d.
//
// Solidity: function setScreener(address newScreener) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) SetScreener(newScreener common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.SetScreener(&_InvisiProtocol.TransactOpts, newScreener)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_InvisiProtocol *InvisiProtocolSession) Unpause() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Unpause(&_InvisiProtocol.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Unpause() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Unpause(&_InvisiProtocol.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x04723758.
//
// Solidity: function withdraw(bytes proof, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, address receivingAddress, uint256 treeId) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Withdraw(opts *bind.TransactOpts, proof []byte, merkleRoot [32]byte, nullifiers [][32]byte, totalAmount *big.Int, receivingAddress common.Address, treeId *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "withdraw", proof, merkleRoot, nullifiers, totalAmount, receivingAddress, treeId)
}

// Withdraw is a paid mutator transaction binding the contract method 0x04723758.
//
// Solidity: function withdraw(bytes proof, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, address receivingAddress, uint256 treeId) returns()
func (_InvisiProtocol *InvisiProtocolSession) Withdraw(proof []byte, merkleRoot [32]byte, nullifiers [][32]byte, totalAmount *big.Int, receivingAddress common.Address, treeId *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Withdraw(&_InvisiProtocol.TransactOpts, proof, merkleRoot, nullifiers, totalAmount, receivingAddress, treeId)
}

// Withdraw is a paid mutator transaction binding the contract method 0x04723758.
//
// Solidity: function withdraw(bytes proof, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, address receivingAddress, uint256 treeId) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Withdraw(proof []byte, merkleRoot [32]byte, nullifiers [][32]byte, totalAmount *big.Int, receivingAddress common.Address, treeId *big.Int) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Withdraw(&_InvisiProtocol.TransactOpts, proof, merkleRoot, nullifiers, totalAmount, receivingAddress, treeId)
}

// InvisiProtocolCreditDepositIterator is returned from FilterCreditDeposit and is used to iterate over the raw logs and unpacked data for CreditDeposit events raised by the InvisiProtocol contract.
type InvisiProtocolCreditDepositIterator struct {
	Event *InvisiProtocolCreditDeposit // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditDeposit)
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
		it.Event = new(InvisiProtocolCreditDeposit)
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
func (it *InvisiProtocolCreditDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditDeposit represents a CreditDeposit event raised by the InvisiProtocol contract.
type InvisiProtocolCreditDeposit struct {
	Commitment         [32]byte
	Depositor          common.Address
	TreeId             *big.Int
	LeafIndex          uint32
	IdentityCommitment [32]byte
	PayerCommitment    [32]byte
	TokenAddress       common.Address
	Amount             *big.Int
	Nonce              [32]byte
	Expiry             *big.Int
	NullifierKeyHash   [32]byte
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterCreditDeposit is a free log retrieval operation binding the contract event 0x0c71f59fc782f32bb0f185b01ff5e6b5a9449adc86fd74c3c8bd27eac4d25689.
//
// Solidity: event CreditDeposit(bytes32 indexed commitment, address indexed depositor, uint256 indexed treeId, uint32 leafIndex, bytes32 identityCommitment, bytes32 payerCommitment, address tokenAddress, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditDeposit(opts *bind.FilterOpts, commitment [][32]byte, depositor []common.Address, treeId []*big.Int) (*InvisiProtocolCreditDepositIterator, error) {

	var commitmentRule []interface{}
	for _, commitmentItem := range commitment {
		commitmentRule = append(commitmentRule, commitmentItem)
	}
	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditDeposit", commitmentRule, depositorRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditDepositIterator{contract: _InvisiProtocol.contract, event: "CreditDeposit", logs: logs, sub: sub}, nil
}

// WatchCreditDeposit is a free log subscription operation binding the contract event 0x0c71f59fc782f32bb0f185b01ff5e6b5a9449adc86fd74c3c8bd27eac4d25689.
//
// Solidity: event CreditDeposit(bytes32 indexed commitment, address indexed depositor, uint256 indexed treeId, uint32 leafIndex, bytes32 identityCommitment, bytes32 payerCommitment, address tokenAddress, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditDeposit(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditDeposit, commitment [][32]byte, depositor []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var commitmentRule []interface{}
	for _, commitmentItem := range commitment {
		commitmentRule = append(commitmentRule, commitmentItem)
	}
	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditDeposit", commitmentRule, depositorRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditDeposit)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditDeposit", log); err != nil {
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

// ParseCreditDeposit is a log parse operation binding the contract event 0x0c71f59fc782f32bb0f185b01ff5e6b5a9449adc86fd74c3c8bd27eac4d25689.
//
// Solidity: event CreditDeposit(bytes32 indexed commitment, address indexed depositor, uint256 indexed treeId, uint32 leafIndex, bytes32 identityCommitment, bytes32 payerCommitment, address tokenAddress, uint256 amount, bytes32 nonce, uint256 expiry, bytes32 nullifierKeyHash)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditDeposit(log types.Log) (*InvisiProtocolCreditDeposit, error) {
	event := new(InvisiProtocolCreditDeposit)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditDeposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditDrawIterator is returned from FilterCreditDraw and is used to iterate over the raw logs and unpacked data for CreditDraw events raised by the InvisiProtocol contract.
type InvisiProtocolCreditDrawIterator struct {
	Event *InvisiProtocolCreditDraw // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditDrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditDraw)
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
		it.Event = new(InvisiProtocolCreditDraw)
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
func (it *InvisiProtocolCreditDrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditDrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditDraw represents a CreditDraw event raised by the InvisiProtocol contract.
type InvisiProtocolCreditDraw struct {
	ReceivingAddress     common.Address
	TokenAddress         common.Address
	TreeId               *big.Int
	MerkleRoot           [32]byte
	Nullifiers           [][32]byte
	TotalAmount          *big.Int
	RemainderCommitments [][32]byte
	RemainderLeafIndices []uint32
	Expiry               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterCreditDraw is a free log retrieval operation binding the contract event 0xf2f2479f1d8f7e92d095561ab31e3cc4c663278249667f4e9a50563d618e7df5.
//
// Solidity: event CreditDraw(address indexed receivingAddress, address indexed tokenAddress, uint256 indexed treeId, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, bytes32[] remainderCommitments, uint32[] remainderLeafIndices, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditDraw(opts *bind.FilterOpts, receivingAddress []common.Address, tokenAddress []common.Address, treeId []*big.Int) (*InvisiProtocolCreditDrawIterator, error) {

	var receivingAddressRule []interface{}
	for _, receivingAddressItem := range receivingAddress {
		receivingAddressRule = append(receivingAddressRule, receivingAddressItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditDraw", receivingAddressRule, tokenAddressRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditDrawIterator{contract: _InvisiProtocol.contract, event: "CreditDraw", logs: logs, sub: sub}, nil
}

// WatchCreditDraw is a free log subscription operation binding the contract event 0xf2f2479f1d8f7e92d095561ab31e3cc4c663278249667f4e9a50563d618e7df5.
//
// Solidity: event CreditDraw(address indexed receivingAddress, address indexed tokenAddress, uint256 indexed treeId, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, bytes32[] remainderCommitments, uint32[] remainderLeafIndices, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditDraw(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditDraw, receivingAddress []common.Address, tokenAddress []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var receivingAddressRule []interface{}
	for _, receivingAddressItem := range receivingAddress {
		receivingAddressRule = append(receivingAddressRule, receivingAddressItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditDraw", receivingAddressRule, tokenAddressRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditDraw)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditDraw", log); err != nil {
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

// ParseCreditDraw is a log parse operation binding the contract event 0xf2f2479f1d8f7e92d095561ab31e3cc4c663278249667f4e9a50563d618e7df5.
//
// Solidity: event CreditDraw(address indexed receivingAddress, address indexed tokenAddress, uint256 indexed treeId, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount, bytes32[] remainderCommitments, uint32[] remainderLeafIndices, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditDraw(log types.Log) (*InvisiProtocolCreditDraw, error) {
	event := new(InvisiProtocolCreditDraw)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditDraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditReclaimIterator is returned from FilterCreditReclaim and is used to iterate over the raw logs and unpacked data for CreditReclaim events raised by the InvisiProtocol contract.
type InvisiProtocolCreditReclaimIterator struct {
	Event *InvisiProtocolCreditReclaim // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditReclaimIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditReclaim)
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
		it.Event = new(InvisiProtocolCreditReclaim)
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
func (it *InvisiProtocolCreditReclaimIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditReclaimIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditReclaim represents a CreditReclaim event raised by the InvisiProtocol contract.
type InvisiProtocolCreditReclaim struct {
	ReceivingAddress common.Address
	TreeId           *big.Int
	Nullifier        [32]byte
	MerkleRoot       [32]byte
	TokenAddress     common.Address
	Amount           *big.Int
	Expiry           *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCreditReclaim is a free log retrieval operation binding the contract event 0x3b2ee014ba86800a5a9ccb4a6e2056d6e8cf1f4bd6a10cf93db9b00086d5d144.
//
// Solidity: event CreditReclaim(address indexed receivingAddress, uint256 indexed treeId, bytes32 indexed nullifier, bytes32 merkleRoot, address tokenAddress, uint256 amount, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditReclaim(opts *bind.FilterOpts, receivingAddress []common.Address, treeId []*big.Int, nullifier [][32]byte) (*InvisiProtocolCreditReclaimIterator, error) {

	var receivingAddressRule []interface{}
	for _, receivingAddressItem := range receivingAddress {
		receivingAddressRule = append(receivingAddressRule, receivingAddressItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}
	var nullifierRule []interface{}
	for _, nullifierItem := range nullifier {
		nullifierRule = append(nullifierRule, nullifierItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditReclaim", receivingAddressRule, treeIdRule, nullifierRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditReclaimIterator{contract: _InvisiProtocol.contract, event: "CreditReclaim", logs: logs, sub: sub}, nil
}

// WatchCreditReclaim is a free log subscription operation binding the contract event 0x3b2ee014ba86800a5a9ccb4a6e2056d6e8cf1f4bd6a10cf93db9b00086d5d144.
//
// Solidity: event CreditReclaim(address indexed receivingAddress, uint256 indexed treeId, bytes32 indexed nullifier, bytes32 merkleRoot, address tokenAddress, uint256 amount, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditReclaim(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditReclaim, receivingAddress []common.Address, treeId []*big.Int, nullifier [][32]byte) (event.Subscription, error) {

	var receivingAddressRule []interface{}
	for _, receivingAddressItem := range receivingAddress {
		receivingAddressRule = append(receivingAddressRule, receivingAddressItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}
	var nullifierRule []interface{}
	for _, nullifierItem := range nullifier {
		nullifierRule = append(nullifierRule, nullifierItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditReclaim", receivingAddressRule, treeIdRule, nullifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditReclaim)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditReclaim", log); err != nil {
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

// ParseCreditReclaim is a log parse operation binding the contract event 0x3b2ee014ba86800a5a9ccb4a6e2056d6e8cf1f4bd6a10cf93db9b00086d5d144.
//
// Solidity: event CreditReclaim(address indexed receivingAddress, uint256 indexed treeId, bytes32 indexed nullifier, bytes32 merkleRoot, address tokenAddress, uint256 amount, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditReclaim(log types.Log) (*InvisiProtocolCreditReclaim, error) {
	event := new(InvisiProtocolCreditReclaim)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditReclaim", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolDepositLimitUpdatedIterator is returned from FilterDepositLimitUpdated and is used to iterate over the raw logs and unpacked data for DepositLimitUpdated events raised by the InvisiProtocol contract.
type InvisiProtocolDepositLimitUpdatedIterator struct {
	Event *InvisiProtocolDepositLimitUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolDepositLimitUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolDepositLimitUpdated)
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
		it.Event = new(InvisiProtocolDepositLimitUpdated)
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
func (it *InvisiProtocolDepositLimitUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolDepositLimitUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolDepositLimitUpdated represents a DepositLimitUpdated event raised by the InvisiProtocol contract.
type InvisiProtocolDepositLimitUpdated struct {
	MaxPerDeposit *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDepositLimitUpdated is a free log retrieval operation binding the contract event 0xc512617347fd848ec9d7347c99c10e4fa7059132c92d0445930a7fb0c8252ff5.
//
// Solidity: event DepositLimitUpdated(uint256 maxPerDeposit)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterDepositLimitUpdated(opts *bind.FilterOpts) (*InvisiProtocolDepositLimitUpdatedIterator, error) {

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "DepositLimitUpdated")
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolDepositLimitUpdatedIterator{contract: _InvisiProtocol.contract, event: "DepositLimitUpdated", logs: logs, sub: sub}, nil
}

// WatchDepositLimitUpdated is a free log subscription operation binding the contract event 0xc512617347fd848ec9d7347c99c10e4fa7059132c92d0445930a7fb0c8252ff5.
//
// Solidity: event DepositLimitUpdated(uint256 maxPerDeposit)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchDepositLimitUpdated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolDepositLimitUpdated) (event.Subscription, error) {

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "DepositLimitUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolDepositLimitUpdated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "DepositLimitUpdated", log); err != nil {
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

// ParseDepositLimitUpdated is a log parse operation binding the contract event 0xc512617347fd848ec9d7347c99c10e4fa7059132c92d0445930a7fb0c8252ff5.
//
// Solidity: event DepositLimitUpdated(uint256 maxPerDeposit)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseDepositLimitUpdated(log types.Log) (*InvisiProtocolDepositLimitUpdated, error) {
	event := new(InvisiProtocolDepositLimitUpdated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "DepositLimitUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the InvisiProtocol contract.
type InvisiProtocolEIP712DomainChangedIterator struct {
	Event *InvisiProtocolEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolEIP712DomainChanged)
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
		it.Event = new(InvisiProtocolEIP712DomainChanged)
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
func (it *InvisiProtocolEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolEIP712DomainChanged represents a EIP712DomainChanged event raised by the InvisiProtocol contract.
type InvisiProtocolEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiProtocol *InvisiProtocolFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*InvisiProtocolEIP712DomainChangedIterator, error) {

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolEIP712DomainChangedIterator{contract: _InvisiProtocol.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_InvisiProtocol *InvisiProtocolFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *InvisiProtocolEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolEIP712DomainChanged)
				if err := _InvisiProtocol.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseEIP712DomainChanged(log types.Log) (*InvisiProtocolEIP712DomainChanged, error) {
	event := new(InvisiProtocolEIP712DomainChanged)
	if err := _InvisiProtocol.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolNonceInvalidatedIterator is returned from FilterNonceInvalidated and is used to iterate over the raw logs and unpacked data for NonceInvalidated events raised by the InvisiProtocol contract.
type InvisiProtocolNonceInvalidatedIterator struct {
	Event *InvisiProtocolNonceInvalidated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolNonceInvalidatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolNonceInvalidated)
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
		it.Event = new(InvisiProtocolNonceInvalidated)
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
func (it *InvisiProtocolNonceInvalidatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolNonceInvalidatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolNonceInvalidated represents a NonceInvalidated event raised by the InvisiProtocol contract.
type InvisiProtocolNonceInvalidated struct {
	Signer   common.Address
	NewNonce *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNonceInvalidated is a free log retrieval operation binding the contract event 0x1800cd2301fbc20790ed94f3d55a28ef2306a9c31cd3c72b5b71b6e4cf5c6241.
//
// Solidity: event NonceInvalidated(address indexed signer, uint256 newNonce)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterNonceInvalidated(opts *bind.FilterOpts, signer []common.Address) (*InvisiProtocolNonceInvalidatedIterator, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "NonceInvalidated", signerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolNonceInvalidatedIterator{contract: _InvisiProtocol.contract, event: "NonceInvalidated", logs: logs, sub: sub}, nil
}

// WatchNonceInvalidated is a free log subscription operation binding the contract event 0x1800cd2301fbc20790ed94f3d55a28ef2306a9c31cd3c72b5b71b6e4cf5c6241.
//
// Solidity: event NonceInvalidated(address indexed signer, uint256 newNonce)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchNonceInvalidated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolNonceInvalidated, signer []common.Address) (event.Subscription, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "NonceInvalidated", signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolNonceInvalidated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "NonceInvalidated", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseNonceInvalidated(log types.Log) (*InvisiProtocolNonceInvalidated, error) {
	event := new(InvisiProtocolNonceInvalidated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "NonceInvalidated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the InvisiProtocol contract.
type InvisiProtocolPausedIterator struct {
	Event *InvisiProtocolPaused // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolPaused)
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
		it.Event = new(InvisiProtocolPaused)
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
func (it *InvisiProtocolPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolPaused represents a Paused event raised by the InvisiProtocol contract.
type InvisiProtocolPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterPaused(opts *bind.FilterOpts) (*InvisiProtocolPausedIterator, error) {

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolPausedIterator{contract: _InvisiProtocol.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *InvisiProtocolPaused) (event.Subscription, error) {

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolPaused)
				if err := _InvisiProtocol.contract.UnpackLog(event, "Paused", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParsePaused(log types.Log) (*InvisiProtocolPaused, error) {
	event := new(InvisiProtocolPaused)
	if err := _InvisiProtocol.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolPayBlindIterator is returned from FilterPayBlind and is used to iterate over the raw logs and unpacked data for PayBlind events raised by the InvisiProtocol contract.
type InvisiProtocolPayBlindIterator struct {
	Event *InvisiProtocolPayBlind // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolPayBlindIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolPayBlind)
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
		it.Event = new(InvisiProtocolPayBlind)
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
func (it *InvisiProtocolPayBlindIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolPayBlindIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolPayBlind represents a PayBlind event raised by the InvisiProtocol contract.
type InvisiProtocolPayBlind struct {
	PaymentCommitment [32]byte
	Payer             common.Address
	TreeId            *big.Int
	ChangeCommitment  [32]byte
	PaymentLeafIndex  uint32
	ChangeLeafIndex   uint32
	TokenAddress      common.Address
	InputAmount       *big.Int
	RefundLink        [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterPayBlind is a free log retrieval operation binding the contract event 0x83e3d65f2607cd88742c493d815d57ee9b77bc5e777a5988bd3cf1b257a0c9ac.
//
// Solidity: event PayBlind(bytes32 indexed paymentCommitment, address indexed payer, uint256 indexed treeId, bytes32 changeCommitment, uint32 paymentLeafIndex, uint32 changeLeafIndex, address tokenAddress, uint256 inputAmount, bytes32 refundLink)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterPayBlind(opts *bind.FilterOpts, paymentCommitment [][32]byte, payer []common.Address, treeId []*big.Int) (*InvisiProtocolPayBlindIterator, error) {

	var paymentCommitmentRule []interface{}
	for _, paymentCommitmentItem := range paymentCommitment {
		paymentCommitmentRule = append(paymentCommitmentRule, paymentCommitmentItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "PayBlind", paymentCommitmentRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolPayBlindIterator{contract: _InvisiProtocol.contract, event: "PayBlind", logs: logs, sub: sub}, nil
}

// WatchPayBlind is a free log subscription operation binding the contract event 0x83e3d65f2607cd88742c493d815d57ee9b77bc5e777a5988bd3cf1b257a0c9ac.
//
// Solidity: event PayBlind(bytes32 indexed paymentCommitment, address indexed payer, uint256 indexed treeId, bytes32 changeCommitment, uint32 paymentLeafIndex, uint32 changeLeafIndex, address tokenAddress, uint256 inputAmount, bytes32 refundLink)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchPayBlind(opts *bind.WatchOpts, sink chan<- *InvisiProtocolPayBlind, paymentCommitment [][32]byte, payer []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var paymentCommitmentRule []interface{}
	for _, paymentCommitmentItem := range paymentCommitment {
		paymentCommitmentRule = append(paymentCommitmentRule, paymentCommitmentItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "PayBlind", paymentCommitmentRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolPayBlind)
				if err := _InvisiProtocol.contract.UnpackLog(event, "PayBlind", log); err != nil {
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

// ParsePayBlind is a log parse operation binding the contract event 0x83e3d65f2607cd88742c493d815d57ee9b77bc5e777a5988bd3cf1b257a0c9ac.
//
// Solidity: event PayBlind(bytes32 indexed paymentCommitment, address indexed payer, uint256 indexed treeId, bytes32 changeCommitment, uint32 paymentLeafIndex, uint32 changeLeafIndex, address tokenAddress, uint256 inputAmount, bytes32 refundLink)
func (_InvisiProtocol *InvisiProtocolFilterer) ParsePayBlind(log types.Log) (*InvisiProtocolPayBlind, error) {
	event := new(InvisiProtocolPayBlind)
	if err := _InvisiProtocol.contract.UnpackLog(event, "PayBlind", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolPayOpenIterator is returned from FilterPayOpen and is used to iterate over the raw logs and unpacked data for PayOpen events raised by the InvisiProtocol contract.
type InvisiProtocolPayOpenIterator struct {
	Event *InvisiProtocolPayOpen // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolPayOpenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolPayOpen)
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
		it.Event = new(InvisiProtocolPayOpen)
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
func (it *InvisiProtocolPayOpenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolPayOpenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolPayOpen represents a PayOpen event raised by the InvisiProtocol contract.
type InvisiProtocolPayOpen struct {
	PaymentCommitment  [32]byte
	Payer              common.Address
	TreeId             *big.Int
	LeafIndex          uint32
	TokenAddress       common.Address
	Amount             *big.Int
	IdentityCommitment [32]byte
	Nonce              [32]byte
	RefundLink         [32]byte
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterPayOpen is a free log retrieval operation binding the contract event 0x9f10377faeb9f73ccf3fe18c5a599dbea119d1838c47019f33307694f89c3a97.
//
// Solidity: event PayOpen(bytes32 indexed paymentCommitment, address indexed payer, uint256 indexed treeId, uint32 leafIndex, address tokenAddress, uint256 amount, bytes32 identityCommitment, bytes32 nonce, bytes32 refundLink)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterPayOpen(opts *bind.FilterOpts, paymentCommitment [][32]byte, payer []common.Address, treeId []*big.Int) (*InvisiProtocolPayOpenIterator, error) {

	var paymentCommitmentRule []interface{}
	for _, paymentCommitmentItem := range paymentCommitment {
		paymentCommitmentRule = append(paymentCommitmentRule, paymentCommitmentItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "PayOpen", paymentCommitmentRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolPayOpenIterator{contract: _InvisiProtocol.contract, event: "PayOpen", logs: logs, sub: sub}, nil
}

// WatchPayOpen is a free log subscription operation binding the contract event 0x9f10377faeb9f73ccf3fe18c5a599dbea119d1838c47019f33307694f89c3a97.
//
// Solidity: event PayOpen(bytes32 indexed paymentCommitment, address indexed payer, uint256 indexed treeId, uint32 leafIndex, address tokenAddress, uint256 amount, bytes32 identityCommitment, bytes32 nonce, bytes32 refundLink)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchPayOpen(opts *bind.WatchOpts, sink chan<- *InvisiProtocolPayOpen, paymentCommitment [][32]byte, payer []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var paymentCommitmentRule []interface{}
	for _, paymentCommitmentItem := range paymentCommitment {
		paymentCommitmentRule = append(paymentCommitmentRule, paymentCommitmentItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "PayOpen", paymentCommitmentRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolPayOpen)
				if err := _InvisiProtocol.contract.UnpackLog(event, "PayOpen", log); err != nil {
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

// ParsePayOpen is a log parse operation binding the contract event 0x9f10377faeb9f73ccf3fe18c5a599dbea119d1838c47019f33307694f89c3a97.
//
// Solidity: event PayOpen(bytes32 indexed paymentCommitment, address indexed payer, uint256 indexed treeId, uint32 leafIndex, address tokenAddress, uint256 amount, bytes32 identityCommitment, bytes32 nonce, bytes32 refundLink)
func (_InvisiProtocol *InvisiProtocolFilterer) ParsePayOpen(log types.Log) (*InvisiProtocolPayOpen, error) {
	event := new(InvisiProtocolPayOpen)
	if err := _InvisiProtocol.contract.UnpackLog(event, "PayOpen", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolScreenerUpdatedIterator is returned from FilterScreenerUpdated and is used to iterate over the raw logs and unpacked data for ScreenerUpdated events raised by the InvisiProtocol contract.
type InvisiProtocolScreenerUpdatedIterator struct {
	Event *InvisiProtocolScreenerUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolScreenerUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolScreenerUpdated)
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
		it.Event = new(InvisiProtocolScreenerUpdated)
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
func (it *InvisiProtocolScreenerUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolScreenerUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolScreenerUpdated represents a ScreenerUpdated event raised by the InvisiProtocol contract.
type InvisiProtocolScreenerUpdated struct {
	NewScreener common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterScreenerUpdated is a free log retrieval operation binding the contract event 0xb52b505537cdccb66a697cc92b4e51933a403c89ac139bf2297b918de3791dbe.
//
// Solidity: event ScreenerUpdated(address indexed newScreener)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterScreenerUpdated(opts *bind.FilterOpts, newScreener []common.Address) (*InvisiProtocolScreenerUpdatedIterator, error) {

	var newScreenerRule []interface{}
	for _, newScreenerItem := range newScreener {
		newScreenerRule = append(newScreenerRule, newScreenerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "ScreenerUpdated", newScreenerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolScreenerUpdatedIterator{contract: _InvisiProtocol.contract, event: "ScreenerUpdated", logs: logs, sub: sub}, nil
}

// WatchScreenerUpdated is a free log subscription operation binding the contract event 0xb52b505537cdccb66a697cc92b4e51933a403c89ac139bf2297b918de3791dbe.
//
// Solidity: event ScreenerUpdated(address indexed newScreener)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchScreenerUpdated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolScreenerUpdated, newScreener []common.Address) (event.Subscription, error) {

	var newScreenerRule []interface{}
	for _, newScreenerItem := range newScreener {
		newScreenerRule = append(newScreenerRule, newScreenerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "ScreenerUpdated", newScreenerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolScreenerUpdated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "ScreenerUpdated", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseScreenerUpdated(log types.Log) (*InvisiProtocolScreenerUpdated, error) {
	event := new(InvisiProtocolScreenerUpdated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "ScreenerUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolTreeCreatedIterator is returned from FilterTreeCreated and is used to iterate over the raw logs and unpacked data for TreeCreated events raised by the InvisiProtocol contract.
type InvisiProtocolTreeCreatedIterator struct {
	Event *InvisiProtocolTreeCreated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolTreeCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolTreeCreated)
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
		it.Event = new(InvisiProtocolTreeCreated)
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
func (it *InvisiProtocolTreeCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolTreeCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolTreeCreated represents a TreeCreated event raised by the InvisiProtocol contract.
type InvisiProtocolTreeCreated struct {
	TreeId *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTreeCreated is a free log retrieval operation binding the contract event 0x8403033038c2f099c559e73d79138265904c614f5925083ccb2cef7bdeaba28a.
//
// Solidity: event TreeCreated(uint256 indexed treeId)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterTreeCreated(opts *bind.FilterOpts, treeId []*big.Int) (*InvisiProtocolTreeCreatedIterator, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "TreeCreated", treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolTreeCreatedIterator{contract: _InvisiProtocol.contract, event: "TreeCreated", logs: logs, sub: sub}, nil
}

// WatchTreeCreated is a free log subscription operation binding the contract event 0x8403033038c2f099c559e73d79138265904c614f5925083ccb2cef7bdeaba28a.
//
// Solidity: event TreeCreated(uint256 indexed treeId)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchTreeCreated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolTreeCreated, treeId []*big.Int) (event.Subscription, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "TreeCreated", treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolTreeCreated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "TreeCreated", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseTreeCreated(log types.Log) (*InvisiProtocolTreeCreated, error) {
	event := new(InvisiProtocolTreeCreated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "TreeCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the InvisiProtocol contract.
type InvisiProtocolUnpausedIterator struct {
	Event *InvisiProtocolUnpaused // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolUnpaused)
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
		it.Event = new(InvisiProtocolUnpaused)
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
func (it *InvisiProtocolUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolUnpaused represents a Unpaused event raised by the InvisiProtocol contract.
type InvisiProtocolUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterUnpaused(opts *bind.FilterOpts) (*InvisiProtocolUnpausedIterator, error) {

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolUnpausedIterator{contract: _InvisiProtocol.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *InvisiProtocolUnpaused) (event.Subscription, error) {

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolUnpaused)
				if err := _InvisiProtocol.contract.UnpackLog(event, "Unpaused", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseUnpaused(log types.Log) (*InvisiProtocolUnpaused, error) {
	event := new(InvisiProtocolUnpaused)
	if err := _InvisiProtocol.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the InvisiProtocol contract.
type InvisiProtocolWithdrawIterator struct {
	Event *InvisiProtocolWithdraw // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolWithdraw)
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
		it.Event = new(InvisiProtocolWithdraw)
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
func (it *InvisiProtocolWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolWithdraw represents a Withdraw event raised by the InvisiProtocol contract.
type InvisiProtocolWithdraw struct {
	ReceivingAddress common.Address
	TokenAddress     common.Address
	TreeId           *big.Int
	MerkleRoot       [32]byte
	Nullifiers       [][32]byte
	TotalAmount      *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x684a67d57bf8a6a947f940bda94a1bc12de75656034e5a98d0ff8af18d29962a.
//
// Solidity: event Withdraw(address indexed receivingAddress, address indexed tokenAddress, uint256 indexed treeId, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterWithdraw(opts *bind.FilterOpts, receivingAddress []common.Address, tokenAddress []common.Address, treeId []*big.Int) (*InvisiProtocolWithdrawIterator, error) {

	var receivingAddressRule []interface{}
	for _, receivingAddressItem := range receivingAddress {
		receivingAddressRule = append(receivingAddressRule, receivingAddressItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "Withdraw", receivingAddressRule, tokenAddressRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolWithdrawIterator{contract: _InvisiProtocol.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x684a67d57bf8a6a947f940bda94a1bc12de75656034e5a98d0ff8af18d29962a.
//
// Solidity: event Withdraw(address indexed receivingAddress, address indexed tokenAddress, uint256 indexed treeId, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *InvisiProtocolWithdraw, receivingAddress []common.Address, tokenAddress []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var receivingAddressRule []interface{}
	for _, receivingAddressItem := range receivingAddress {
		receivingAddressRule = append(receivingAddressRule, receivingAddressItem)
	}
	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}
	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "Withdraw", receivingAddressRule, tokenAddressRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolWithdraw)
				if err := _InvisiProtocol.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x684a67d57bf8a6a947f940bda94a1bc12de75656034e5a98d0ff8af18d29962a.
//
// Solidity: event Withdraw(address indexed receivingAddress, address indexed tokenAddress, uint256 indexed treeId, bytes32 merkleRoot, bytes32[] nullifiers, uint256 totalAmount)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseWithdraw(log types.Log) (*InvisiProtocolWithdraw, error) {
	event := new(InvisiProtocolWithdraw)
	if err := _InvisiProtocol.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
