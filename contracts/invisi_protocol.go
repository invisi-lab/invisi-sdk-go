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

// IInvisiMetaTxTokenPermit is an auto generated low-level Go binding around an user-defined struct.
type IInvisiMetaTxTokenPermit struct {
	Deadline *big.Int
	V        uint8
	R        [32]byte
	S        [32]byte
}

// IInvisiProtocolCancelCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCancelCreditParams struct {
	Source    IInvisiProtocolInputSource
	Nullifier [32]byte
	Owner     common.Address
	Amount    *big.Int
	Deadline  *big.Int
}

// IInvisiProtocolCreateCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreateCreditParams struct {
	Sources          []IInvisiProtocolInputSource
	Nullifiers       [][32]byte
	CreditCommitment [32]byte
	ChangeCommitment [32]byte
	Owner            common.Address
	Expiry           *big.Int
	Deadline         *big.Int
}

// IInvisiProtocolCreditAuthorization is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreditAuthorization struct {
	SigNonce  *big.Int
	Signature []byte
}

// IInvisiProtocolCreditDrawParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreditDrawParams struct {
	Sources              []IInvisiProtocolInputSource
	Nullifiers           [][32]byte
	RemainderCommitments [][32]byte
	Expiries             []*big.Int
	RecoveryAmounts      [][32]byte
	Spender              common.Address
	TotalAmount          *big.Int
	Deadline             *big.Int
}

// IInvisiProtocolCrossTreeTransactParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCrossTreeTransactParams struct {
	Sources           []IInvisiProtocolInputSource
	Nullifiers        [][32]byte
	OutputCommitments [][32]byte
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
}

// IInvisiProtocolFundingParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolFundingParams struct {
	Payer     common.Address
	Amount    *big.Int
	SigNonce  *big.Int
	Deadline  *big.Int
	Signature []byte
	Permit    IInvisiMetaTxTokenPermit
}

// IInvisiProtocolInputSource is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolInputSource struct {
	TreeId     *big.Int
	MerkleRoot [32]byte
}

// IInvisiProtocolShieldOutput is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolShieldOutput struct {
	IdentityCommitment [32]byte
	Amount             *big.Int
	Nonce              [32]byte
}

// IInvisiProtocolShieldParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolShieldParams struct {
	OutputCommitments [][32]byte
	Deadline          *big.Int
}

// IInvisiProtocolTopUpCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolTopUpCreditParams struct {
	Creation        IInvisiProtocolCreateCreditParams
	CreditSource    IInvisiProtocolInputSource
	CreditNullifier [32]byte
	PreviousExpiry  *big.Int
}

// IInvisiProtocolTransactRequest is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolTransactRequest struct {
	Params      IInvisiProtocolCrossTreeTransactParams
	Ciphertexts [][]byte
	Funding     IInvisiProtocolFundingParams
}

// InstanceParams is an auto generated low-level Go binding around an user-defined struct.
type InstanceParams struct {
	Factory         common.Address
	Token           common.Address
	GlobalEmergency common.Address
	EmergencyAdmin  common.Address
	TimelockAdmin   common.Address
	MaxPerDeposit   *big.Int
	VerifierRouter  common.Address
}

// InvisiProtocolMetaData contains all meta data concerning the InvisiProtocol contract.
var InvisiProtocolMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structInstanceParams\",\"components\":[{\"name\":\"factory\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"globalEmergency\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"emergencyAdmin\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timelockAdmin\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"maxPerDeposit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifierRouter\",\"type\":\"address\",\"internalType\":\"contractIVerifierRouter\"}]},{\"name\":\"_screener\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"CHAIN_ID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"FUND_SHIELD_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"FUND_TRANSACT_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"activeNoteTreeId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cancelCredit\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CancelCreditParams\",\"components\":[{\"name\":\"source\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.InputSource\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreditAuthorization\",\"components\":[{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createCredit\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreateCreditParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"changeCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIInvisiMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreditAuthorization\",\"components\":[{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditDraw\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreditDrawParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"remainderCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"expiries\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"recoveryAmounts\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"emergencyAdmin\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"factory\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNoteTreeNextLeafIndex\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNoteTreeRoot\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTreeCount\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalEmergency\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalPause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"invalidateNonce\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isCommitmentPaid\",\"inputs\":[{\"name\":\"noteCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isNullifierSpent\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isTreeFull\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxPerDeposit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nonces\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rootHistory\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"screener\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setDepositLimit\",\"inputs\":[{\"name\":\"newMaxPerDeposit\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setScreener\",\"inputs\":[{\"name\":\"newScreener\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"shield\",\"inputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.ShieldParams\",\"components\":[{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.ShieldOutput[]\",\"components\":[{\"name\":\"identityCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nonce\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIInvisiMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"timelockAdmin\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"topUpCredit\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.TopUpCreditParams\",\"components\":[{\"name\":\"creation\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreateCreditParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"changeCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"creditSource\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.InputSource\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"creditNullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIInvisiMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreditAuthorization\",\"components\":[{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transact\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.TransactRequest\",\"components\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CrossTreeTransactParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"funding\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIInvisiMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transact\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.TransactRequest[]\",\"components\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CrossTreeTransactParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"},{\"name\":\"funding\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIInvisiMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verifierRouter\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifierRouter\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"BatchOutputs\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"startLeafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"firstOutputOffset\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"commitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"packedCiphertexts\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"BatchRequest\",\"inputs\":[{\"name\":\"requestIndex\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"capacity\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"outputCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payer\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"depositAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"statementHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CommitmentsInserted\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"startLeafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"commitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditCancelled\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"inputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"nullifier\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditCreated\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"changeCommitment\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditDraw\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"params\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.CreditDrawParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"remainderCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"expiries\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"recoveryAmounts\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"remainderLeafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditToppedUp\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"params\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.TopUpCreditParams\",\"components\":[{\"name\":\"creation\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreateCreditParams\",\"components\":[{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"changeCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"creditSource\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.InputSource\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"creditNullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"previousExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CrossTreeTransact\",\"inputs\":[{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DepositLimitUpdated\",\"inputs\":[{\"name\":\"maxPerDeposit\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NonceInvalidated\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newNonce\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PublicDeposit\",\"inputs\":[{\"name\":\"payer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"proofHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ScreenerUpdated\",\"inputs\":[{\"name\":\"newScreener\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Shield\",\"inputs\":[{\"name\":\"payer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"startLeafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transact\",\"inputs\":[{\"name\":\"inputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"ciphertexts\",\"type\":\"bytes[]\",\"indexed\":false,\"internalType\":\"bytes[]\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TransactAggregated\",\"inputs\":[{\"name\":\"proofHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"requestCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreeCreated\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressScreeningFailed\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChainIdMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CommitmentAlreadyPaid\",\"inputs\":[{\"name\":\"commitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"CreditDrawWindowClosed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DoubleSpend\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExceedsDepositLimit\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPerDeposit\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpiryInPast\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InexactTokenReceipt\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"received\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidAggregateBatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBatchSize\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCreditExpiry\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidFunding\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidIdentityCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidMerkleRoot\",\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidNonce\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidNullifierKeyHash\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOwnerIdentityCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPrivateTransferCiphertext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPrivateTransferContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPrivateTransferShape\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidProof\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReceivingAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShieldContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShieldOutputCount\",\"inputs\":[{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTree\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"LengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAdmin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotEmergencyAdmin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotGlobalEmergency\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NullifierAlreadySpent\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"NullifiersEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PrivateTransferExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ShieldExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SignatureExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StaleRoot\",\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"TransferFailed\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"TreeFull\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"WrongPermitToken\",\"inputs\":[{\"name\":\"provided\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expected\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ZeroNullifier\",\"inputs\":[]}]",
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

// FUNDSHIELDTYPEHASH is a free data retrieval call binding the contract method 0xbe9f7d0e.
//
// Solidity: function FUND_SHIELD_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) FUNDSHIELDTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "FUND_SHIELD_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// FUNDSHIELDTYPEHASH is a free data retrieval call binding the contract method 0xbe9f7d0e.
//
// Solidity: function FUND_SHIELD_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) FUNDSHIELDTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.FUNDSHIELDTYPEHASH(&_InvisiProtocol.CallOpts)
}

