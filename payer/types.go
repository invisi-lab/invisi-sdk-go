package payer

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// ─── Gas Estimates ──────────────────────────────────────────────────

// GasEstimate holds recommended gas parameters for EIP-4337 UserOps.
type GasEstimate struct {
	CallGasLimit         uint64
	VerificationGasLimit uint64
	PreVerificationGas   uint64
}

var (
	GasEstimatePayOpen = GasEstimate{
		CallGasLimit:         150_000,
		VerificationGasLimit: 70_000,
		PreVerificationGas:   21_000,
	}
	GasEstimatePayBlind = GasEstimate{
		CallGasLimit:         350_000,
		VerificationGasLimit: 70_000,
		PreVerificationGas:   21_000,
	}
	GasEstimateDepositCredit = GasEstimate{
		CallGasLimit:         180_000,
		VerificationGasLimit: 70_000,
		PreVerificationGas:   21_000,
	}
)

// ─── PayOpen ────────────────────────────────────────────────────────

// PayOpenParams configures a public (open) payment.
type PayOpenParams struct {
	ChainID            *big.Int
	IdentityCommitment common.Hash
	TokenAddress       common.Address
	Amount             *big.Int
	TreeID             *big.Int
	RefundLink         *fr.Element // optional; nil = zero
	Nonce              *fr.Element // required; nil -> error (caller supplies per payment mode)
}

// PayOpenCalldata is the result of BuildPayOpenCalldata.
type PayOpenCalldata struct {
	PaymentCommitment common.Hash
	Nonce             common.Hash
	Data              []byte   // ABI-encoded payOpen(...) calldata
	ApproveAmount     *big.Int // ERC20 approve amount (== params.Amount)
}

// ─── PayBlind ───────────────────────────────────────────────────────

// PayBlindParams configures a shielded (blind) payment.
// Proof comes from a remote prover — this function only assembles calldata.
type PayBlindParams struct {
	Proof             []byte
	PaymentCommitment common.Hash
	ChangeCommitment  common.Hash
	InputAmount       *big.Int
	TreeID            *big.Int
	RefundLink        *fr.Element // optional; nil = zero
}

// PayBlindCalldata is the result of BuildPayBlindCalldata.
type PayBlindCalldata struct {
	Data          []byte   // ABI-encoded payBlind(...) calldata
	ApproveAmount *big.Int // ERC20 approve amount (== params.InputAmount)
}

// ─── SettleChange ───────────────────────────────────────────────────

// SettleChangeParams configures a change note settlement (single withdraw).
type SettleChangeParams struct {
	ChainID          *big.Int
	NullifierKey     fr.Element
	ReceivingAddress common.Address
	TokenAddress     common.Address
	Amount           *big.Int
	Nonce            fr.Element
	MerkleRoot       common.Hash
	Proof            []byte
	TreeID           *big.Int
}

// SettleChangeCalldata is the result of BuildSettleChangeCalldata.
type SettleChangeCalldata struct {
	Nullifier common.Hash
	Data      []byte // ABI-encoded withdraw(...) calldata
}

// ─── Credit ─────────────────────────────────────────────────────────

// CreateCreditParams configures the payer's post-authorize credit deposit build.
//
// The payer does not derive creditNullifierKey locally: it arrives in the
// authorize response (CreditNullifierKey), together with the server-assigned Nonce
// and the verification anchors ServerNullifierKeyHash / ServerCreditCommitment.
// PayerCommitment is the value the payer derived pre-authorize (see
// DeriveCreditAuthorizeFields) and sent in the authorize request.
type CreateCreditParams struct {
	CreditID           fr.Element  // from the one-time credit code's credit_id
	IdentityCommitment common.Hash // Receiver's identity commitment (from the credit code)
	PayerCommitment    common.Hash // payer-derived pre-authorize; echoed into the deposit leaf
	CreditNullifierKey fr.Element  // server-derived, delivered by the authorize response (TLS)
	TokenAddress       common.Address
	Amount             *big.Int
	Nonce              fr.Element // server-assigned, from the authorize response
	Expiry             *big.Int
	TreeID             *big.Int
	ChainID            *big.Int
	// Verification anchors from the authorize response.
	ServerNullifierKeyHash common.Hash // authorize response nullifier_key_hash
	ServerCreditCommitment common.Hash // authorize response credit_commitment
}

// CreditAuthorization contains all credit-related data the receiver needs.
type CreditAuthorization struct {
	CreditID           common.Hash
	IdentityCommitment common.Hash // Receiver's
	PayerCommitment    common.Hash
	CreditNullifierKey fr.Element
	NullifierKeyHash   common.Hash
	TokenAddress       common.Address
	Amount             *big.Int
	Nonce              fr.Element
	Expiry             *big.Int
	TreeID             *big.Int
	ChainID            *big.Int
}

// CreateCreditResult is the result of CreateCredit.
type CreateCreditResult struct {
	Authorization CreditAuthorization
	Data          []byte // ABI-encoded depositCredit(...) calldata
	ApproveAmount *big.Int
}

// ReclaimParams configures a credit reclaim by the payer.
type ReclaimParams struct {
	Nullifier        common.Hash
	Amount           *big.Int
	ReceivingAddress common.Address
	MerkleRoot       common.Hash
	Proof            []byte
	TreeID           *big.Int
	Expiry           *big.Int
}

// ReclaimCalldata is the result of BuildReclaimCalldata.
type ReclaimCalldata struct {
	Data      []byte // ABI-encoded reclaim(...) calldata
	Nullifier common.Hash
}

// ─── EIP-712 Meta-Transaction ───────────────────────────────────────

