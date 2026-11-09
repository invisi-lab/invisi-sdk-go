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

// IInvisiProtocolCreateCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreateCreditParams struct {
	Token            common.Address
	Owner            common.Address
	RevocationShard  uint64
	RevocationSlot   uint16
	CreditCommitment [32]byte
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}

// IInvisiProtocolCreditDrawParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreditDrawParams struct {
	Token            common.Address
	Sources          []IInvisiProtocolInputSource
	Payers           []IInvisiProtocolCreditPayer
	GrantRefs        []IInvisiProtocolCreditGrantRef
	Nullifiers       [][32]byte
	StateCommitments [][32]byte
	RecoveryAmounts  [][32]byte
	NoteCommitment   [32]byte
	TotalAmount      *big.Int
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}

// IInvisiProtocolCreditGrantRef is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreditGrantRef struct {
	PayerIndex *big.Int
	Shard      uint64
	Slot       uint16
}

// IInvisiProtocolCreditPayer is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCreditPayer struct {
	Payer  common.Address
	Amount *big.Int
}

// IInvisiProtocolCrossTreeTransactParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolCrossTreeTransactParams struct {
	Token             common.Address
	Sources           []IInvisiProtocolInputSource
	Nullifiers        [][32]byte
	OutputCommitments [][32]byte
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	FeeRecipient      common.Address
	FeeAmount         *big.Int
}

// IInvisiProtocolDelivery is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolDelivery struct {
	Ciphertext []byte
	Clue       []byte
}

// IInvisiProtocolFundingParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolFundingParams struct {
	Payer     common.Address
	Amount    *big.Int
	SigNonce  *big.Int
	Deadline  *big.Int
	Signature []byte
	Permit    IMetaTxTokenPermit
}

// IInvisiProtocolInputSource is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolInputSource struct {
	TreeId     *big.Int
	MerkleRoot [32]byte
}

// IInvisiProtocolShieldOutput is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolShieldOutput struct {
	Identity [32]byte
	Amount   *big.Int
}

// IInvisiProtocolShieldParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolShieldParams struct {
	Token             common.Address
	OutputCommitments [][32]byte
	Deadline          *big.Int
	FeeRecipient      common.Address
	FeeAmount         *big.Int
}

// IInvisiProtocolTransactRequest is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolTransactRequest struct {
	Params     IInvisiProtocolCrossTreeTransactParams
	Deliveries []IInvisiProtocolDelivery
	Funding    IInvisiProtocolFundingParams
}

// IInvisiProtocolUpdateCreditParams is an auto generated low-level Go binding around an user-defined struct.
type IInvisiProtocolUpdateCreditParams struct {
	Token            common.Address
	Owner            common.Address
	Source           IInvisiProtocolInputSource
	Nullifier        [32]byte
	CreditCommitment [32]byte
	RevocationShard  uint64
	RevocationSlot   uint16
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}

// IMetaTxTokenPermit is an auto generated low-level Go binding around an user-defined struct.
type IMetaTxTokenPermit struct {
	Deadline *big.Int
	V        uint8
	R        [32]byte
	S        [32]byte
}

// InstanceParams is an auto generated low-level Go binding around an user-defined struct.
type InstanceParams struct {
	Factory         common.Address
	InitialOwner    common.Address
	GlobalEmergency common.Address
	EmergencyAdmin  common.Address
	VerifierRouter  common.Address
}

// InvisiProtocolMetaData contains all meta data concerning the InvisiProtocol contract.
var InvisiProtocolMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structInstanceParams\",\"components\":[{\"name\":\"factory\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"globalEmergency\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"emergencyAdmin\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"verifierRouter\",\"type\":\"address\",\"internalType\":\"contractIVerifierRouter\"}]},{\"name\":\"_screener\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_addressRegistry\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"CHAIN_ID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"FUND_SHIELD_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"FUND_TRANSACT_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"activeNoteTreeId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"addressRegistry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAddressRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"broadcasterFeeBalance\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"circuitPaused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"createCredit\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"p\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreateCreditParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"revocationShard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revocationSlot\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditAllocation\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"shard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"slot\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"creditDraw\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"p\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CreditDrawParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"payers\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.CreditPayer[]\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"grantRefs\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.CreditGrantRef[]\",\"components\":[{\"name\":\"payerIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"shard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"slot\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"stateCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"recoveryAmounts\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"noteCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditRevocationBitmap\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"shard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"word\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"emergencyAdmin\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"factory\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNoteTreeNextLeafIndex\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNoteTreeRoot\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTreeCount\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"globalEmergency\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"invalidateNonce\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isCommitmentPaid\",\"inputs\":[{\"name\":\"noteCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isNullifierSpent\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isTreeFull\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nonces\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"pauseCircuit\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeCredit\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"shard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"slot\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"rootHistory\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"screener\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setScreener\",\"inputs\":[{\"name\":\"newScreener\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"shield\",\"inputs\":[{\"name\":\"p\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.ShieldParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.ShieldOutput[]\",\"components\":[{\"name\":\"identity\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transact\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"r\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.TransactRequest\",\"components\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CrossTreeTransactParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"funding\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transact\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"requests\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.TransactRequest[]\",\"components\":[{\"name\":\"params\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.CrossTreeTransactParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"funding\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.FundingParams\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"sigNonce\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"permit\",\"type\":\"tuple\",\"internalType\":\"structIMetaTx.TokenPermit\",\"components\":[{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unpauseCircuit\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateCredit\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"p\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.UpdateCreditParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"source\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.InputSource\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"revocationShard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revocationSlot\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verifierRouter\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifierRouter\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdrawBroadcasterFees\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"BatchOutputs\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"startLeafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"firstOutputOffset\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"tokens\",\"type\":\"address[]\",\"indexed\":false,\"internalType\":\"address[]\"},{\"name\":\"commitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"packedDeliveries\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"BatchRequest\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"requestIndex\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"capacity\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"outputCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payer\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"depositAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"statementHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"BroadcasterFee\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"BroadcasterFeeWithdrawn\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CircuitPauseUpdated\",\"inputs\":[{\"name\":\"paused\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CommitmentsInserted\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"startLeafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"commitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditCreated\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"revocationShard\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"},{\"name\":\"revocationSlot\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"leafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditDraw\",\"inputs\":[{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"params\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.CreditDrawParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"payers\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.CreditPayer[]\",\"components\":[{\"name\":\"payer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"grantRefs\",\"type\":\"tuple[]\",\"internalType\":\"structIInvisiProtocol.CreditGrantRef[]\",\"components\":[{\"name\":\"payerIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"shard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"slot\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"stateCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"recoveryAmounts\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"noteCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditRevoked\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"shard\",\"type\":\"uint64\",\"indexed\":true,\"internalType\":\"uint64\"},{\"name\":\"slot\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CreditUpdated\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"params\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.UpdateCreditParams\",\"components\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"source\",\"type\":\"tuple\",\"internalType\":\"structIInvisiProtocol.InputSource\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"creditCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"revocationShard\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revocationSlot\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"leafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CrossTreeTransact\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"sources\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.InputSource[]\",\"components\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NonceInvalidated\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newNonce\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PublicDeposit\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"payer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"proofHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ScreenerUpdated\",\"inputs\":[{\"name\":\"newScreener\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Shield\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"payer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"startLeafIndex\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transact\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"inputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"outputTreeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"nullifiers\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"outputCommitments\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"leafIndices\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"},{\"name\":\"deliveries\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIInvisiProtocol.Delivery[]\",\"components\":[{\"name\":\"ciphertext\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"clue\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"withdrawalAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"withdrawalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TransactAggregated\",\"inputs\":[{\"name\":\"proofHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"requestCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreeCreated\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AddressInsufficientBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AddressScreeningFailed\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChainIdMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CommitmentAlreadyPaid\",\"inputs\":[{\"name\":\"commitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"CreditAlreadyRevoked\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DoubleSpend\",\"inputs\":[{\"name\":\"nullifier\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedInnerCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InexactTokenReceipt\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"received\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidAddressRegistry\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAggregateBatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBatchSize\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCommitment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCreditAllocation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDelivery\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidFunding\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidMerkleRoot\",\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidNonce\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPrivateTransferContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPrivateTransferShape\",\"inputs\":[{\"name\":\"inputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidProof\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReceivingAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShieldContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShieldOutputCount\",\"inputs\":[{\"name\":\"outputs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidToken\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidTree\",\"inputs\":[{\"name\":\"treeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"LengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoBroadcasterFees\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotEmergencyAdmin\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotGlobalEmergency\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PrivateTransferExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ShieldExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SignatureExpired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StaleRoot\",\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"ZeroNullifier\",\"inputs\":[]}]",
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