// FUNDSHIELDTYPEHASH is a free data retrieval call binding the contract method 0xbe9f7d0e.
//
// Solidity: function FUND_SHIELD_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) FUNDSHIELDTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.FUNDSHIELDTYPEHASH(&_InvisiProtocol.CallOpts)
}

// FUNDTRANSACTTYPEHASH is a free data retrieval call binding the contract method 0x8ebb23e8.
//
// Solidity: function FUND_TRANSACT_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) FUNDTRANSACTTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "FUND_TRANSACT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// FUNDTRANSACTTYPEHASH is a free data retrieval call binding the contract method 0x8ebb23e8.
//
// Solidity: function FUND_TRANSACT_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) FUNDTRANSACTTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.FUNDTRANSACTTYPEHASH(&_InvisiProtocol.CallOpts)
}

// FUNDTRANSACTTYPEHASH is a free data retrieval call binding the contract method 0x8ebb23e8.
//
// Solidity: function FUND_TRANSACT_TYPEHASH() view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) FUNDTRANSACTTYPEHASH() ([32]byte, error) {
	return _InvisiProtocol.Contract.FUNDTRANSACTTYPEHASH(&_InvisiProtocol.CallOpts)
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

// ActiveNoteTreeId is a free data retrieval call binding the contract method 0x236c8dd6.
//
// Solidity: function activeNoteTreeId() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) ActiveNoteTreeId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "activeNoteTreeId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ActiveNoteTreeId is a free data retrieval call binding the contract method 0x236c8dd6.
//
// Solidity: function activeNoteTreeId() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) ActiveNoteTreeId() (*big.Int, error) {
	return _InvisiProtocol.Contract.ActiveNoteTreeId(&_InvisiProtocol.CallOpts)
}

// ActiveNoteTreeId is a free data retrieval call binding the contract method 0x236c8dd6.
//
// Solidity: function activeNoteTreeId() view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) ActiveNoteTreeId() (*big.Int, error) {
	return _InvisiProtocol.Contract.ActiveNoteTreeId(&_InvisiProtocol.CallOpts)
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

// GetNoteTreeNextLeafIndex is a free data retrieval call binding the contract method 0xb7b81f59.
//
// Solidity: function getNoteTreeNextLeafIndex(uint256 treeId) view returns(uint32)
func (_InvisiProtocol *InvisiProtocolCaller) GetNoteTreeNextLeafIndex(opts *bind.CallOpts, treeId *big.Int) (uint32, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "getNoteTreeNextLeafIndex", treeId)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetNoteTreeNextLeafIndex is a free data retrieval call binding the contract method 0xb7b81f59.
//
// Solidity: function getNoteTreeNextLeafIndex(uint256 treeId) view returns(uint32)
func (_InvisiProtocol *InvisiProtocolSession) GetNoteTreeNextLeafIndex(treeId *big.Int) (uint32, error) {
	return _InvisiProtocol.Contract.GetNoteTreeNextLeafIndex(&_InvisiProtocol.CallOpts, treeId)
}

// GetNoteTreeNextLeafIndex is a free data retrieval call binding the contract method 0xb7b81f59.
//
// Solidity: function getNoteTreeNextLeafIndex(uint256 treeId) view returns(uint32)
func (_InvisiProtocol *InvisiProtocolCallerSession) GetNoteTreeNextLeafIndex(treeId *big.Int) (uint32, error) {
	return _InvisiProtocol.Contract.GetNoteTreeNextLeafIndex(&_InvisiProtocol.CallOpts, treeId)
}

// GetNoteTreeRoot is a free data retrieval call binding the contract method 0xd7793a38.
//
// Solidity: function getNoteTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCaller) GetNoteTreeRoot(opts *bind.CallOpts, treeId *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "getNoteTreeRoot", treeId)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetNoteTreeRoot is a free data retrieval call binding the contract method 0xd7793a38.
//
// Solidity: function getNoteTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolSession) GetNoteTreeRoot(treeId *big.Int) ([32]byte, error) {
	return _InvisiProtocol.Contract.GetNoteTreeRoot(&_InvisiProtocol.CallOpts, treeId)
}

