package receiver

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
	"github.com/invisi-lab/invisi-sdk-go/tree"
)

// ── Invoice ─────────────────────────────────────────────────────────

// GenerateInvoiceParams holds inputs for GenerateInvoice.
type GenerateInvoiceParams struct {
	MasterKey        fr.Element
	InvoiceID        fr.Element
	ReceivingAddress common.Address
	TokenAddress     common.Address
	Amount           *big.Int
	TreeID           *big.Int
	Operator         common.Address
}

// Invoice contains the generated invoice data.
// NullifierKey is sensitive and must never be transmitted or logged.
type Invoice struct {
	InvoiceID          fr.Element
	NullifierKey       fr.Element // @sensitive — never transmit
	IdentityCommitment common.Hash
	ReceivingAddress   common.Address
	TokenAddress       common.Address
	Amount             *big.Int
	TreeID             *big.Int
	Operator           common.Address
}

// ── Payment Request ─────────────────────────────────────────────────

// CreatePaymentRequestParams holds inputs for CreatePaymentRequest.
type CreatePaymentRequestParams struct {
	MasterKey        fr.Element
	InvoiceID        fr.Element
	ReceivingAddress common.Address
	TokenAddress     common.Address
	ChainID          *big.Int
	Operator         common.Address
	PaymentTypes     []ipp.PaymentType
	// Mode is the receive-request mode (SPEC-029 §1): "single" (1a) or "multi" (1b/2).
	Mode     string
	Amount   *big.Int          // optional
	Endpoint *string           // optional
	Meta     map[string]string // optional
}

// ── Propose ─────────────────────────────────────────────────────────

// ProposeParams holds inputs for ComputePaymentCommitmentForPropose.
type ProposeParams struct {
	IdentityCommitment common.Hash
	TokenAddress       common.Address
	Amount             *big.Int
	ChainID            *big.Int
}

// ProposeResult holds the output of ComputePaymentCommitmentForPropose.
type ProposeResult struct {
	Nonce             common.Hash
	PaymentCommitment common.Hash
}

// ── Batch Circuit Selection ─────────────────────────────────────────

// BatchCircuitInfo identifies a batch ZK circuit and its maximum capacity.
type BatchCircuitInfo struct {
	Name    string
	MaxSize int
}

// ── Withdraw Witness ────────────────────────────────────────────────

// BuildWithdrawWitnessParams holds inputs for BuildWithdrawWitness.
type BuildWithdrawWitnessParams struct {
	NullifierKeys    []fr.Element
	ReceivingAddress common.Address
	TokenAddress     common.Address
	Amounts          []*big.Int
	Nonces           []fr.Element
	MerkleProofs     []tree.MerkleProof
}

// WithdrawWitness contains all ZK circuit inputs for a batch settlement proof.
// This is the Go SDK's key differentiator from the TS SDK — witness is built
// locally, then sent to a remote prover for proof generation.
type WithdrawWitness struct {
	NullifierKeys []fr.Element
	Amounts       []*big.Int
	Nonces        []fr.Element
	MerkleProofs  []tree.MerkleProof
	TreeRoot      fr.Element
	TotalAmount   *big.Int
	Receiver      common.Address
}

// ── Withdraw Calldata ───────────────────────────────────────────────

// WithdrawParams holds inputs for BuildWithdrawCalldata.
type WithdrawParams struct {
	Nullifiers       []common.Hash
	TotalAmount      *big.Int
	ReceivingAddress common.Address
	MerkleRoot       common.Hash
	Proof            []byte
	TreeID           *big.Int
}

// WithdrawCalldata holds the encoded contract calldata for a withdraw transaction.
type WithdrawCalldata struct {
	Data        []byte
	Nullifiers  []common.Hash
	TotalAmount *big.Int
}

// ── Credit Draw ─────────────────────────────────────────────────────

// BuildCreditDrawWitnessParams holds inputs for BuildCreditDrawWitness.
type BuildCreditDrawWitnessParams struct {
	NullifierKeys      []fr.Element
	ReceivingAddress   common.Address
	TokenAddress       common.Address
	Amounts            []*big.Int
	Nonces             []fr.Element
	Expiries           []*big.Int
	PayerCommitments   []fr.Element
	NullifierKeyHashes []fr.Element
	MerkleProofs       []tree.MerkleProof
}

// CreditDrawWitness contains all ZK circuit inputs for a credit draw proof.
type CreditDrawWitness struct {
	NullifierKeys        []fr.Element
	Amounts              []*big.Int
	Nonces               []fr.Element
	MerkleProofs         []tree.MerkleProof
	TreeRoot             fr.Element
	TotalAmount          *big.Int
	Receiver             common.Address
	RemainderCommitments []fr.Element
	Expiries             []*big.Int
	PayerCommitments     []fr.Element
	NullifierKeyHashes   []fr.Element
}

// CreditDrawParams holds inputs for BuildCreditDrawCalldata.
type CreditDrawParams struct {
	Nullifiers           []common.Hash
	TotalAmount          *big.Int
	ReceivingAddress     common.Address
	MerkleRoot           common.Hash
	Proof                []byte
	TreeID               *big.Int
	RemainderCommitments []common.Hash
	Expiry               *big.Int
}

// CreditDrawCalldata holds the encoded contract calldata for a creditDraw transaction.
type CreditDrawCalldata struct {
	Data        []byte
	Nullifiers  []common.Hash
	TotalAmount *big.Int
}

// ── Credit Reclaim ──────────────────────────────────────────────────

// CreditReclaimParams holds inputs for BuildCreditReclaimCalldata.
type CreditReclaimParams struct {
	Nullifier        common.Hash
	Amount           *big.Int
	ReceivingAddress common.Address
	MerkleRoot       common.Hash
	Proof            []byte
	TreeID           *big.Int
	Expiry           *big.Int
}

// CreditReclaimCalldata holds the encoded contract calldata for a reclaim transaction.
type CreditReclaimCalldata struct {
	Data []byte
}

// ── Refund ──────────────────────────────────────────────────────────

// RefundPayOpenParams holds inputs for BuildRefundPayOpenCalldata.
type RefundPayOpenParams struct {
	IdentityCommitment common.Hash
	RefundAmount       *big.Int
	Nonce              common.Hash
	TreeID             *big.Int
	RefundLink         common.Hash
}

// RefundPayOpenCalldata holds the encoded contract calldata for a refund payOpen transaction.
type RefundPayOpenCalldata struct {
	Data []byte
}