// AddressRegistry is a free data retrieval call binding the contract method 0xf3ad65f4.
//
// Solidity: function addressRegistry() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) AddressRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "addressRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddressRegistry is a free data retrieval call binding the contract method 0xf3ad65f4.
//
// Solidity: function addressRegistry() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) AddressRegistry() (common.Address, error) {
	return _InvisiProtocol.Contract.AddressRegistry(&_InvisiProtocol.CallOpts)
}

// AddressRegistry is a free data retrieval call binding the contract method 0xf3ad65f4.
//
// Solidity: function addressRegistry() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) AddressRegistry() (common.Address, error) {
	return _InvisiProtocol.Contract.AddressRegistry(&_InvisiProtocol.CallOpts)
}

// BroadcasterFeeBalance is a free data retrieval call binding the contract method 0x0bc8d1bd.
//
// Solidity: function broadcasterFeeBalance(address token, address recipient) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) BroadcasterFeeBalance(opts *bind.CallOpts, token common.Address, recipient common.Address) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "broadcasterFeeBalance", token, recipient)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BroadcasterFeeBalance is a free data retrieval call binding the contract method 0x0bc8d1bd.
//
// Solidity: function broadcasterFeeBalance(address token, address recipient) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) BroadcasterFeeBalance(token common.Address, recipient common.Address) (*big.Int, error) {
	return _InvisiProtocol.Contract.BroadcasterFeeBalance(&_InvisiProtocol.CallOpts, token, recipient)
}

// BroadcasterFeeBalance is a free data retrieval call binding the contract method 0x0bc8d1bd.
//
// Solidity: function broadcasterFeeBalance(address token, address recipient) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) BroadcasterFeeBalance(token common.Address, recipient common.Address) (*big.Int, error) {
	return _InvisiProtocol.Contract.BroadcasterFeeBalance(&_InvisiProtocol.CallOpts, token, recipient)
}

// CircuitPaused is a free data retrieval call binding the contract method 0x5c12671f.
//
// Solidity: function circuitPaused() view returns(bool)
func (_InvisiProtocol *InvisiProtocolCaller) CircuitPaused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "circuitPaused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CircuitPaused is a free data retrieval call binding the contract method 0x5c12671f.
//
// Solidity: function circuitPaused() view returns(bool)
func (_InvisiProtocol *InvisiProtocolSession) CircuitPaused() (bool, error) {
	return _InvisiProtocol.Contract.CircuitPaused(&_InvisiProtocol.CallOpts)
}

// CircuitPaused is a free data retrieval call binding the contract method 0x5c12671f.
//
// Solidity: function circuitPaused() view returns(bool)
func (_InvisiProtocol *InvisiProtocolCallerSession) CircuitPaused() (bool, error) {
	return _InvisiProtocol.Contract.CircuitPaused(&_InvisiProtocol.CallOpts)
}

// CreditAllocation is a free data retrieval call binding the contract method 0x70a7588f.
//
// Solidity: function creditAllocation(address owner, address token) view returns(uint64 shard, uint16 slot)
func (_InvisiProtocol *InvisiProtocolCaller) CreditAllocation(opts *bind.CallOpts, owner common.Address, token common.Address) (struct {
	Shard uint64
	Slot  uint16
}, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "creditAllocation", owner, token)

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
func (_InvisiProtocol *InvisiProtocolSession) CreditAllocation(owner common.Address, token common.Address) (struct {
	Shard uint64
	Slot  uint16
}, error) {
	return _InvisiProtocol.Contract.CreditAllocation(&_InvisiProtocol.CallOpts, owner, token)
}

