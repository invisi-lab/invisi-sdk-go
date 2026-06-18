package payer

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// CreateCredit computes all credit-related commitments and encodes
// depositCredit(identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId).
//
// Derivation steps:
//  1. payerNullifierKey  = DeriveNullifierKey(payerMasterKey, creditId)
//  2. payerCommitment    = ComputeIdentityCommitment(payerNullifierKey, payerAddress)
//  3. creditNullifierKey = Poseidon2(D_DERIVE, payerNullifierKey, creditId)
//  4. nullifierKeyHash   = ComputeNullifierKeyHash(creditNullifierKey)
//  5. nonce              = random
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

	// Step 1: payerNullifierKey = DeriveNullifierKey(payerMasterKey, creditId)
	payerNullifierKey, err := core.DeriveNullifierKey(params.PayerMasterKey, params.CreditID)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to derive payer nullifier key: %w", err)
	}

	// Step 2: payerCommitment = ComputeIdentityCommitment(payerNullifierKey, payerAddress)
	payerCommitment, err := core.ComputeIdentityCommitment(payerNullifierKey, params.PayerAddress)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to compute payer commitment: %w", err)
	}

	// Step 3: creditNullifierKey = Poseidon2(D_DERIVE, payerNullifierKey, creditId)
	domainDerive := core.ElementFromUint64(core.DomainDerive)
	creditNullifierKey, err := core.Poseidon2Hash(domainDerive, payerNullifierKey, params.CreditID)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to derive credit nullifier key: %w", err)
	}

	// Step 4: nullifierKeyHash = ComputeNullifierKeyHash(creditNullifierKey)
	nullifierKeyHash, err := core.ComputeNullifierKeyHash(creditNullifierKey)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to compute nullifier key hash: %w", err)
	}

	// Step 5: nonce = random
	nonce, err := core.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("payer: failed to generate nonce: %w", err)
	}

	// Convert to hashes for ABI encoding
	payerCommHash := common.HexToHash(core.ElementToHex(payerCommitment))
	nonceHash := common.HexToHash(core.ElementToHex(nonce))
	nkHash := common.HexToHash(core.ElementToHex(nullifierKeyHash))
	creditIDHash := common.HexToHash(core.ElementToHex(params.CreditID))

	// ABI encode: depositCredit(identityCommitment, payerCommitment, amount, nonce, expiry, nullifierKeyHash, treeId)
	data, err := packCalldata(
		invisiProtocolABI,
		"depositCredit",
		params.IdentityCommitment, // bytes32
		payerCommHash,             // bytes32
		params.Amount,             // uint256
		nonceHash,                 // bytes32
		params.Expiry,             // uint256
		nkHash,                    // bytes32
		params.TreeID,             // uint256
	)
	if err != nil {
		return nil, err
	}

	return &CreateCreditResult{
		Authorization: CreditAuthorization{
			CreditID:           creditIDHash,
			IdentityCommitment: params.IdentityCommitment,
			PayerCommitment:    payerCommHash,
			CreditNullifierKey: creditNullifierKey,
			NullifierKeyHash:   nkHash,
			TokenAddress:       params.TokenAddress,
			Amount:             new(big.Int).Set(params.Amount),
			Nonce:              nonce,
			Expiry:             new(big.Int).Set(params.Expiry),
			TreeID:             new(big.Int).Set(params.TreeID),
			ChainID:            new(big.Int).Set(params.ChainID),
		},
		Data:          data,
		ApproveAmount: new(big.Int).Set(params.Amount),
	}, nil
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
