package payer

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"
)

type VerifyNoteCommitmentParams struct {
	ChainID               *big.Int
	IdentityCommitment    common.Hash
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
	if err := Field(p.IdentityCommitment); err != nil {
		return false, err
	}
	if err := Field(p.Nonce); err != nil {
		return false, err
	}
	if err := Field(p.NoteCommitment); err != nil {
		return false, err
	}
	expected, err := core.ComputeNoteCommitment(core.BigIntToElement(p.ChainID), core.BigIntToElement(p.IdentityCommitment.Big()), core.BigIntToElement(p.TokenAddress.Big()), core.BigIntToElement(p.Amount), core.BigIntToElement(p.Nonce.Big()))
	return expected.Bytes() == [32]byte(p.NoteCommitment), err
}

type VerifyCreditCommitmentParams struct {
	ChainID, SpenderIdentityCommitment, OwnerIdentityCommitment, TokenAddress, Amount, Nonce, Expiry, CreditSharedKey, CreditCommitment fr.Element
}

func VerifyCreditCommitment(p VerifyCreditCommitmentParams) (bool, error) {
	if p.CreditSharedKey.IsZero() {
		return false, fmt.Errorf("creditSharedKey must be nonzero")
	}
	for _, check := range []struct {
		value fr.Element
		bits  int
		name  string
	}{{p.ChainID, 64, "chainID"}, {p.TokenAddress, 160, "token"}, {p.Amount, 126, "amount"}, {p.Expiry, 64, "expiry"}} {
		if err := CheckUint(check.value.BigInt(new(big.Int)), check.bits, check.name); err != nil {
			return false, err
		}
	}
	keyHash, err := core.ComputeCreditSharedKeyHash(p.CreditSharedKey)
	if err != nil {
		return false, err
	}
	expected, err := core.ComputeCreditCommitment(p.ChainID, p.SpenderIdentityCommitment, p.OwnerIdentityCommitment, p.TokenAddress, p.Amount, p.Nonce, p.Expiry, keyHash)
	return expected.Equal(&p.CreditCommitment), err
}