// CreditAllocation is a free data retrieval call binding the contract method 0x70a7588f.
//
// Solidity: function creditAllocation(address owner, address token) view returns(uint64 shard, uint16 slot)
func (_InvisiProtocol *InvisiProtocolCallerSession) CreditAllocation(owner common.Address, token common.Address) (struct {
	Shard uint64
	Slot  uint16
}, error) {
	return _InvisiProtocol.Contract.CreditAllocation(&_InvisiProtocol.CallOpts, owner, token)
}

// CreditRevocationBitmap is a free data retrieval call binding the contract method 0xc91ae210.
//
// Solidity: function creditRevocationBitmap(address owner, address token, uint64 shard, uint8 word) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCaller) CreditRevocationBitmap(opts *bind.CallOpts, owner common.Address, token common.Address, shard uint64, word uint8) (*big.Int, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "creditRevocationBitmap", owner, token, shard, word)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CreditRevocationBitmap is a free data retrieval call binding the contract method 0xc91ae210.
//
// Solidity: function creditRevocationBitmap(address owner, address token, uint64 shard, uint8 word) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) CreditRevocationBitmap(owner common.Address, token common.Address, shard uint64, word uint8) (*big.Int, error) {
	return _InvisiProtocol.Contract.CreditRevocationBitmap(&_InvisiProtocol.CallOpts, owner, token, shard, word)
}

// CreditRevocationBitmap is a free data retrieval call binding the contract method 0xc91ae210.
//
// Solidity: function creditRevocationBitmap(address owner, address token, uint64 shard, uint8 word) view returns(uint256)
func (_InvisiProtocol *InvisiProtocolCallerSession) CreditRevocationBitmap(owner common.Address, token common.Address, shard uint64, word uint8) (*big.Int, error) {
	return _InvisiProtocol.Contract.CreditRevocationBitmap(&_InvisiProtocol.CallOpts, owner, token, shard, word)
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

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) Owner() (common.Address, error) {
	return _InvisiProtocol.Contract.Owner(&_InvisiProtocol.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) Owner() (common.Address, error) {
	return _InvisiProtocol.Contract.Owner(&_InvisiProtocol.CallOpts)
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

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiProtocol *InvisiProtocolCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _InvisiProtocol.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiProtocol *InvisiProtocolSession) PendingOwner() (common.Address, error) {
	return _InvisiProtocol.Contract.PendingOwner(&_InvisiProtocol.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_InvisiProtocol *InvisiProtocolCallerSession) PendingOwner() (common.Address, error) {
	return _InvisiProtocol.Contract.PendingOwner(&_InvisiProtocol.CallOpts)
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

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiProtocol *InvisiProtocolSession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.AcceptOwnership(&_InvisiProtocol.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.AcceptOwnership(&_InvisiProtocol.TransactOpts)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x5550f27a.
//
// Solidity: function createCredit(bytes , (address,address,uint64,uint16,bytes32,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) CreateCredit(opts *bind.TransactOpts, arg0 []byte, p IInvisiProtocolCreateCreditParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "createCredit", arg0, p, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x5550f27a.
//
// Solidity: function createCredit(bytes , (address,address,uint64,uint16,bytes32,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolSession) CreateCredit(arg0 []byte, p IInvisiProtocolCreateCreditParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreateCredit(&_InvisiProtocol.TransactOpts, arg0, p, arg2)
}

// CreateCredit is a paid mutator transaction binding the contract method 0x5550f27a.
//
// Solidity: function createCredit(bytes , (address,address,uint64,uint16,bytes32,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) CreateCredit(arg0 []byte, p IInvisiProtocolCreateCreditParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreateCredit(&_InvisiProtocol.TransactOpts, arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa288961c.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) CreditDraw(opts *bind.TransactOpts, arg0 []byte, p IInvisiProtocolCreditDrawParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "creditDraw", arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa288961c.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolSession) CreditDraw(arg0 []byte, p IInvisiProtocolCreditDrawParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreditDraw(&_InvisiProtocol.TransactOpts, arg0, p, arg2)
}

// CreditDraw is a paid mutator transaction binding the contract method 0xa288961c.
//
// Solidity: function creditDraw(bytes , (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) CreditDraw(arg0 []byte, p IInvisiProtocolCreditDrawParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.CreditDraw(&_InvisiProtocol.TransactOpts, arg0, p, arg2)
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

// PauseCircuit is a paid mutator transaction binding the contract method 0x50424cd0.
//
// Solidity: function pauseCircuit() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) PauseCircuit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "pauseCircuit")
}

// PauseCircuit is a paid mutator transaction binding the contract method 0x50424cd0.
//
// Solidity: function pauseCircuit() returns()
func (_InvisiProtocol *InvisiProtocolSession) PauseCircuit() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PauseCircuit(&_InvisiProtocol.TransactOpts)
}

// PauseCircuit is a paid mutator transaction binding the contract method 0x50424cd0.
//
// Solidity: function pauseCircuit() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) PauseCircuit() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.PauseCircuit(&_InvisiProtocol.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiProtocol *InvisiProtocolSession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.RenounceOwnership(&_InvisiProtocol.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.RenounceOwnership(&_InvisiProtocol.TransactOpts)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x400b3547.
//
// Solidity: function revokeCredit(address token, address owner, uint64 shard, uint16 slot, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) RevokeCredit(opts *bind.TransactOpts, token common.Address, owner common.Address, shard uint64, slot uint16, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "revokeCredit", token, owner, shard, slot, deadline, feeRecipient, feeAmount, signature)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x400b3547.
//
// Solidity: function revokeCredit(address token, address owner, uint64 shard, uint16 slot, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiProtocol *InvisiProtocolSession) RevokeCredit(token common.Address, owner common.Address, shard uint64, slot uint16, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.RevokeCredit(&_InvisiProtocol.TransactOpts, token, owner, shard, slot, deadline, feeRecipient, feeAmount, signature)
}

