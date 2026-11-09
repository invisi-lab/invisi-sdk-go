package protocol

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

type VerifyNoteCommitmentParams struct {
	ChainID               *big.Int
	OwnerAddress          common.Hash
	TokenAddress          common.Address
	Amount                *big.Int
	Nonce, NoteCommitment common.Hash
}

func VerifyNoteCommitment(p VerifyNoteCommitmentParams) (bool, error) {
	if err := CheckUint(p.ChainID, 64, "chainID"); err != nil {
		return false, err
	}
	if err := CheckUint(p.Amount, 128, "amount"); err != nil {
		return false, err
	}
	for _, value := range [][32]byte{p.OwnerAddress, p.Nonce, p.NoteCommitment} {
		if err := Field(value); err != nil {
			return false, err
		}
	}
	identity, err := core.ComputeNoteIdentity(core.BigIntToElement(p.OwnerAddress.Big()), core.BigIntToElement(p.Nonce.Big()))
	if err != nil {
		return false, err
	}
	expected, err := core.ComputeNoteCommitment(core.BigIntToElement(p.ChainID), identity, core.BigIntToElement(p.TokenAddress.Big()), core.BigIntToElement(p.Amount))
	if err != nil {
		return false, err
	}
	return expected.Bytes() == [32]byte(p.NoteCommitment), nil
}
