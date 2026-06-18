package receiver

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// GenerateInvoice derives a nullifierKey from masterKey + invoiceId,
// computes the identityCommitment, and returns the full Invoice.
//
// The returned Invoice.NullifierKey is sensitive — it must never be
// transmitted, logged, or persisted in unencrypted storage.
func GenerateInvoice(params GenerateInvoiceParams) (*Invoice, error) {
	// Derive nullifierKey = Poseidon2(D_NKEY, masterKey, invoiceId)
	nullifierKey, err := core.DeriveNullifierKey(params.MasterKey, params.InvoiceID)
	if err != nil {
		return nil, err
	}

	// Convert receiving address to field element
	recvAddrBigInt := new(big.Int).SetBytes(params.ReceivingAddress.Bytes())
	recvAddrElem := core.BigIntToElement(recvAddrBigInt)

	// Compute identityCommitment = Poseidon2(D_IDENT, nullifierKey, receivingAddress)
	identityCommitment, err := core.ComputeIdentityCommitment(nullifierKey, recvAddrElem)
	if err != nil {
		return nil, err
	}

	icHash := common.HexToHash(core.ElementToHex(identityCommitment))

	return &Invoice{
		InvoiceID:          params.InvoiceID,
		NullifierKey:       nullifierKey,
		IdentityCommitment: icHash,
		ReceivingAddress:   params.ReceivingAddress,
		TokenAddress:       params.TokenAddress,
		Amount:             new(big.Int).Set(params.Amount),
		TreeID:             new(big.Int).Set(params.TreeID),
		Operator:           params.Operator,
	}, nil
}

// addressToElement converts a common.Address to an fr.Element.
// This is a package-internal helper used throughout receiver.
func addressToElement(addr common.Address) fr.Element {
	return core.BigIntToElement(new(big.Int).SetBytes(addr.Bytes()))
}