// EIP712DomainParams configures the EIP-712 domain.
type EIP712DomainParams struct {
	ChainID           *big.Int
	VerifyingContract common.Address
}

// SignPayOpenParams configures EIP-712 typed data for payOpen.
type SignPayOpenParams struct {
	ChainID            *big.Int
	VerifyingContract  common.Address
	IdentityCommitment common.Hash
	TokenAddress       common.Address
	Amount             *big.Int
	TreeID             *big.Int
	RefundLink         *common.Hash // optional; nil = zero
	Nonce              *fr.Element  // required; nil -> error (caller supplies per payment mode)
	SigNonce           *big.Int
	Deadline           *big.Int
}

// SignPayOpenResult is the result of BuildSignPayOpenTypedData.
type SignPayOpenResult struct {
	PaymentCommitment common.Hash
	Nonce             common.Hash
	TypedData         apitypes.TypedData
}

// SignPayBlindParams configures EIP-712 typed data for payBlind.
type SignPayBlindParams struct {
	ChainID           *big.Int
	VerifyingContract common.Address
	Proof             []byte
	PaymentCommitment common.Hash
	ChangeCommitment  common.Hash
	InputAmount       *big.Int
	TreeID            *big.Int
	RefundLink        *common.Hash // optional; nil = zero
	SigNonce          *big.Int
	Deadline          *big.Int
}

// SignPayBlindResult is the result of BuildSignPayBlindTypedData.
type SignPayBlindResult struct {
	TypedData apitypes.TypedData
}

// SignDepositCreditParams configures EIP-712 typed data for depositCredit.
type SignDepositCreditParams struct {
	ChainID            *big.Int
	VerifyingContract  common.Address
	IdentityCommitment common.Hash
	PayerCommitment    common.Hash
	Amount             *big.Int
	Nonce              common.Hash
	Expiry             *big.Int
	NullifierKeyHash   common.Hash
	TreeID             *big.Int
	SigNonce           *big.Int
	Deadline           *big.Int
}

// SignDepositCreditResult is the result of BuildSignDepositCreditTypedData.
type SignDepositCreditResult struct {
	TypedData apitypes.TypedData
}

// ─── EIP-4337 UserOperation ─────────────────────────────────────────

// PaymasterConfig holds paymaster service parameters for UserOps.
type PaymasterConfig struct {
	URL      string
	PolicyID string
	Context  map[string]interface{}
}

// BaseUserOp is the common return type for all UserOp builders.
type BaseUserOp struct {
	To                   common.Address
	Sender               common.Address
	CallData             []byte
	CallGasLimit         uint64
	VerificationGasLimit uint64
	PreVerificationGas   uint64
	Paymaster            *PaymasterConfig // optional
}

// PayOpenUserOpParams configures a UserOp for payOpen.
type PayOpenUserOpParams struct {
	IdentityCommitment common.Hash
	TokenAddress       common.Address
	Amount             *big.Int
	TreeID             *big.Int
	RefundLink         *fr.Element // optional
	Nonce              *fr.Element // required; nil -> error (caller supplies per payment mode)
	ProtocolAddress    common.Address
	Sender             common.Address
	ChainID            *big.Int
	Paymaster          *PaymasterConfig
}

// PayOpenUserOp extends BaseUserOp with payOpen-specific fields.
type PayOpenUserOp struct {
	BaseUserOp
	PaymentCommitment common.Hash
	Nonce             common.Hash
}

// PayBlindUserOpParams configures a UserOp for payBlind.
type PayBlindUserOpParams struct {
	Proof             []byte
	PaymentCommitment common.Hash
	ChangeCommitment  common.Hash
	InputAmount       *big.Int
	TreeID            *big.Int
	RefundLink        *fr.Element // optional
	ProtocolAddress   common.Address
	Sender            common.Address
	ChainID           *big.Int
	Paymaster         *PaymasterConfig
}

// DepositCreditUserOpParams configures a UserOp for depositCredit.
type DepositCreditUserOpParams struct {
	IdentityCommitment common.Hash
	PayerCommitment    common.Hash
	Amount             *big.Int
	Nonce              common.Hash
	Expiry             *big.Int
	NullifierKeyHash   common.Hash
	TreeID             *big.Int
	ProtocolAddress    common.Address
	Sender             common.Address
	ChainID            *big.Int
	Paymaster          *PaymasterConfig
}

// ─── Utilities ──────────────────────────────────────────────────────

// ApproveParams configures an ERC20 approve calldata build.
type ApproveParams struct {
	ProtocolAddress common.Address // spender
	Token           common.Address
	Amount          *big.Int
}

// ApproveCalldata is the result of BuildApproveCalldata.
type ApproveCalldata struct {
	To       common.Address // token contract address
	CallData []byte         // ABI-encoded approve(spender, amount)
}

// VerifyParams configures a payment commitment verification.
type VerifyParams struct {
	IdentityCommitment common.Hash
	TokenAddress       common.Address
	Amount             *big.Int
	Nonce              common.Hash
	PaymentCommitment  common.Hash
	ChainID            *big.Int
}

// BuildFromRequestParams configures building calldata from an IPP request.
type BuildFromRequestParams struct {
	Request           ipp.IPPRequest
	Nonce             common.Hash
	PaymentCommitment common.Hash
	Amount            *big.Int
	// TreeID now comes from the propose response, not the request payload (SPEC-029 §3).
	TreeID *big.Int
}

// BuildFromRequestResult is the result of BuildPaymentFromRequest.
type BuildFromRequestResult struct {
	Calldata      []byte
	ApproveAmount *big.Int
}