// RevokeCredit is a paid mutator transaction binding the contract method 0x400b3547.
//
// Solidity: function revokeCredit(address token, address owner, uint64 shard, uint16 slot, uint256 deadline, address feeRecipient, uint256 feeAmount, bytes signature) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) RevokeCredit(token common.Address, owner common.Address, shard uint64, slot uint16, deadline *big.Int, feeRecipient common.Address, feeAmount *big.Int, signature []byte) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.RevokeCredit(&_InvisiProtocol.TransactOpts, token, owner, shard, slot, deadline, feeRecipient, feeAmount, signature)
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

// Shield is a paid mutator transaction binding the contract method 0xae897f6a.
//
// Solidity: function shield((address,bytes32[],uint256,address,uint256) p, (bytes32,uint256)[] , (bytes,bytes)[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Shield(opts *bind.TransactOpts, p IInvisiProtocolShieldParams, arg1 []IInvisiProtocolShieldOutput, arg2 []IInvisiProtocolDelivery, arg3 IInvisiProtocolFundingParams) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "shield", p, arg1, arg2, arg3)
}

// Shield is a paid mutator transaction binding the contract method 0xae897f6a.
//
// Solidity: function shield((address,bytes32[],uint256,address,uint256) p, (bytes32,uint256)[] , (bytes,bytes)[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiProtocol *InvisiProtocolSession) Shield(p IInvisiProtocolShieldParams, arg1 []IInvisiProtocolShieldOutput, arg2 []IInvisiProtocolDelivery, arg3 IInvisiProtocolFundingParams) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Shield(&_InvisiProtocol.TransactOpts, p, arg1, arg2, arg3)
}

// Shield is a paid mutator transaction binding the contract method 0xae897f6a.
//
// Solidity: function shield((address,bytes32[],uint256,address,uint256) p, (bytes32,uint256)[] , (bytes,bytes)[] , (address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)) ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Shield(p IInvisiProtocolShieldParams, arg1 []IInvisiProtocolShieldOutput, arg2 []IInvisiProtocolDelivery, arg3 IInvisiProtocolFundingParams) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Shield(&_InvisiProtocol.TransactOpts, p, arg1, arg2, arg3)
}

// Transact is a paid mutator transaction binding the contract method 0x2eb90f9b.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) r) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Transact(opts *bind.TransactOpts, arg0 []byte, r IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "transact", arg0, r)
}

// Transact is a paid mutator transaction binding the contract method 0x2eb90f9b.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) r) returns()
func (_InvisiProtocol *InvisiProtocolSession) Transact(arg0 []byte, r IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact(&_InvisiProtocol.TransactOpts, arg0, r)
}

// Transact is a paid mutator transaction binding the contract method 0x2eb90f9b.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32))) r) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Transact(arg0 []byte, r IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact(&_InvisiProtocol.TransactOpts, arg0, r)
}

// Transact0 is a paid mutator transaction binding the contract method 0x870949a9.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] requests) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) Transact0(opts *bind.TransactOpts, arg0 []byte, requests []IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "transact0", arg0, requests)
}

// Transact0 is a paid mutator transaction binding the contract method 0x870949a9.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] requests) returns()
func (_InvisiProtocol *InvisiProtocolSession) Transact0(arg0 []byte, requests []IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact0(&_InvisiProtocol.TransactOpts, arg0, requests)
}

// Transact0 is a paid mutator transaction binding the contract method 0x870949a9.
//
// Solidity: function transact(bytes , ((address,(uint256,bytes32)[],bytes32[],bytes32[],address,uint256,uint256,address,uint256),(bytes,bytes)[],(address,uint256,uint256,uint256,bytes,(uint256,uint8,bytes32,bytes32)))[] requests) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) Transact0(arg0 []byte, requests []IInvisiProtocolTransactRequest) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.Transact0(&_InvisiProtocol.TransactOpts, arg0, requests)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiProtocol *InvisiProtocolSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.TransferOwnership(&_InvisiProtocol.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.TransferOwnership(&_InvisiProtocol.TransactOpts, newOwner)
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

// UnpauseCircuit is a paid mutator transaction binding the contract method 0xa3bf4880.
//
// Solidity: function unpauseCircuit() returns()
func (_InvisiProtocol *InvisiProtocolTransactor) UnpauseCircuit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "unpauseCircuit")
}

// UnpauseCircuit is a paid mutator transaction binding the contract method 0xa3bf4880.
//
// Solidity: function unpauseCircuit() returns()
func (_InvisiProtocol *InvisiProtocolSession) UnpauseCircuit() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.UnpauseCircuit(&_InvisiProtocol.TransactOpts)
}

// UnpauseCircuit is a paid mutator transaction binding the contract method 0xa3bf4880.
//
// Solidity: function unpauseCircuit() returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) UnpauseCircuit() (*types.Transaction, error) {
	return _InvisiProtocol.Contract.UnpauseCircuit(&_InvisiProtocol.TransactOpts)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0x2dab44e5.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactor) UpdateCredit(opts *bind.TransactOpts, arg0 []byte, p IInvisiProtocolUpdateCreditParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "updateCredit", arg0, p, arg2)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0x2dab44e5.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolSession) UpdateCredit(arg0 []byte, p IInvisiProtocolUpdateCreditParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.UpdateCredit(&_InvisiProtocol.TransactOpts, arg0, p, arg2)
}

// UpdateCredit is a paid mutator transaction binding the contract method 0x2dab44e5.
//
// Solidity: function updateCredit(bytes , (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) p, (bytes,bytes)[] ) returns()
func (_InvisiProtocol *InvisiProtocolTransactorSession) UpdateCredit(arg0 []byte, p IInvisiProtocolUpdateCreditParams, arg2 []IInvisiProtocolDelivery) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.UpdateCredit(&_InvisiProtocol.TransactOpts, arg0, p, arg2)
}

// WithdrawBroadcasterFees is a paid mutator transaction binding the contract method 0x099e7064.
//
// Solidity: function withdrawBroadcasterFees(address token, address recipient) returns(uint256)
func (_InvisiProtocol *InvisiProtocolTransactor) WithdrawBroadcasterFees(opts *bind.TransactOpts, token common.Address, recipient common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.contract.Transact(opts, "withdrawBroadcasterFees", token, recipient)
}

