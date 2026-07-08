package payer

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// Two-step client-verification sentinels (IPP 6.2 authorize). CreateCredit returns
// one of these when the accepted authorize response is internally inconsistent;
// the caller MUST abort and never depositCredit.
var (
	ErrCreditKeyHashMismatch    = errors.New("payer: credit_nullifier_key does not hash to the server nullifier_key_hash")
	ErrCreditCommitmentMismatch = errors.New("payer: recomputed credit commitment does not match the server value")
)

// CreateCredit consumes an accepted IPP authorize response and builds the
// depositCredit(identityCommitment, payerCommitment, amount, nonce, expiry,
// nullifierKeyHash, treeId) calldata.
//
// creditNullifierKey is derived server-side from the receiver's master key and
// delivered in the authorize response; the payer does not derive it locally.
// CreateCredit therefore:
//  1. recomputes nullifierKeyHash = Poseidon2(D_NKHASH, creditNullifierKey) locally,
//  2. verifies it equals the server's nullifier_key_hash (else ErrCreditKeyHashMismatch),
//  3. recomputes the creditCommitment with that local hash and verifies it equals
//     the server value (else ErrCreditCommitmentMismatch),
//  4. encodes depositCredit(...) using the LOCALLY recomputed hash — the payer's
//     reclaim safety anchor: the on-chain nullifierKeyHash is provably the hash of
//     the key the payer holds, so reclaim never depends on trusting the server.
func CreateCredit(params CreateCreditParams) (*CreateCreditResult, error) {
	if params.Amount == nil {
		return nil, fmt.Errorf("payer: Amount is required")
	}
	if params.Expiry == nil {
		return nil, fmt.Errorf("payer: Expiry is required")
	}
	if params.TreeID == nil {
		return nil, fmt.Errorf("payer: TreeID is required")
	}
	if params.ChainID == nil {
		return nil, fmt.Errorf("payer: ChainID is required")
	}
	if params.ServerNullifierKeyHash == (common.Hash{}) {
		return nil, fmt.Errorf("payer: ServerNullifierKeyHash is required")
	}
	if params.ServerCreditCommitment == (common.Hash{}) {
		return nil, fmt.Errorf("payer: ServerCreditCommitment is required")
	}

	// Step 1: recompute nullifierKeyHash from the server-delivered key.
	nullifierKeyHash, err := core.ComputeNullifierKeyHash(params.CreditNullifierKey)
	if err != nil {
		return nil, fmt.Errorf("payer: compute nullifier key hash: %w", err)
	}
	nkHash := common.HexToHash(core.ElementToHex(nullifierKeyHash))

	// Step 2 (client verification #1): the delivered key must hash to the
	// server-returned nullifier_key_hash.
	if nkHash != params.ServerNullifierKeyHash {
		return nil, ErrCreditKeyHashMismatch
	}

	// Step 3 (client verification #2): recompute the credit commitment with the
	// local hash and confirm it equals the server value (nonce folded correctly).
	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid identityCommitment: %w", err)
	}
	pcElem, err := core.HexToElement(params.PayerCommitment.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid payerCommitment: %w", err)
	}
	tokenElem, err := core.HexToElement(addressToTokenElement(params.TokenAddress).Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid tokenAddress: %w", err)
	}
	creditCommitment, err := core.ComputeCreditCommitment(
		core.BigIntToElement(params.ChainID), icElem, pcElem, tokenElem,
		core.BigIntToElement(params.Amount), params.Nonce,
		core.BigIntToElement(params.Expiry), nullifierKeyHash,
	)
	if err != nil {
		return nil, fmt.Errorf("payer: compute credit commitment: %w", err)
	}
	if common.HexToHash(core.ElementToHex(creditCommitment)) != params.ServerCreditCommitment {
		return nil, ErrCreditCommitmentMismatch
	}

	// Step 4: encode depositCredit(...) with the locally recomputed hash.
	nonceHash := common.HexToHash(core.ElementToHex(params.Nonce))
	creditIDHash := common.HexToHash(core.ElementToHex(params.CreditID))

	data, err := packCalldata(
		invisiProtocolABI,
		"depositCredit",
		params.IdentityCommitment, // bytes32
		params.PayerCommitment,    // bytes32
		params.Amount,             // uint256
		nonceHash,                 // bytes32
		params.Expiry,             // uint256
		nkHash,                    // bytes32 — locally recomputed (reclaim anchor)
		params.TreeID,             // uint256
	)
	if err != nil {
		return nil, err
	}

	return &CreateCreditResult{
		Authorization: CreditAuthorization{
			CreditID:           creditIDHash,
			IdentityCommitment: params.IdentityCommitment,
			PayerCommitment:    params.PayerCommitment,
			CreditNullifierKey: params.CreditNullifierKey,
			NullifierKeyHash:   nkHash,
			TokenAddress:       params.TokenAddress,
			Amount:             new(big.Int).Set(params.Amount),
			Nonce:              params.Nonce,
			Expiry:             new(big.Int).Set(params.Expiry),
			TreeID:             new(big.Int).Set(params.TreeID),
			ChainID:            new(big.Int).Set(params.ChainID),
		},
		Data:          data,
		ApproveAmount: new(big.Int).Set(params.Amount),
	}, nil
}

// DeriveCreditAuthorizeFields derives the single payer-side value — payerCommitment —
// needed to populate an IPP credit authorize REQUEST, before the server-derived
// creditNullifierKey is known. The payer cannot compute nullifierKeyHash at this
// stage — that needs the key, which arrives only in the authorize response — so
// this helper returns payerCommitment alone.
//
//	payerNullifierKey = Poseidon2(D_NKEY, payerMasterKey, creditId)
//	payerCommitment   = Poseidon2(D_IDENT, payerNullifierKey, payerAddress)
func DeriveCreditAuthorizeFields(payerMasterKey, creditID, payerAddress fr.Element) (string, error) {
	payerNullifierKey, err := core.DeriveNullifierKey(payerMasterKey, creditID)
	if err != nil {
		return "", fmt.Errorf("payer: derive payer nullifier key: %w", err)
	}
	payerCommitment, err := core.ComputeIdentityCommitment(payerNullifierKey, payerAddress)
	if err != nil {
		return "", fmt.Errorf("payer: compute payer commitment: %w", err)
	}
	return core.ElementToHex(payerCommitment), nil
}

// BuildReclaimCalldata encodes reclaim(proof, merkleRoot, nullifier, amount, receivingAddress, treeId, expiry).
// Used by the payer to reclaim an expired credit deposit.
func BuildReclaimCalldata(params ReclaimParams) (*ReclaimCalldata, error) {
	if params.Amount == nil {
		return nil, fmt.Errorf("payer: Amount is required")
	}
	if params.TreeID == nil {
		return nil, fmt.Errorf("payer: TreeID is required")
	}
	if params.Expiry == nil {
		return nil, fmt.Errorf("payer: Expiry is required")
	}

	// ABI encode: reclaim(proof, merkleRoot, nullifier, amount, receivingAddress, treeId, expiry)
	data, err := packCalldata(
		invisiProtocolABI,
		"reclaim",
		params.Proof,            // bytes
		params.MerkleRoot,       // bytes32
		params.Nullifier,        // bytes32
		params.Amount,           // uint256
		params.ReceivingAddress, // address
		params.TreeID,           // uint256
		params.Expiry,           // uint256
	)
	if err != nil {
		return nil, err
	}

	return &ReclaimCalldata{
		Data:      data,
		Nullifier: params.Nullifier,
	}, nil
}
