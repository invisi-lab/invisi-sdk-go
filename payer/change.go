package payer

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// ErrPayerMasterKeyRequired is returned by DeriveChange when PayerMasterKey is the
// zero element. A throwaway key would make the change unrecoverable, so payBlind
// change derivation rejects a zero master key instead of substituting one.
var ErrPayerMasterKeyRequired = errors.New("payer: PayerMasterKey is required for payBlind change derivation (no random fallback)")

// DeriveChangeParams configures deterministic change derivation for a shielded payment.
type DeriveChangeParams struct {
	PayerMasterKey         fr.Element     // REQUIRED (non-zero); no random fallback
	PaymentCommitment      common.Hash    // this payment's PC — the derivation salt
	ChangeReceivingAddress common.Address // payer's own wallet address (= PayBlind event payer)
	TokenAddress           common.Address
	ChangeAmount           *big.Int   // inputAmount - paymentAmount
	Nonce                  fr.Element // payment nonce, reused verbatim as changeNonce
	ChainID                *big.Int
}

// ChangeDerivation is the deterministic result: the changeCommitment fed into the ZK
// proof + BuildPayBlindCalldata, and the changeNullifierKey the caller stores locally
// for a later standard withdraw (settle_change).
type ChangeDerivation struct {
	ChangeNullifierKey fr.Element
	ChangeCommitment   common.Hash
}

// DeriveChange deterministically derives a shielded payment's change.
// Salt = this payment's paymentCommitment (necessarily present, already re-verified);
// changeNonce reuses the payment nonce. No randomness, no network. treeId / leafIndex
// are not produced here: they are known only from the propose response / on-chain
// PayBlind event.
func DeriveChange(params DeriveChangeParams) (*ChangeDerivation, error) {
	if params.PayerMasterKey.IsZero() {
		return nil, ErrPayerMasterKeyRequired
	}
	if params.ChangeAmount == nil {
		return nil, fmt.Errorf("payer: ChangeAmount is required")
	}
	if params.ChainID == nil {
		return nil, fmt.Errorf("payer: ChainID is required")
	}

	pcElem, err := core.HexToElement(params.PaymentCommitment.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid paymentCommitment: %w", err)
	}
	changeNullifierKey, err := core.DeriveChangeNullifierKey(params.PayerMasterKey, pcElem)
	if err != nil {
		return nil, fmt.Errorf("payer: derive change nullifier key: %w", err)
	}

	addrElem := core.BigIntToElement(new(big.Int).SetBytes(params.ChangeReceivingAddress.Bytes()))
	tokenElem := core.BigIntToElement(new(big.Int).SetBytes(params.TokenAddress.Bytes()))
	changeCommitment, err := core.ComputeChangeCommitment(
		core.BigIntToElement(params.ChainID),
		changeNullifierKey,
		addrElem,
		tokenElem,
		core.BigIntToElement(params.ChangeAmount),
		params.Nonce, // changeNonce = payment nonce
	)
	if err != nil {
		return nil, fmt.Errorf("payer: compute change commitment: %w", err)
	}

	return &ChangeDerivation{
		ChangeNullifierKey: changeNullifierKey,
		ChangeCommitment:   common.HexToHash(core.ElementToHex(changeCommitment)),
	}, nil
}
