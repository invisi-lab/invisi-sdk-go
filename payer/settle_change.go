package payer

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// BuildSettleChangeCalldata computes the payment nullifier and encodes
// withdraw(proof, merkleRoot, [nullifier], amount, receivingAddress, treeId) calldata.
// This settles a single change note from a prior shielded payment.
func BuildSettleChangeCalldata(params SettleChangeParams) (*SettleChangeCalldata, error) {
	if params.ChainID == nil {
		return nil, fmt.Errorf("payer: ChainID is required")
	}
	if params.Amount == nil {
		return nil, fmt.Errorf("payer: Amount is required")
	}
	if params.TreeID == nil {
		return nil, fmt.Errorf("payer: TreeID is required")
	}

	// Convert to field elements
	chainElem := core.BigIntToElement(params.ChainID)
	tokenHash := addressToTokenElement(params.TokenAddress)
	tokenElem, err := core.HexToElement(tokenHash.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid tokenAddress: %w", err)
	}
	amountElem := core.BigIntToElement(params.Amount)

	// Compute nullifier = Poseidon2(D_NULL, chainId, nullifierKey, tokenAddress, amount, nonce)
	nullifier, err := core.ComputePaymentNullifier(chainElem, params.NullifierKey, tokenElem, amountElem, params.Nonce)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to compute nullifier: %w", err)
	}

	nullifierHash := common.HexToHash(core.ElementToHex(nullifier))

	// ABI encode: withdraw(proof, merkleRoot, nullifiers[], totalAmount, receivingAddress, treeId)
	nullifiers := [][32]byte{nullifierHash}
	data, err := packCalldata(
		invisiProtocolABI,
		"withdraw",
		params.Proof,            // bytes
		params.MerkleRoot,       // bytes32
		nullifiers,              // bytes32[]
		params.Amount,           // uint256
		params.ReceivingAddress, // address
		params.TreeID,           // uint256
	)
	if err != nil {
		return nil, err
	}

	return &SettleChangeCalldata{
		Nullifier: nullifierHash,
		Data:      data,
	}, nil
}
