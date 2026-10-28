package payer

import (
	"errors"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// BuildVoucherParams holds all parameters needed to assemble a payment voucher.
type BuildVoucherParams struct {
	Network          string
	ReceivingAddress string
	NoteCommitment   string
	TokenAddress     string
	Operator         string
	Amount           *big.Int
	Nonce            string
	NoteSecret       string
	TreeID           string
	LeafIndex        string // optional, "" = omit
	// Endpoint is the server address that accepts settlement for this voucher;
	// required (see ipp.IPPVoucherV1.Endpoint).
	Endpoint string
	Meta     map[string]string // optional
}

// BuildVoucher assembles a payment voucher (mode 3 address transfer) from values
// the payer already holds after confirmed commitment insertion. Pure local assembly — no chain/endpoint call.
func BuildVoucher(params BuildVoucherParams) (ipp.IPPVoucher, error) {
	if params.Amount == nil {
		return ipp.IPPVoucher{}, errors.New("amount is nil")
	}
	return ipp.IPPVoucher{
		V:       1,
		Network: params.Network,
		Type:    ipp.MessageTypeVoucher,
		Meta:    params.Meta,
		Payload: ipp.IPPVoucherV1{
			ReceivingAddress: params.ReceivingAddress,
			NoteCommitment:   params.NoteCommitment,
			TokenAddress:     params.TokenAddress,
			Operator:         params.Operator,
			Amount:           params.Amount.String(),
			Nonce:            params.Nonce,
			NoteSecret:       params.NoteSecret,
			TreeID:           params.TreeID,
			LeafIndex:        params.LeafIndex,
			Endpoint:         params.Endpoint,
		},
	}, nil
}
