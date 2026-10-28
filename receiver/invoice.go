package receiver

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"math/big"
)

func GenerateInvoice(p GenerateInvoiceParams) (*Invoice, error) {
	secret, err := core.DeriveNoteSecret(p.MasterKey, p.InvoiceID)
	if err != nil {
		return nil, err
	}
	identity, err := core.ComputeIdentityCommitment(secret, addressToElement(p.ReceivingAddress))
	if err != nil {
		return nil, err
	}
	return &Invoice{InvoiceID: p.InvoiceID, NoteSecret: secret, IdentityCommitment: identity.Bytes(), ReceivingAddress: p.ReceivingAddress, TokenAddress: p.TokenAddress, Operator: p.Operator}, nil
}
func addressToElement(a common.Address) fr.Element { return core.BigIntToElement(a.Big()) }
func fieldBig(v fr.Element) *big.Int               { return v.BigInt(new(big.Int)) }
