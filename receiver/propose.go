package receiver

import (
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/payer"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// ComputeNoteCommitmentForPropose generates a nonce and computes the
// noteCommitment for a propose response.
func ComputeNoteCommitmentForPropose(params ProposeParams) (*ProposeResult, error) {
	if err := payer.CheckUint(params.ChainID, 64, "chainID"); err != nil {
		return nil, err
	}
	if err := payer.CheckUint(params.Amount, 128, "amount"); err != nil {
		return nil, err
	}
	if params.Amount.Sign() == 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	// Generate random nonce
	nonce, err := core.GenerateNonce()
	if err != nil {
		return nil, err
	}

	// Convert params to field elements
	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		return nil, err
	}
	tokenElem := core.BigIntToElement(new(big.Int).SetBytes(params.TokenAddress.Bytes()))
	amountElem := core.BigIntToElement(params.Amount)
	chainElem := core.BigIntToElement(params.ChainID)

	// noteCommitment = Poseidon2(DNote, chainId, identityCommitment, tokenAddress, amount, nonce)
	pc, err := core.ComputeNoteCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	if err != nil {
		return nil, err
	}

	nonceHash := common.HexToHash(core.ElementToHex(nonce))
	pcHash := common.HexToHash(core.ElementToHex(pc))

	return &ProposeResult{
		Nonce:          nonceHash,
		NoteCommitment: pcHash,
	}, nil
}