// WithdrawBroadcasterFees is a paid mutator transaction binding the contract method 0x099e7064.
//
// Solidity: function withdrawBroadcasterFees(address token, address recipient) returns(uint256)
func (_InvisiProtocol *InvisiProtocolSession) WithdrawBroadcasterFees(token common.Address, recipient common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.WithdrawBroadcasterFees(&_InvisiProtocol.TransactOpts, token, recipient)
}

// WithdrawBroadcasterFees is a paid mutator transaction binding the contract method 0x099e7064.
//
// Solidity: function withdrawBroadcasterFees(address token, address recipient) returns(uint256)
func (_InvisiProtocol *InvisiProtocolTransactorSession) WithdrawBroadcasterFees(token common.Address, recipient common.Address) (*types.Transaction, error) {
	return _InvisiProtocol.Contract.WithdrawBroadcasterFees(&_InvisiProtocol.TransactOpts, token, recipient)
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
	Tokens            []common.Address
	Commitments       [][32]byte
	PackedDeliveries  []byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterBatchOutputs is a free log retrieval operation binding the contract event 0xf5c3c1e590ac45e708a69991fe84df4bb1dc4c185bf0a1ec6794c31b776a1be0.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, address[] tokens, bytes32[] commitments, bytes packedDeliveries)
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

// WatchBatchOutputs is a free log subscription operation binding the contract event 0xf5c3c1e590ac45e708a69991fe84df4bb1dc4c185bf0a1ec6794c31b776a1be0.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, address[] tokens, bytes32[] commitments, bytes packedDeliveries)
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

// ParseBatchOutputs is a log parse operation binding the contract event 0xf5c3c1e590ac45e708a69991fe84df4bb1dc4c185bf0a1ec6794c31b776a1be0.
//
// Solidity: event BatchOutputs(uint256 indexed treeId, uint32 startLeafIndex, uint256 firstOutputOffset, address[] tokens, bytes32[] commitments, bytes packedDeliveries)
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
	Token             common.Address
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

// FilterBatchRequest is a free log retrieval operation binding the contract event 0xbceb7ab95985373d11b26c1454947414ae0b754a3b33e0f3c0d779bf540f9db5.
//
// Solidity: event BatchRequest(address token, uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
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

// WatchBatchRequest is a free log subscription operation binding the contract event 0xbceb7ab95985373d11b26c1454947414ae0b754a3b33e0f3c0d779bf540f9db5.
//
// Solidity: event BatchRequest(address token, uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
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