// GetNoteTreeRoot is a free data retrieval call binding the contract method 0xd7793a38.
//
// Solidity: function getNoteTreeRoot(uint256 treeId) view returns(bytes32)
func (_InvisiProtocol *InvisiProtocolCallerSession) GetNoteTreeRoot(treeId *big.Int) ([32]byte, error) {
	return _InvisiProtocol.Contract.GetNoteTreeRoot(&_InvisiProtocol.CallOpts, treeId)
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
// Solidity: function isCommitmentPaid(bytes32 noteCommitment) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) IsCommitmentPaid(opts *bind.CallOpts, noteCommitment [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "isCommitmentPaid", noteCommitment)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCommitmentPaid is a free data retrieval call binding the contract method 0xe7c3ffe7.
//
// Solidity: function isCommitmentPaid(bytes32 noteCommitment) view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) IsCommitmentPaid(noteCommitment [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.IsCommitmentPaid(&_InvisiProtocol.CallOpts, noteCommitment)
}

// IsCommitmentPaid is a free data retrieval call binding the contract method 0xe7c3ffe7.
//
// Solidity: function isCommitmentPaid(bytes32 noteCommitment) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) IsCommitmentPaid(noteCommitment [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.IsCommitmentPaid(&_InvisiProtocol.CallOpts, noteCommitment)
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

// RootHistory is a free data retrieval call binding the contract method 0xc718dbda.
//
// Solidity: function rootHistory(uint256 , bytes32 ) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) RootHistory(opts *bind.CallOpts, arg0 *big.Int, arg1 [32]byte) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "rootHistory", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RootHistory is a free data retrieval call binding the contract method 0xc718dbda.
//
// Solidity: function rootHistory(uint256 , bytes32 ) view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) RootHistory(arg0 *big.Int, arg1 [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.RootHistory(&_InvisiProtocol.CallOpts, arg0, arg1)
}

// RootHistory is a free data retrieval call binding the contract method 0xc718dbda.
//
// Solidity: function rootHistory(uint256 , bytes32 ) view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) RootHistory(arg0 *big.Int, arg1 [32]byte) (bool, error) {
	return _InvisiProtocol.Contract.RootHistory(&_InvisiProtocol.CallOpts, arg0, arg1)
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

// CancelCredit is a paid mutator transaction binding the contract method 0xcbda6fb5.
//
// Solidity: function cancelCredit(bytes , ((uint256,bytes32),bytes32,address,uint256,uint256) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) CancelCredit(opts *bind.TransactOpts, arg0 []byte, arg1 IInvisiProtocolCancelCreditParams, arg2 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "cancelCredit", arg0, arg1, arg2)
}

// CancelCredit is a paid mutator transaction binding the contract method 0xcbda6fb5.
//
// Solidity: function cancelCredit(bytes , ((uint256,bytes32),bytes32,address,uint256,uint256) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolSession) CancelCredit(arg0 []byte, arg1 IInvisiProtocolCancelCreditParams, arg2 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CancelCredit(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2)
}

// CancelCredit is a paid mutator transaction binding the contract method 0xcbda6fb5.
//
// Solidity: function cancelCredit(bytes , ((uint256,bytes32),bytes32,address,uint256,uint256) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) CancelCredit(arg0 []byte, arg1 IInvisiProtocolCancelCreditParams, arg2 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CancelCredit(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0xd3d7c536.
//
// Solidity: function createCredit(bytes , ((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256) , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) CreateCredit(opts *bind.TransactOpts, arg0 []byte, arg1 IInvisiProtocolCreateCreditParams, arg2 [][]byte, arg3 IInvisiProtocolFundingParams, arg4 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "createCredit", arg0, arg1, arg2, arg3, arg4)
}

// CreateCredit is a paid mutator transaction binding the contract method 0xd3d7c536.
//
// Solidity: function createCredit(bytes , ((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256) , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolSession) CreateCredit(arg0 []byte, arg1 IInvisiProtocolCreateCreditParams, arg2 [][]byte, arg3 IInvisiProtocolFundingParams, arg4 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreateCredit(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// CreateCredit is a paid mutator transaction binding the contract method 0xd3d7c536.
//
// Solidity: function createCredit(bytes , ((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256) , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) CreateCredit(arg0 []byte, arg1 IInvisiProtocolCreateCreditParams, arg2 [][]byte, arg3 IInvisiProtocolFundingParams, arg4 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreateCredit(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// CreditDraw is a paid mutator transaction binding the contract method 0x402c4421.
//
// Solidity: function creditDraw(bytes , ((uint256,bytes32)[],bytes32[],bytes32[],uint256[],bytes32[],address,uint256,uint256) , bytes[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) CreditDraw(opts *bind.TransactOpts, arg0 []byte, arg1 IInvisiProtocolCreditDrawParams, arg2 [][]byte) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "creditDraw", arg0, arg1, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0x402c4421.
//
// Solidity: function creditDraw(bytes , ((uint256,bytes32)[],bytes32[],bytes32[],uint256[],bytes32[],address,uint256,uint256) , bytes[] ) returns()
func (_InvisiProtocol *InvisiProtocolSession) CreditDraw(arg0 []byte, arg1 IInvisiProtocolCreditDrawParams, arg2 [][]byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreditDraw(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0x402c4421.
//
// Solidity: function creditDraw(bytes , ((uint256,bytes32)[],bytes32[],bytes32[],uint256[],bytes32[],address,uint256,uint256) , bytes[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) CreditDraw(arg0 []byte, arg1 IInvisiProtocolCreditDrawParams, arg2 [][]byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreditDraw(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2)
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

// Shield is a paid mutator transaction binding the contract method 0x702e7d33.
//
// Solidity: function shield((bytes32[],uint256) , (bytes32,uint256,bytes32)[] , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Shield(opts *bind.TransactOpts, arg0 IInvisiProtocolShieldParams, arg1 []IInvisiProtocolShieldOutput, arg2 [][]byte, arg3 IInvisiProtocolFundingParams) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "shield", arg0, arg1, arg2, arg3)
}

// Shield is a paid mutator transaction binding the contract method 0x702e7d33.
//
// Solidity: function shield((bytes32[],uint256) , (bytes32,uint256,bytes32)[] , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiProtocol *InvisiProtocolSession) Shield(arg0 IInvisiProtocolShieldParams, arg1 []IInvisiProtocolShieldOutput, arg2 [][]byte, arg3 IInvisiProtocolFundingParams) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Shield(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2, arg3)
}

// Shield is a paid mutator transaction binding the contract method 0x702e7d33.
//
// Solidity: function shield((bytes32[],uint256) , (bytes32,uint256,bytes32)[] , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Shield(arg0 IInvisiProtocolShieldParams, arg1 []IInvisiProtocolShieldOutput, arg2 [][]byte, arg3 IInvisiProtocolFundingParams) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Shield(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2, arg3)
}

