package receiver

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
	"github.com/invisi-lab/invisi-sdk-go/payer"
	"math/big"
)

type GenerateInvoiceParams struct {
	MasterKey, InvoiceID                     fr.Element
	ReceivingAddress, TokenAddress, Operator common.Address
}

// NoteSecret is private invoice material and must not be included in public service queries.
type Invoice struct {
	InvoiceID, NoteSecret                    fr.Element
	IdentityCommitment                       common.Hash
	ReceivingAddress, TokenAddress, Operator common.Address
}
type CreatePaymentRequestParams struct {
	MasterKey, InvoiceID                     fr.Element
	ReceivingAddress, TokenAddress, Operator common.Address
	ChainID                                  *big.Int
	PaymentTypes                             []ipp.PaymentType
	Mode                                     string
	Amount                                   *big.Int
	Endpoint                                 *string
	Meta                                     map[string]string
}
type ProposeParams struct {
	IdentityCommitment common.Hash
	TokenAddress       common.Address
	Amount, ChainID    *big.Int
}
type ProposeResult struct{ Nonce, NoteCommitment common.Hash }
type BuildUnsignedWithdrawalCalldataParams struct {
	Context     payer.TransactionContext
	Proof       []byte
	Params      payer.CrossTreeTransactParams
	Funding     payer.FundingParams
	Caller      *common.Address
	SourceNotes []core.Note
}
type BuildCreditDrawCalldataParams struct {
	Context     payer.TransactionContext
	Proof       []byte
	Params      payer.CreditDrawParams
	Ciphertexts [][]byte
}
type BuildCancelCreditCalldataParams struct {
	Context       payer.TransactionContext
	Proof         []byte
	Params        payer.CancelCreditParams
	Authorization payer.CreditAuthorization
}
type BuildTopUpCreditCalldataParams struct {
	Context       payer.TransactionContext
	Proof         []byte
	Params        payer.TopUpCreditParams
	Ciphertexts   [][]byte
	Funding       payer.FundingParams
	Authorization payer.CreditAuthorization
	Caller        *common.Address
}