// ParseBatchRequest is a log parse operation binding the contract event 0xbceb7ab95985373d11b26c1454947414ae0b754a3b33e0f3c0d779bf540f9db5.
//
// Solidity: event BatchRequest(address token, uint256 indexed requestIndex, uint256 capacity, (uint256,bytes32)[] sources, bytes32[] nullifiers, uint256 outputCount, address payer, uint256 depositAmount, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline, bytes32 statementHash)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseBatchRequest(log types.Log) (*InvisiProtocolBatchRequest, error) {
	event := new(InvisiProtocolBatchRequest)
	if err := _InvisiProtocol.contract.UnpackLog(event, "BatchRequest", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolBroadcasterFeeIterator is returned from FilterBroadcasterFee and is used to iterate over the raw logs and unpacked data for BroadcasterFee events raised by the InvisiProtocol contract.
type InvisiProtocolBroadcasterFeeIterator struct {
	Event *InvisiProtocolBroadcasterFee // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolBroadcasterFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolBroadcasterFee)
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
		it.Event = new(InvisiProtocolBroadcasterFee)
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
func (it *InvisiProtocolBroadcasterFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolBroadcasterFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolBroadcasterFee represents a BroadcasterFee event raised by the InvisiProtocol contract.
type InvisiProtocolBroadcasterFee struct {
	Token     common.Address
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBroadcasterFee is a free log retrieval operation binding the contract event 0xe5428e81365966d5479efecda1c1de8cf8624f59d57290a0501f499baeb281aa.
//
// Solidity: event BroadcasterFee(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterBroadcasterFee(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*InvisiProtocolBroadcasterFeeIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "BroadcasterFee", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolBroadcasterFeeIterator{contract: _InvisiProtocol.contract, event: "BroadcasterFee", logs: logs, sub: sub}, nil
}

// WatchBroadcasterFee is a free log subscription operation binding the contract event 0xe5428e81365966d5479efecda1c1de8cf8624f59d57290a0501f499baeb281aa.
//
// Solidity: event BroadcasterFee(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchBroadcasterFee(opts *bind.WatchOpts, sink chan<- *InvisiProtocolBroadcasterFee, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "BroadcasterFee", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolBroadcasterFee)
				if err := _InvisiProtocol.contract.UnpackLog(event, "BroadcasterFee", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseBroadcasterFee(log types.Log) (*InvisiProtocolBroadcasterFee, error) {
	event := new(InvisiProtocolBroadcasterFee)
	if err := _InvisiProtocol.contract.UnpackLog(event, "BroadcasterFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolBroadcasterFeeWithdrawnIterator is returned from FilterBroadcasterFeeWithdrawn and is used to iterate over the raw logs and unpacked data for BroadcasterFeeWithdrawn events raised by the InvisiProtocol contract.
type InvisiProtocolBroadcasterFeeWithdrawnIterator struct {
	Event *InvisiProtocolBroadcasterFeeWithdrawn // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolBroadcasterFeeWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolBroadcasterFeeWithdrawn)
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
		it.Event = new(InvisiProtocolBroadcasterFeeWithdrawn)
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
func (it *InvisiProtocolBroadcasterFeeWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolBroadcasterFeeWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolBroadcasterFeeWithdrawn represents a BroadcasterFeeWithdrawn event raised by the InvisiProtocol contract.
type InvisiProtocolBroadcasterFeeWithdrawn struct {
	Token     common.Address
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBroadcasterFeeWithdrawn is a free log retrieval operation binding the contract event 0x1b7af30e773fc72482ff43806732398e343a2e6a689255c8f9236871c4128139.
//
// Solidity: event BroadcasterFeeWithdrawn(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterBroadcasterFeeWithdrawn(opts *bind.FilterOpts, token []common.Address, recipient []common.Address) (*InvisiProtocolBroadcasterFeeWithdrawnIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "BroadcasterFeeWithdrawn", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolBroadcasterFeeWithdrawnIterator{contract: _InvisiProtocol.contract, event: "BroadcasterFeeWithdrawn", logs: logs, sub: sub}, nil
}

// WatchBroadcasterFeeWithdrawn is a free log subscription operation binding the contract event 0x1b7af30e773fc72482ff43806732398e343a2e6a689255c8f9236871c4128139.
//
// Solidity: event BroadcasterFeeWithdrawn(address indexed token, address indexed recipient, uint256 amount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchBroadcasterFeeWithdrawn(opts *bind.WatchOpts, sink chan<- *InvisiProtocolBroadcasterFeeWithdrawn, token []common.Address, recipient []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "BroadcasterFeeWithdrawn", tokenRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolBroadcasterFeeWithdrawn)
				if err := _InvisiProtocol.contract.UnpackLog(event, "BroadcasterFeeWithdrawn", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseBroadcasterFeeWithdrawn(log types.Log) (*InvisiProtocolBroadcasterFeeWithdrawn, error) {
	event := new(InvisiProtocolBroadcasterFeeWithdrawn)
	if err := _InvisiProtocol.contract.UnpackLog(event, "BroadcasterFeeWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCircuitPauseUpdatedIterator is returned from FilterCircuitPauseUpdated and is used to iterate over the raw logs and unpacked data for CircuitPauseUpdated events raised by the InvisiProtocol contract.
type InvisiProtocolCircuitPauseUpdatedIterator struct {
	Event *InvisiProtocolCircuitPauseUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCircuitPauseUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCircuitPauseUpdated)
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
		it.Event = new(InvisiProtocolCircuitPauseUpdated)
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
func (it *InvisiProtocolCircuitPauseUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCircuitPauseUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCircuitPauseUpdated represents a CircuitPauseUpdated event raised by the InvisiProtocol contract.
type InvisiProtocolCircuitPauseUpdated struct {
	Paused bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCircuitPauseUpdated is a free log retrieval operation binding the contract event 0xadc9431437c5558e6a2e308a3be13eebadd57c3d88323a532a827a019aecb2d8.
//
// Solidity: event CircuitPauseUpdated(bool paused)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCircuitPauseUpdated(opts *bind.FilterOpts) (*InvisiProtocolCircuitPauseUpdatedIterator, error) {

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CircuitPauseUpdated")
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCircuitPauseUpdatedIterator{contract: _InvisiProtocol.contract, event: "CircuitPauseUpdated", logs: logs, sub: sub}, nil
}

// WatchCircuitPauseUpdated is a free log subscription operation binding the contract event 0xadc9431437c5558e6a2e308a3be13eebadd57c3d88323a532a827a019aecb2d8.
//
// Solidity: event CircuitPauseUpdated(bool paused)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCircuitPauseUpdated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCircuitPauseUpdated) (event.Subscription, error) {

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CircuitPauseUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCircuitPauseUpdated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CircuitPauseUpdated", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCircuitPauseUpdated(log types.Log) (*InvisiProtocolCircuitPauseUpdated, error) {
	event := new(InvisiProtocolCircuitPauseUpdated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CircuitPauseUpdated", log); err != nil {
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
	Token            common.Address
	Owner            common.Address
	OutputTreeId     *big.Int
	RevocationShard  uint64
	RevocationSlot   uint16
	CreditCommitment [32]byte
	LeafIndex        uint32
	Deliveries       []IInvisiProtocolDelivery
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterCreditCreated is a free log retrieval operation binding the contract event 0x3ff78f0284d33ce5183bc7c13bbedfddbf90722c4b3d90a5b8e59da9bbd70565.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint64 revocationShard, uint16 revocationSlot, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditCreated(opts *bind.FilterOpts, token []common.Address, owner []common.Address, outputTreeId []*big.Int) (*InvisiProtocolCreditCreatedIterator, error) {

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

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditCreated", tokenRule, ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditCreatedIterator{contract: _InvisiProtocol.contract, event: "CreditCreated", logs: logs, sub: sub}, nil
}

// WatchCreditCreated is a free log subscription operation binding the contract event 0x3ff78f0284d33ce5183bc7c13bbedfddbf90722c4b3d90a5b8e59da9bbd70565.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint64 revocationShard, uint16 revocationSlot, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditCreated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditCreated, token []common.Address, owner []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditCreated", tokenRule, ownerRule, outputTreeIdRule)
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

// ParseCreditCreated is a log parse operation binding the contract event 0x3ff78f0284d33ce5183bc7c13bbedfddbf90722c4b3d90a5b8e59da9bbd70565.
//
// Solidity: event CreditCreated(address indexed token, address indexed owner, uint256 indexed outputTreeId, uint64 revocationShard, uint16 revocationSlot, bytes32 creditCommitment, uint32 leafIndex, (bytes,bytes)[] deliveries, uint256 deadline, address feeRecipient, uint256 feeAmount)
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
	OutputTreeId *big.Int
	Params       IInvisiProtocolCreditDrawParams
	LeafIndices  []uint32
	Deliveries   []IInvisiProtocolDelivery
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditDraw is a free log retrieval operation binding the contract event 0xa68be504130022fbd729573715873c6c515c318a014d5273815e20abcc08439d.
//
// Solidity: event CreditDraw(uint256 indexed outputTreeId, (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) params, uint32[] leafIndices, (bytes,bytes)[] deliveries)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditDraw(opts *bind.FilterOpts, outputTreeId []*big.Int) (*InvisiProtocolCreditDrawIterator, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditDraw", outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditDrawIterator{contract: _InvisiProtocol.contract, event: "CreditDraw", logs: logs, sub: sub}, nil
}

// WatchCreditDraw is a free log subscription operation binding the contract event 0xa68be504130022fbd729573715873c6c515c318a014d5273815e20abcc08439d.
//
// Solidity: event CreditDraw(uint256 indexed outputTreeId, (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) params, uint32[] leafIndices, (bytes,bytes)[] deliveries)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditDraw(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditDraw, outputTreeId []*big.Int) (event.Subscription, error) {

	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditDraw", outputTreeIdRule)
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

// ParseCreditDraw is a log parse operation binding the contract event 0xa68be504130022fbd729573715873c6c515c318a014d5273815e20abcc08439d.
//
// Solidity: event CreditDraw(uint256 indexed outputTreeId, (address,(uint256,bytes32)[],(address,uint256)[],(uint256,uint64,uint16)[],bytes32[],bytes32[],bytes32[],bytes32,uint256,uint256,address,uint256) params, uint32[] leafIndices, (bytes,bytes)[] deliveries)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditDraw(log types.Log) (*InvisiProtocolCreditDraw, error) {
	event := new(InvisiProtocolCreditDraw)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditDraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditRevokedIterator is returned from FilterCreditRevoked and is used to iterate over the raw logs and unpacked data for CreditRevoked events raised by the InvisiProtocol contract.
type InvisiProtocolCreditRevokedIterator struct {
	Event *InvisiProtocolCreditRevoked // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditRevoked)
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
		it.Event = new(InvisiProtocolCreditRevoked)
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
func (it *InvisiProtocolCreditRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditRevoked represents a CreditRevoked event raised by the InvisiProtocol contract.
type InvisiProtocolCreditRevoked struct {
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
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditRevoked(opts *bind.FilterOpts, owner []common.Address, token []common.Address, shard []uint64) (*InvisiProtocolCreditRevokedIterator, error) {

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

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditRevoked", ownerRule, tokenRule, shardRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditRevokedIterator{contract: _InvisiProtocol.contract, event: "CreditRevoked", logs: logs, sub: sub}, nil
}

// WatchCreditRevoked is a free log subscription operation binding the contract event 0x417df449bda975b802268d8343b7ebb855bdc3f24c8b5c75d8772f56642c9710.
//
// Solidity: event CreditRevoked(address indexed owner, address indexed token, uint64 indexed shard, uint16 slot, address feeRecipient, uint256 feeAmount)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditRevoked(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditRevoked, owner []common.Address, token []common.Address, shard []uint64) (event.Subscription, error) {

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

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditRevoked", ownerRule, tokenRule, shardRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditRevoked)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditRevoked", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditRevoked(log types.Log) (*InvisiProtocolCreditRevoked, error) {
	event := new(InvisiProtocolCreditRevoked)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolCreditUpdatedIterator is returned from FilterCreditUpdated and is used to iterate over the raw logs and unpacked data for CreditUpdated events raised by the InvisiProtocol contract.
type InvisiProtocolCreditUpdatedIterator struct {
	Event *InvisiProtocolCreditUpdated // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolCreditUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolCreditUpdated)
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
		it.Event = new(InvisiProtocolCreditUpdated)
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
func (it *InvisiProtocolCreditUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolCreditUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolCreditUpdated represents a CreditUpdated event raised by the InvisiProtocol contract.
type InvisiProtocolCreditUpdated struct {
	Owner        common.Address
	OutputTreeId *big.Int
	Params       IInvisiProtocolUpdateCreditParams
	LeafIndex    uint32
	Deliveries   []IInvisiProtocolDelivery
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreditUpdated is a free log retrieval operation binding the contract event 0x4a3fdcaf61cef0446996557873d07793a373211d25c66949a80fe711b29b8d15.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterCreditUpdated(opts *bind.FilterOpts, owner []common.Address, outputTreeId []*big.Int) (*InvisiProtocolCreditUpdatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "CreditUpdated", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolCreditUpdatedIterator{contract: _InvisiProtocol.contract, event: "CreditUpdated", logs: logs, sub: sub}, nil
}

// WatchCreditUpdated is a free log subscription operation binding the contract event 0x4a3fdcaf61cef0446996557873d07793a373211d25c66949a80fe711b29b8d15.
//
// Solidity: event CreditUpdated(address indexed owner, uint256 indexed outputTreeId, (address,address,(uint256,bytes32),bytes32,bytes32,uint64,uint16,uint256,address,uint256) params, uint32 leafIndex, (bytes,bytes)[] deliveries)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchCreditUpdated(opts *bind.WatchOpts, sink chan<- *InvisiProtocolCreditUpdated, owner []common.Address, outputTreeId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var outputTreeIdRule []interface{}
	for _, outputTreeIdItem := range outputTreeId {
		outputTreeIdRule = append(outputTreeIdRule, outputTreeIdItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "CreditUpdated", ownerRule, outputTreeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolCreditUpdated)
				if err := _InvisiProtocol.contract.UnpackLog(event, "CreditUpdated", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCreditUpdated(log types.Log) (*InvisiProtocolCreditUpdated, error) {
	event := new(InvisiProtocolCreditUpdated)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CreditUpdated", log); err != nil {
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
	Token             common.Address
	OutputTreeId      *big.Int
	Sources           []IInvisiProtocolInputSource
	Nullifiers        [][32]byte
	OutputCommitments [][32]byte
	LeafIndices       []uint32
	Deliveries        []IInvisiProtocolDelivery
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCrossTreeTransact is a free log retrieval operation binding the contract event 0x4f0edefe49cf4f3b4346ad9b94338d41abb8bdd8fdcdd7fdb9debeb27b80113d.
//
// Solidity: event CrossTreeTransact(address token, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
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

// WatchCrossTreeTransact is a free log subscription operation binding the contract event 0x4f0edefe49cf4f3b4346ad9b94338d41abb8bdd8fdcdd7fdb9debeb27b80113d.
//
// Solidity: event CrossTreeTransact(address token, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
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

// ParseCrossTreeTransact is a log parse operation binding the contract event 0x4f0edefe49cf4f3b4346ad9b94338d41abb8bdd8fdcdd7fdb9debeb27b80113d.
//
// Solidity: event CrossTreeTransact(address token, uint256 indexed outputTreeId, (uint256,bytes32)[] sources, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) ParseCrossTreeTransact(log types.Log) (*InvisiProtocolCrossTreeTransact, error) {
	event := new(InvisiProtocolCrossTreeTransact)
	if err := _InvisiProtocol.contract.UnpackLog(event, "CrossTreeTransact", log); err != nil {
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

// InvisiProtocolOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the InvisiProtocol contract.
type InvisiProtocolOwnershipTransferStartedIterator struct {
	Event *InvisiProtocolOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolOwnershipTransferStarted)
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
		it.Event = new(InvisiProtocolOwnershipTransferStarted)
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
func (it *InvisiProtocolOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the InvisiProtocol contract.
type InvisiProtocolOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiProtocolOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolOwnershipTransferStartedIterator{contract: _InvisiProtocol.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *InvisiProtocolOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolOwnershipTransferStarted)
				if err := _InvisiProtocol.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseOwnershipTransferStarted(log types.Log) (*InvisiProtocolOwnershipTransferStarted, error) {
	event := new(InvisiProtocolOwnershipTransferStarted)
	if err := _InvisiProtocol.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// InvisiProtocolOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the InvisiProtocol contract.
type InvisiProtocolOwnershipTransferredIterator struct {
	Event *InvisiProtocolOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *InvisiProtocolOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InvisiProtocolOwnershipTransferred)
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
		it.Event = new(InvisiProtocolOwnershipTransferred)
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
func (it *InvisiProtocolOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InvisiProtocolOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InvisiProtocolOwnershipTransferred represents a OwnershipTransferred event raised by the InvisiProtocol contract.
type InvisiProtocolOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*InvisiProtocolOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolOwnershipTransferredIterator{contract: _InvisiProtocol.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *InvisiProtocolOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InvisiProtocolOwnershipTransferred)
				if err := _InvisiProtocol.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_InvisiProtocol *InvisiProtocolFilterer) ParseOwnershipTransferred(log types.Log) (*InvisiProtocolOwnershipTransferred, error) {
	event := new(InvisiProtocolOwnershipTransferred)
	if err := _InvisiProtocol.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
	Token     common.Address
	Payer     common.Address
	Amount    *big.Int
	ProofHash [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPublicDeposit is a free log retrieval operation binding the contract event 0x1626a1d6f055052717d0a40e4bb7e02f2b2b379432abf8fcd2ef80d2374e6661.
//
// Solidity: event PublicDeposit(address indexed token, address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterPublicDeposit(opts *bind.FilterOpts, token []common.Address, payer []common.Address) (*InvisiProtocolPublicDepositIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "PublicDeposit", tokenRule, payerRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolPublicDepositIterator{contract: _InvisiProtocol.contract, event: "PublicDeposit", logs: logs, sub: sub}, nil
}

// WatchPublicDeposit is a free log subscription operation binding the contract event 0x1626a1d6f055052717d0a40e4bb7e02f2b2b379432abf8fcd2ef80d2374e6661.
//
// Solidity: event PublicDeposit(address indexed token, address indexed payer, uint256 amount, bytes32 proofHash)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchPublicDeposit(opts *bind.WatchOpts, sink chan<- *InvisiProtocolPublicDeposit, token []common.Address, payer []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "PublicDeposit", tokenRule, payerRule)
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

// ParsePublicDeposit is a log parse operation binding the contract event 0x1626a1d6f055052717d0a40e4bb7e02f2b2b379432abf8fcd2ef80d2374e6661.
//
// Solidity: event PublicDeposit(address indexed token, address indexed payer, uint256 amount, bytes32 proofHash)
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
	Token          common.Address
	Payer          common.Address
	Amount         *big.Int
	TreeId         *big.Int
	StartLeafIndex uint32
	Deliveries     []IInvisiProtocolDelivery
	Deadline       *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterShield is a free log retrieval operation binding the contract event 0xefe7bab38d42993e5331d7dd43bbf9a0e1fa493cdd18f5d6d0b980f5b324f0d1.
//
// Solidity: event Shield(address indexed token, address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, (bytes,bytes)[] deliveries, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) FilterShield(opts *bind.FilterOpts, token []common.Address, payer []common.Address, treeId []*big.Int) (*InvisiProtocolShieldIterator, error) {

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

	logs, sub, err := _InvisiProtocol.contract.FilterLogs(opts, "Shield", tokenRule, payerRule, treeIdRule)
	if err != nil {
		return nil, err
	}
	return &InvisiProtocolShieldIterator{contract: _InvisiProtocol.contract, event: "Shield", logs: logs, sub: sub}, nil
}

// WatchShield is a free log subscription operation binding the contract event 0xefe7bab38d42993e5331d7dd43bbf9a0e1fa493cdd18f5d6d0b980f5b324f0d1.
//
// Solidity: event Shield(address indexed token, address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, (bytes,bytes)[] deliveries, uint256 deadline)
func (_InvisiProtocol *InvisiProtocolFilterer) WatchShield(opts *bind.WatchOpts, sink chan<- *InvisiProtocolShield, token []common.Address, payer []common.Address, treeId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _InvisiProtocol.contract.WatchLogs(opts, "Shield", tokenRule, payerRule, treeIdRule)
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

// ParseShield is a log parse operation binding the contract event 0xefe7bab38d42993e5331d7dd43bbf9a0e1fa493cdd18f5d6d0b980f5b324f0d1.
//
// Solidity: event Shield(address indexed token, address indexed payer, uint256 amount, uint256 indexed treeId, uint32 startLeafIndex, (bytes,bytes)[] deliveries, uint256 deadline)
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
	Token             common.Address
	InputTreeId       *big.Int
	OutputTreeId      *big.Int
	Nullifiers        [][32]byte
	OutputCommitments [][32]byte
	LeafIndices       []uint32
	Deliveries        []IInvisiProtocolDelivery
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterTransact is a free log retrieval operation binding the contract event 0xac5599eb2d39ff304a3cfae313226db689176cd11a081b088a08b2ade0c88382.
//
// Solidity: event Transact(address token, uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
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

// WatchTransact is a free log subscription operation binding the contract event 0xac5599eb2d39ff304a3cfae313226db689176cd11a081b088a08b2ade0c88382.
//
// Solidity: event Transact(address token, uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
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

// ParseTransact is a log parse operation binding the contract event 0xac5599eb2d39ff304a3cfae313226db689176cd11a081b088a08b2ade0c88382.
//
// Solidity: event Transact(address token, uint256 indexed inputTreeId, uint256 indexed outputTreeId, bytes32[] nullifiers, bytes32[] outputCommitments, uint32[] leafIndices, (bytes,bytes)[] deliveries, address withdrawalAddress, uint256 withdrawalAmount, uint256 deadline)
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