// TopUpCredit is a paid mutator transaction binding the contract method 0x36d16863.
//
// Solidity: function topUpCredit(bytes , (((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256),(uint256,bytes32),bytes32,uint256) , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) TopUpCredit(opts *bind.TransactOpts, arg0 []byte, arg1 IInvisiProtocolTopUpCreditParams, arg2 [][]byte, arg3 IInvisiProtocolFundingParams, arg4 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "topUpCredit", arg0, arg1, arg2, arg3, arg4)
}

// TopUpCredit is a paid mutator transaction binding the contract method 0x36d16863.
//
// Solidity: function topUpCredit(bytes , (((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256),(uint256,bytes32),bytes32,uint256) , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolSession) TopUpCredit(arg0 []byte, arg1 IInvisiProtocolTopUpCreditParams, arg2 [][]byte, arg3 IInvisiProtocolFundingParams, arg4 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.TopUpCredit(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// TopUpCredit is a paid mutator transaction binding the contract method 0x36d16863.
//
// Solidity: function topUpCredit(bytes , (((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256),(uint256,bytes32),bytes32,uint256) , bytes[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) , (uint256,bytes) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) TopUpCredit(arg0 []byte, arg1 IInvisiProtocolTopUpCreditParams, arg2 [][]byte, arg3 IInvisiProtocolFundingParams, arg4 IInvisiProtocolCreditAuthorization) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.TopUpCredit(&_InvisiProtocol.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// Transact is a paid mutator transaction binding the contract method 0x536d0c1e.
//
// Solidity: function transact(bytes , (((uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256),bytes[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Transact(opts *bind.TransactOpts, arg0 []byte, arg1 IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "transact", arg0, arg1)
}

// Transact is a paid mutator transaction binding the contract method 0x536d0c1e.
//
// Solidity: function transact(bytes , (((uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256),bytes[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) ) returns()
func (_InvisiProtocol *InvisiProtocolSession) Transact(arg0 []byte, arg1 IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact(&_InvisiProtocol.TransactOpts, arg0, arg1)
}

// Transact is a paid mutator transaction binding the contract method 0x536d0c1e.
//
// Solidity: function transact(bytes , (((uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256),bytes[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Transact(arg0 []byte, arg1 IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact(&_InvisiProtocol.TransactOpts, arg0, arg1)
}

// Transact0 is a paid mutator transaction binding the contract method 0x9bd36a3f.
//
// Solidity: function transact(bytes , (((uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256),bytes[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Transact0(opts *bind.TransactOpts, arg0 []byte, arg1 []IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "transact0", arg0, arg1)
}

// Transact0 is a paid mutator transaction binding the contract method 0x9bd36a3f.
//
// Solidity: function transact(bytes , (((uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256),bytes[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] ) returns()
func (_InvisiProtocol *InvisiProtocolSession) Transact0(arg0 []byte, arg1 []IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact0(&_InvisiProtocol.TransactOpts, arg0, arg1)
}

// Transact0 is a paid mutator transaction binding the contract method 0x9bd36a3f.
//
// Solidity: function transact(bytes , (((uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256),bytes[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Transact0(arg0 []byte, arg1 []IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact0(&_InvisiProtocol.TransactOpts, arg0, arg1)
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

