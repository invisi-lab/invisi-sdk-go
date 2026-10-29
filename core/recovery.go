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
	if err := ValidateMemo(note.Memo); err != nil {
		return err
	}
	for _, v := range []fr.Element{note.SpenderPayout, note.OwnerPayout} {
		if err := checkUint(v, 160, "Credit payout"); err != nil {
			return err
		}
	}
	if err := checkUint(note.Amount, 126, "amount"); err != nil {
		return err
	}
	if err := checkUint(note.Expiry, 64, "expiry"); err != nil {
		return err
	}
	if err := checkUint(note.TokenAddress, 160, "tokenAddress"); err != nil {
		return err
	}
	if note.SpenderWalletIdentity.IsZero() || note.OwnerWalletIdentity.IsZero() || note.SpenderPayout.IsZero() || note.OwnerPayout.IsZero() || note.Amount.IsZero() || note.Expiry.IsZero() || note.CreditSharedKey.IsZero() {
		return fmt.Errorf("Credit roles, payouts, amount, expiry and shared key must be nonzero")
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
	if previous.ProtocolContext != context {
		return nil, fmt.Errorf("previous Credit context mismatch")
	}
	keyHash, err := ComputeCreditSharedKeyHash(previous.CreditSharedKey)
	if err != nil {
		return nil, err
	}
	leaf, err := ComputeCreditCommitment(context.ChainID, previous.SpenderWalletIdentity, previous.OwnerWalletIdentity, previous.SpenderPayout, previous.OwnerPayout, previous.TokenAddress, previous.Amount, previous.Nonce, previous.Expiry, keyHash)
	if err != nil {
		return nil, err
	}
	expected, err := ComputeCreditNullifier(context.ChainID, previous.CreditSharedKey, leaf)
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
	keyHash, err = ComputeCreditSharedKeyHash(previous.CreditSharedKey)
	if err != nil {
		return nil, err
	}
	commitment, err := ComputeCreditCommitment(context.ChainID, previous.SpenderWalletIdentity, previous.OwnerWalletIdentity, previous.SpenderPayout, previous.OwnerPayout, previous.TokenAddress, amount, nonce, previous.Expiry, keyHash)
	if err != nil {
		return nil, err
	}
	if commitment != recovery.RemainderCommitment {
		return nil, fmt.Errorf("recovered credit commitment mismatch")
	}
	// The proof binds financial fields only; the child memo is unknown without its authenticated envelope.
	result := previous
	result.Memo = nil
	result.Amount = amount
	result.Nonce = nonce
	return &result, nil
}
