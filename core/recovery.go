package core

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func checkUint(value fr.Element, bits int, name string) error {
	if ElementToBigInt(value).BitLen() > bits {
		return fmt.Errorf("%s out of uint%d range", name, bits)
	}
	return nil
}
func validateProtocolContext(context ProtocolContext, tokenAddress fr.Element) error {
	for _, v := range []struct {
		value fr.Element
		bits  int
		name  string
	}{{context.ChainID, 64, "chainId"}, {context.ProtocolAddress, 160, "protocolAddress"}, {tokenAddress, 160, "tokenAddress"}} {
		if err := checkUint(v.value, v.bits, v.name); err != nil {
			return err
		}
	}
	return nil
}
func validateCredit(note CreditNote) error {
	if err := checkUint(note.Amount, 126, "amount"); err != nil {
		return err
	}
	if err := checkUint(note.Expiry, 64, "expiry"); err != nil {
		return err
	}
	if err := checkUint(note.TokenAddress, 160, "tokenAddress"); err != nil {
		return err
	}
	if note.Amount.IsZero() || note.CreditSharedKey.IsZero() {
		return fmt.Errorf("amount and creditSharedKey must be nonzero")
	}
	return nil
}

// RecoverCreditRemainder requires a previously authenticated note and authenticated
// draw event. It verifies the consumed nullifier, recovered amount and commitment.
// A full draw returns nil only when both remainder fields are zero.
func RecoverCreditRemainder(context ProtocolContext, previous CreditNote, recovery CreditRemainderRecovery) (*CreditNote, error) {
	if err := validateProtocolContext(context, previous.TokenAddress); err != nil {
		return nil, err
	}
	if err := validateCredit(previous); err != nil {
		return nil, err
	}
	expected, err := ComputeCreditNullifier(context.ChainID, previous.CreditSharedKey, previous.TokenAddress, previous.Amount, previous.Nonce, previous.Expiry)
	if err != nil {
		return nil, err
	}
	if recovery.CreditNullifier.IsZero() || expected != recovery.CreditNullifier {
		return nil, fmt.Errorf("previous credit nullifier mismatch")
	}
	if recovery.RemainderCommitment.IsZero() {
		if !recovery.RecoveryAmount.IsZero() {
			return nil, fmt.Errorf("full draw requires zero recovery amount")
		}
		return nil, nil
	}
	nonce, err := DeriveRemainderNonce(context.ChainID, context.ProtocolAddress, previous.TokenAddress, recovery.CreditNullifier, previous.CreditSharedKey)
	if err != nil {
		return nil, err
	}
	pad, err := ComputeCreditRecoveryPad(context.ChainID, context.ProtocolAddress, previous.TokenAddress, recovery.CreditNullifier, previous.CreditSharedKey)
	if err != nil {
		return nil, err
	}
	var amount fr.Element
	amount.Sub(&recovery.RecoveryAmount, &pad)
	if amount.IsZero() || amount.Cmp(&previous.Amount) >= 0 || ElementToBigInt(amount).BitLen() > 126 {
		return nil, fmt.Errorf("invalid recovered remainder amount")
	}
	if nonce == previous.Nonce {
		return nil, fmt.Errorf("remainder nonce must change")
	}
	keyHash, err := ComputeCreditSharedKeyHash(previous.CreditSharedKey)
	if err != nil {
		return nil, err
	}
	commitment, err := ComputeCreditCommitment(context.ChainID, previous.SpenderIdentityCommitment, previous.OwnerIdentityCommitment, previous.TokenAddress, amount, nonce, previous.Expiry, keyHash)
	if err != nil {
		return nil, err
	}
	if commitment != recovery.RemainderCommitment {
		return nil, fmt.Errorf("recovered credit commitment mismatch")
	}
	result := previous
	result.Amount = amount
	result.Nonce = nonce
	return &result, nil
}