// InvisiProtocolBatchOutputsIterator is returned from FilterBatchOutputs and is used to iterate over the raw logs and unpacked data for BatchOutputs events raised by the InvisiProtocol contract.
type InvisiProtocolBatchOutputsIterator struct {
	Event *InvisiProtocolBatchOutputs // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolBatchOutputsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolBatchOutputs)
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
		it.Event = new(InvisiProtocolBatchOutputs)
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
func (it *InvisiProtocolBatchOutputsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolBatchOutputsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolBatchOutputs represents a BatchOutputs event raised by the InvisiProtocol contract.
type InvisiProtocolBatchOutputs struct {
	TreeId            *big.Int
	StartLeafIndex    uint32
	FirstOutputOffset *big.Int
	Commitments       [][32]byte
	PackedCiphertexts []byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterBatchOutputs is a free log retrieval operation binding the contract event 0xb777242f5badb86f5f4c26fb210306f07680ae3c44342065283f32ae6a8e890a.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, bytes32[] commitments, bytes packedCiphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterBatchOutputs(opts *bind.FilterOpts, treeId []*big.Int) (*InvisiProtocolBatchOutputsIterator, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "BatchOutputs", treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolBatchOutputsIterator{contract: _InvisiProtocol.contract, event: "BatchOutputs", logs: logs, sub: sub}, nil
}

// WatchBatchOutputs is a free log subscription operation binding the contract event 0xb777242f5badb86f5f4c26fb210306f07680ae3c44342065283f32ae6a8e890a.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, bytes32[] commitments, bytes packedCiphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchBatchOutputs(opts *bind.WatchOpts, sink chan<- *InvisiProtocolBatchOutputs, treeId []*big.Int) (event.Subscription, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "BatchOutputs", treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolBatchOutputs)
				if err := _InvisiProtocol.contract.UnpackLog(event, "BatchOutputs", log); err != nil {
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

// ParseBatchOutputs is a log parse operation binding the contract event 0xb777242f5badb86f5f4c26fb210306f07680ae3c44342065283f32ae6a8e890a.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, bytes32[] commitments, bytes packedCiphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseBatchOutputs(log types.Log) (*InvisiProtocolBatchOutputs, error) {
	event := new(InvisiProtocolBatchOutputs)
	if err := _InvisiProtocol.contract.UnpackLog(event, "BatchOutputs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolBatchRequestIterator is returned from FilterBatchRequest and is used to iterate over the raw logs and unpacked data for BatchRequest events raised by the InvisiProtocol contract.
type InvisiProtocolBatchRequestIterator struct {
	Event *InvisiProtocolBatchRequest // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolBatchRequestIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolBatchRequest)
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
		it.Event = new(InvisiProtocolBatchRequest)
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
func (it *InvisiProtocolBatchRequestIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolBatchRequestIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolBatchRequest represents a BatchRequest event raised by the InvisiProtocol contract.
type InvisiProtocolBatchRequest struct {
	RequestIndex      *big.Int
	Capacity          *big.Int
	Sources           []IInvisiProtocolInputSource
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

// FilterBatchRequest is a free log retrieval operation binding the contract event 0x3bdeaeb068d1980b36c4faa4e183df2aa30b9d9b7acc367ef33e68dfd5fb9119.
//
// Solidity: event BatchRequest(uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterBatchRequest(opts *bind.FilterOpts, requestIndex []*big.Int) (*InvisiProtocolBatchRequestIterator, error) {

	var requestIndexRule []interface{}
	for _, requestIndexItem := range requestIndex {
		requestIndexRule = append(requestIndexRule, requestIndexItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "BatchRequest", requestIndexRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolBatchRequestIterator{contract: _InvisiProtocol.contract, event: "BatchRequest", logs: logs, sub: sub}, nil
}

// WatchBatchRequest is a free log subscription operation binding the contract event 0x3bdeaeb068d1980b36c4faa4e183df2aa30b9d9b7acc367ef33e68dfd5fb9119.
//
// Solidity: event BatchRequest(uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchBatchRequest(opts *bind.WatchOpts, sink chan<- *InvisiProtocolBatchRequest, requestIndex []*big.Int) (event.Subscription, error) {

	var requestIndexRule []interface{}
	for _, requestIndexItem := range requestIndex {
		requestIndexRule = append(requestIndexRule, requestIndexItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "BatchRequest", requestIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolBatchRequest)
				if err := _InvisiProtocol.contract.UnpackLog(event, "BatchRequest", log); err != nil {
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

// ParseBatchRequest is a log parse operation binding the contract event 0x3bdeaeb068d1980b36c4faa4e183df2aa30b9d9b7acc367ef33e68dfd5fb9119.
//
// Solidity: event BatchRequest(uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseBatchRequest(log types.Log) (*InvisiProtocolBatchRequest, error) {
	event := new(InvisiProtocolBatchRequest)
	if err := _InvisiProtocol.contract.UnpackLog(event, "BatchRequest", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCommitmentsInsertedIterator is returned from FilterCommitmentsInserted and is used to iterate over the raw logs and unpacked data for CommitmentsInserted events raised by the InvisiProtocol contract.
type InvisiProtocolCommitmentsInsertedIterator struct {
	Event *InvisiProtocolCommitmentsInserted // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCommitmentsInsertedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCommitmentsInserted)
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
		it.Event = new(InvisiProtocolCommitmentsInserted)
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
func (it *InvisiProtocolCommitmentsInsertedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCommitmentsInsertedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCommitmentsInserted represents a CommitmentsInserted event raised by the InvisiProtocol contract.
type InvisiProtocolCommitmentsInserted struct {
	TreeId         *big.Int
	StartLeafIndex uint32
	Commitments    [][32]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterCommitmentsInserted is a free log retrieval operation binding the contract event 0x184a7bd65e05e172f493712d9cf2e9cd5fab522a20803086fec8c1502742214a.
//
// Solidity: event CommitmentsInserted(uint256 indexed treeId, uint32 startLeafIndex, bytes32[] commitments)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCommitmentsInserted(opts *bind.FilterOpts, treeId []*big.Int) (*InvisiProtocolCommitmentsInsertedIterator, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CommitmentsInserted", treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCommitmentsInsertedIterator{contract: _InvisiProtocol.contract, event: "CommitmentsInserted", logs: logs, sub: sub}, nil
}

// WatchCommitmentsInserted is a free log subscription operation binding the contract event 0x184a7bd65e05e172f493712d9cf2e9cd5fab522a20803086fec8c1502742214a.
//
// Solidity: event CommitmentsInserted(uint256 indexed treeId, uint32 startLeafIndex, bytes32[] commitments)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCommitmentsInserted(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCommitmentsInserted, treeId []*big.Int) (event.Subscription, error) {

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CommitmentsInserted", treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCommitmentsInserted)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CommitmentsInserted", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCommitmentsInserted(log types.Log) (*InvisiProtocolCommitmentsInserted, error) {
	event := new(InvisiProtocolCommitmentsInserted)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CommitmentsInserted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditCancelledIterator is returned from FilterCreditCancelled and is used to iterate over the raw logs and unpacked data for CreditCancelled events raised by the InvisiProtocol contract.
type InvisiProtocolCreditCancelledIterator struct {
	Event *InvisiProtocolCreditCancelled // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditCancelled)
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
		it.Event = new(InvisiProtocolCreditCancelled)
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
func (it *InvisiProtocolCreditCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditCancelled represents a CreditCancelled event raised by the InvisiProtocol contract.
type InvisiProtocolCreditCancelled struct {
	Owner       common.Address
	InputTreeId *big.Int
	Nullifier   [32]byte
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterCreditCancelled is a free log retrieval operation binding the contract event 0xd3a01af206da81bc9f6b178a7e45bd5b3794b25fa7af09d3ead80789f983a482.
//
// Solidity: event CreditCancelled(address indexed owner, uint256 indexed inputTreeId, bytes32 indexed nullifier, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditCancelled(opts *bind.FilterOpts, owner []common.Address, inputTreeId []*big.Int, nullifier [][32]byte) (*InvisiProtocolCreditCancelledIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var inputTreeIdRule []interface{}
	for _, inputTreeIdItem := range inputTreeId {
		inputTreeIdRule = append(inputTreeIdRule, inputTreeIdItem)
	}
	var nullifierRule []interface{}
	for _, nullifierItem := range nullifier {
		nullifierRule = append(nullifierRule, nullifierItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditCancelled", ownerRule, inputTreeIdRule, nullifierRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditCancelledIterator{contract: _InvisiProtocol.contract, event: "CreditCancelled", logs: logs, sub: sub}, nil
}

// WatchCreditCancelled is a free log subscription operation binding the contract event 0xd3a01af206da81bc9f6b178a7e45bd5b3794b25fa7af09d3ead80789f983a482.
//
// Solidity: event CreditCancelled(address indexed owner, uint256 indexed inputTreeId, bytes32 indexed nullifier, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditCancelled(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditCancelled, owner []common.Address, inputTreeId []*big.Int, nullifier [][32]byte) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var inputTreeIdRule []interface{}
	for _, inputTreeIdItem := range inputTreeId {
		inputTreeIdRule = append(inputTreeIdRule, inputTreeIdItem)
	}
	var nullifierRule []interface{}
	for _, nullifierItem := range nullifier {
		nullifierRule = append(nullifierRule, nullifierItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditCancelled", ownerRule, inputTreeIdRule, nullifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditCancelled)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditCancelled", log); err != nil {
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

// ParseCreditCancelled is a log parse operation binding the contract event 0xd3a01af206da81bc9f6b178a7e45bd5b3794b25fa7af09d3ead80789f983a482.
//
// Solidity: event CreditCancelled(address indexed owner, uint256 indexed inputTreeId, bytes32 indexed nullifier, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditCancelled(log types.Log) (*InvisiProtocolCreditCancelled, error) {
	event := new(InvisiProtocolCreditCancelled)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditCreatedIterator is returned from FilterCreditCreated and is used to iterate over the raw logs and unpacked data for CreditCreated events raised by the InvisiProtocol contract.
type InvisiProtocolCreditCreatedIterator struct {
	Event *InvisiProtocolCreditCreated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditCreated)
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
		it.Event = new(InvisiProtocolCreditCreated)
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
func (it *InvisiProtocolCreditCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditCreated represents a CreditCreated event raised by the InvisiProtocol contract.
type InvisiProtocolCreditCreated struct {
	Owner            common.Address
	OutputTreeId     *big.Int
	Sources          []IInvisiProtocolInputSource
	Nullifiers       [][32]byte
	CreditCommitment [32]byte
	ChangeCommitment [32]byte
	LeafIndices      []uint32
	Ciphertexts      [][]byte
	Expiry           *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCreditCreated is a free log retrieval operation binding the contract event 0x0610845e6868b37caef6415707fb47f139760ae7ca75efda2fb134e6c5cfbb1b.
//
// Solidity: event CreditCreated(address indexed owner, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32 creditCommitment, bytes32 changeCommitment, uint32[] leafIndices, bytes[] ciphertexts, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditCreated(opts *bind.FilterOpts, owner []common.Address, outputTreeId []*big.Int) (*InvisiProtocolCreditCreatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditCreated", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditCreatedIterator{contract: _InvisiProtocol.contract, event: "CreditCreated", logs: logs, sub: sub}, nil
}

// WatchCreditCreated is a free log subscription operation binding the contract event 0x0610845e6868b37caef6415707fb47f139760ae7ca75efda2fb134e6c5cfbb1b.
//
// Solidity: event CreditCreated(address indexed owner, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32 creditCommitment, bytes32 changeCommitment, uint32[] leafIndices, bytes[] ciphertexts, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditCreated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditCreated, owner []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditCreated", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditCreated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditCreated", log); err != nil {
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

// ParseCreditCreated is a log parse operation binding the contract event 0x0610845e6868b37caef6415707fb47f139760ae7ca75efda2fb134e6c5cfbb1b.
//
// Solidity: event CreditCreated(address indexed owner, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32 creditCommitment, bytes32 changeCommitment, uint32[] leafIndices, bytes[] ciphertexts, uint256 expiry)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditCreated(log types.Log) (*InvisiProtocolCreditCreated, error) {
	event := new(InvisiProtocolCreditCreated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditCreated", log); err != nil {
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
	Spender              common.Address
	OutputTreeId         *big.Int
	Params               IInvisiProtocolCreditDrawParams
	RemainderLeafIndices []uint32
	Ciphertexts          [][]byte
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterCreditDraw is a free log retrieval operation binding the contract event 0x0b01e5ed3eb8bfbba123707c187741678eb97e5133240f33e85cd46bdf034b11.
//
// Solidity: event CreditDraw(address indexed spender, uint256 indexed outputTreeId, ((uint256,bytes32)[],bytes32[],bytes32[],uint256[],bytes32[],address,uint256,uint256) params, uint32[] remainderLeafIndices, bytes[] ciphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditDraw(opts *bind.FilterOpts, spender []common.Address, outputTreeId []*big.Int) (*InvisiProtocolCreditDrawIterator, error) {

	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditDraw", spenderRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditDrawIterator{contract: _InvisiProtocol.contract, event: "CreditDraw", logs: logs, sub: sub}, nil
}

// WatchCreditDraw is a free log subscription operation binding the contract event 0x0b01e5ed3eb8bfbba123707c187741678eb97e5133240f33e85cd46bdf034b11.
//
// Solidity: event CreditDraw(address indexed spender, uint256 indexed outputTreeId, ((uint256,bytes32)[],bytes32[],bytes32[],uint256[],bytes32[],address,uint256,uint256) params, uint32[] remainderLeafIndices, bytes[] ciphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditDraw(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditDraw, spender []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditDraw", spenderRule, outputTreeIdRule)
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

// ParseCreditDraw is a log parse operation binding the contract event 0x0b01e5ed3eb8bfbba123707c187741678eb97e5133240f33e85cd46bdf034b11.
//
// Solidity: event CreditDraw(address indexed spender, uint256 indexed outputTreeId, ((uint256,bytes32)[],bytes32[],bytes32[],uint256[],bytes32[],address,uint256,uint256) params, uint32[] remainderLeafIndices, bytes[] ciphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditDraw(log types.Log) (*InvisiProtocolCreditDraw, error) {
	event := new(InvisiProtocolCreditDraw)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditDraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditToppedUpIterator is returned from FilterCreditToppedUp and is used to iterate over the raw logs and unpacked data for CreditToppedUp events raised by the InvisiProtocol contract.
type InvisiProtocolCreditToppedUpIterator struct {
	Event *InvisiProtocolCreditToppedUp // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditToppedUpIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditToppedUp)
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
		it.Event = new(InvisiProtocolCreditToppedUp)
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
func (it *InvisiProtocolCreditToppedUpIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditToppedUpIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditToppedUp represents a CreditToppedUp event raised by the InvisiProtocol contract.
type InvisiProtocolCreditToppedUp struct {
	Owner        common.Address
	OutputTreeId *big.Int
	Params       IInvisiProtocolTopUpCreditParams
	LeafIndices  []uint32
	Ciphertexts  [][]byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditToppedUp is a free log retrieval operation binding the contract event 0x08b435688fb827d10a681a405ce7b441fbe167154f75767224351e07715637c9.
//
// Solidity: event CreditToppedUp(address indexed owner, uint256 indexed outputTreeId, (((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256),(uint256,bytes32),bytes32,uint256) params, uint32[] leafIndices, bytes[] ciphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditToppedUp(opts *bind.FilterOpts, owner []common.Address, outputTreeId []*big.Int) (*InvisiProtocolCreditToppedUpIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditToppedUp", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditToppedUpIterator{contract: _InvisiProtocol.contract, event: "CreditToppedUp", logs: logs, sub: sub}, nil
}

// WatchCreditToppedUp is a free log subscription operation binding the contract event 0x08b435688fb827d10a681a405ce7b441fbe167154f75767224351e07715637c9.
//
// Solidity: event CreditToppedUp(address indexed owner, uint256 indexed outputTreeId, (((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256),(uint256,bytes32),bytes32,uint256) params, uint32[] leafIndices, bytes[] ciphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditToppedUp(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditToppedUp, owner []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditToppedUp", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditToppedUp)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditToppedUp", log); err != nil {
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

// ParseCreditToppedUp is a log parse operation binding the contract event 0x08b435688fb827d10a681a405ce7b441fbe167154f75767224351e07715637c9.
//
// Solidity: event CreditToppedUp(address indexed owner, uint256 indexed outputTreeId, (((uint256,bytes32)[],bytes32[],bytes32,bytes32,address,uint256,uint256),(uint256,bytes32),bytes32,uint256) params, uint32[] leafIndices, bytes[] ciphertexts)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditToppedUp(log types.Log) (*InvisiProtocolCreditToppedUp, error) {
	event := new(InvisiProtocolCreditToppedUp)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditToppedUp", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCrossTreeTransactIterator is returned from FilterCrossTreeTransact and is used to iterate over the raw logs and unpacked data for CrossTreeTransact events raised by the InvisiProtocol contract.
type InvisiProtocolCrossTreeTransactIterator struct {
	Event *InvisiProtocolCrossTreeTransact // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCrossTreeTransactIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCrossTreeTransact)
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
		it.Event = new(InvisiProtocolCrossTreeTransact)
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
func (it *InvisiProtocolCrossTreeTransactIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCrossTreeTransactIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCrossTreeTransact represents a CrossTreeTransact event raised by the InvisiProtocol contract.
type InvisiProtocolCrossTreeTransact struct {
	OutputTreeId      *big.Int
	Sources           []IInvisiProtocolInputSource
	Nullifiers        [][32]byte
	OutputCommitments [][32]byte
	LeafIndices       []uint32
	Ciphertexts       [][]byte
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCrossTreeTransact is a free log retrieval operation binding the contract event 0x3a397292385a3a48664272813b6d002763f919f88527d7c30ba40bfb8b3b3172.
//
// Solidity: event CrossTreeTransact(uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, bytes[] ciphertexts, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCrossTreeTransact(opts *bind.FilterOpts, outputTreeId []*big.Int) (*InvisiProtocolCrossTreeTransactIterator, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CrossTreeTransact", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCrossTreeTransactIterator{contract: _InvisiProtocol.contract, event: "CrossTreeTransact", logs: logs, sub: sub}, nil
}

// WatchCrossTreeTransact is a free log subscription operation binding the contract event 0x3a397292385a3a48664272813b6d002763f919f88527d7c30ba40bfb8b3b3172.
//
// Solidity: event CrossTreeTransact(uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, bytes[] ciphertexts, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCrossTreeTransact(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCrossTreeTransact, outputTreeId []*big.Int) (event.Subscription, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CrossTreeTransact", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCrossTreeTransact)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CrossTreeTransact", log); err != nil {
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

// ParseCrossTreeTransact is a log parse operation binding the contract event 0x3a397292385a3a48664272813b6d002763f919f88527d7c30ba40bfb8b3b3172.
//
// Solidity: event CrossTreeTransact(uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, bytes[] ciphertexts, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCrossTreeTransact(log types.Log) (*InvisiProtocolCrossTreeTransact, error) {
	event := new(InvisiProtocolCrossTreeTransact)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CrossTreeTransact", log); err != nil {
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

// InvisiProtocolPublicDepositIterator is returned from FilterPublicDeposit and is used to iterate over the raw logs and unpacked data for PublicDeposit events raised by the InvisiProtocol contract.
type InvisiProtocolPublicDepositIterator struct {
	Event *InvisiProtocolPublicDeposit // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolPublicDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolPublicDeposit)
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
		it.Event = new(InvisiProtocolPublicDeposit)
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
func (it *InvisiProtocolPublicDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolPublicDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolPublicDeposit represents a PublicDeposit event raised by the InvisiProtocol contract.
type InvisiProtocolPublicDeposit struct {
	Payer     common.Address
	Amount    *big.Int
	ProofHash [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPublicDeposit is a free log retrieval operation binding the contract event 0xacec8d7ea4e2c39bfdce9213c449892c2cf4d27d5734eb154dbbcadda139542a.
//
// Solidity: event PublicDeposit(address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterPublicDeposit(opts *bind.FilterOpts, payer []common.Address) (*InvisiProtocolPublicDepositIterator, error) {

	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "PublicDeposit", payerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolPublicDepositIterator{contract: _InvisiProtocol.contract, event: "PublicDeposit", logs: logs, sub: sub}, nil
}

// WatchPublicDeposit is a free log subscription operation binding the contract event 0xacec8d7ea4e2c39bfdce9213c449892c2cf4d27d5734eb154dbbcadda139542a.
//
// Solidity: event PublicDeposit(address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchPublicDeposit(opts *bind.WatchOpts, sink chan<- *InvisiProtocolPublicDeposit, payer []common.Address) (event.Subscription, error) {

	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "PublicDeposit", payerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolPublicDeposit)
				if err := _InvisiProtocol.contract.UnpackLog(event, "PublicDeposit", log); err != nil {
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

// ParsePublicDeposit is a log parse operation binding the contract event 0xacec8d7ea4e2c39bfdce9213c449892c2cf4d27d5734eb154dbbcadda139542a.
//
// Solidity: event PublicDeposit(address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiProtocol *InvisiProtocolFilterer) ParsePublicDeposit(log types.Log) (*InvisiProtocolPublicDeposit, error) {
	event := new(InvisiProtocolPublicDeposit)
	if err := _InvisiProtocol.contract.UnpackLog(event, "PublicDeposit", log); err != nil {
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

// InvisiProtocolShieldIterator is returned from FilterShield and is used to iterate over the raw logs and unpacked data for Shield events raised by the InvisiProtocol contract.
type InvisiProtocolShieldIterator struct {
	Event *InvisiProtocolShield // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolShieldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolShield)
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
		it.Event = new(InvisiProtocolShield)
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
func (it *InvisiProtocolShieldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolShieldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolShield represents a Shield event raised by the InvisiProtocol contract.
type InvisiProtocolShield struct {
	Payer          common.Address
	Amount         *big.Int
	TreeId         *big.Int
	StartLeafIndex uint32
	Ciphertexts    [][]byte
	Deadline       *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterShield is a free log retrieval operation binding the contract event 0x97a98cd508547a5c343e8521eaf926f177a1f02816379623eb074fd5e43d6db6.
//
// Solidity: event Shield(address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, bytes[] ciphertexts, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterShield(opts *bind.FilterOpts, payer []common.Address, treeId []*big.Int) (*InvisiProtocolShieldIterator, error) {

	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "Shield", payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolShieldIterator{contract: _InvisiProtocol.contract, event: "Shield", logs: logs, sub: sub}, nil
}

// WatchShield is a free log subscription operation binding the contract event 0x97a98cd508547a5c343e8521eaf926f177a1f02816379623eb074fd5e43d6db6.
//
// Solidity: event Shield(address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, bytes[] ciphertexts, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchShield(opts *bind.WatchOpts, sink chan<- *InvisiProtocolShield, payer []common.Address, treeId []*big.Int) (event.Subscription, error) {

	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	var treeIdRule []interface{}
	for _, treeIdItem := range treeId {
		treeIdRule = append(treeIdRule, treeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "Shield", payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolShield)
				if err := _InvisiProtocol.contract.UnpackLog(event, "Shield", log); err != nil {
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

// ParseShield is a log parse operation binding the contract event 0x97a98cd508547a5c343e8521eaf926f177a1f02816379623eb074fd5e43d6db6.
//
// Solidity: event Shield(address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, bytes[] ciphertexts, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseShield(log types.Log) (*InvisiProtocolShield, error) {
	event := new(InvisiProtocolShield)
	if err := _InvisiProtocol.contract.UnpackLog(event, "Shield", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolTransactIterator is returned from FilterTransact and is used to iterate over the raw logs and unpacked data for Transact events raised by the InvisiProtocol contract.
type InvisiProtocolTransactIterator struct {
	Event *InvisiProtocolTransact // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolTransactIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolTransact)
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
		it.Event = new(InvisiProtocolTransact)
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
func (it *InvisiProtocolTransactIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolTransactIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolTransact represents a Transact event raised by the InvisiProtocol contract.
type InvisiProtocolTransact struct {
	InputTreeId       *big.Int
	OutputTreeId      *big.Int
	Nullifiers        [][32]byte
	OutputCommitments [][32]byte
	LeafIndices       []uint32
	Ciphertexts       [][]byte
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterTransact is a free log retrieval operation binding the contract event 0x9cedaa0dd097dadc64592eef3d97f25bd2e79597f9760511d64fae3c28f51b84.
//
// Solidity: event Transact(uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, bytes[] ciphertexts, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterTransact(opts *bind.FilterOpts, inputTreeId []*big.Int, outputTreeId []*big.Int) (*InvisiProtocolTransactIterator, error) {

	var inputTreeIdRule []interface{}
	for _, inputTreeIdItem := range inputTreeId {
		inputTreeIdRule = append(inputTreeIdRule, inputTreeIdItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "Transact", inputTreeIdRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolTransactIterator{contract: _InvisiProtocol.contract, event: "Transact", logs: logs, sub: sub}, nil
}

// WatchTransact is a free log subscription operation binding the contract event 0x9cedaa0dd097dadc64592eef3d97f25bd2e79597f9760511d64fae3c28f51b84.
//
// Solidity: event Transact(uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, bytes[] ciphertexts, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchTransact(opts *bind.WatchOpts, sink chan<- *InvisiProtocolTransact, inputTreeId []*big.Int, outputTreeId []*big.Int) (event.Subscription, error) {

	var inputTreeIdRule []interface{}
	for _, inputTreeIdItem := range inputTreeId {
		inputTreeIdRule = append(inputTreeIdRule, inputTreeIdItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "Transact", inputTreeIdRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolTransact)
				if err := _InvisiProtocol.contract.UnpackLog(event, "Transact", log); err != nil {
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

// ParseTransact is a log parse operation binding the contract event 0x9cedaa0dd097dadc64592eef3d97f25bd2e79597f9760511d64fae3c28f51b84.
//
// Solidity: event Transact(uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, bytes[] ciphertexts, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseTransact(log types.Log) (*InvisiProtocolTransact, error) {
	event := new(InvisiProtocolTransact)
	if err := _InvisiProtocol.contract.UnpackLog(event, "Transact", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolTransactAggregatedIterator is returned from FilterTransactAggregated and is used to iterate over the raw logs and unpacked data for TransactAggregated events raised by the InvisiProtocol contract.
type InvisiProtocolTransactAggregatedIterator struct {
	Event *InvisiProtocolTransactAggregated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolTransactAggregatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolTransactAggregated)
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
		it.Event = new(InvisiProtocolTransactAggregated)
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
func (it *InvisiProtocolTransactAggregatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolTransactAggregatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolTransactAggregated represents a TransactAggregated event raised by the InvisiProtocol contract.
type InvisiProtocolTransactAggregated struct {
	ProofHash    [32]byte
	RequestCount *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTransactAggregated is a free log retrieval operation binding the contract event 0xb550ba00978bffde806b7356916d552dea8794ebdc978bc7d205be9d2ce1a8df.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, uint256 requestCount)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterTransactAggregated(opts *bind.FilterOpts, proofHash [][32]byte) (*InvisiProtocolTransactAggregatedIterator, error) {

	var proofHashRule []interface{}
	for _, proofHashItem := range proofHash {
		proofHashRule = append(proofHashRule, proofHashItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "TransactAggregated", proofHashRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolTransactAggregatedIterator{contract: _InvisiProtocol.contract, event: "TransactAggregated", logs: logs, sub: sub}, nil
}

// WatchTransactAggregated is a free log subscription operation binding the contract event 0xb550ba00978bffde806b7356916d552dea8794ebdc978bc7d205be9d2ce1a8df.
//
// Solidity: event TransactAggregated(bytes32 indexed proofHash, uint256 requestCount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchTransactAggregated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolTransactAggregated, proofHash [][32]byte) (event.Subscription, error) {

	var proofHashRule []interface{}
	for _, proofHashItem := range proofHash {
		proofHashRule = append(proofHashRule, proofHashItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "TransactAggregated", proofHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolTransactAggregated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "TransactAggregated", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseTransactAggregated(log types.Log) (*InvisiProtocolTransactAggregated, error) {
	event := new(InvisiProtocolTransactAggregated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "TransactAggregated", log); err != nil {
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
