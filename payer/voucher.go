package payer

import (
	"errors"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// BuildVoucherParams holds all parameters needed to assemble a payment voucher.
type BuildVoucherParams struct {
	Network           string
	ReceivingAddress  string
	PaymentCommitment string
	TokenAddress      string
	Operator          string
	Amount            *big.Int
	Nonce             string
	NullifierKey      string
	TreeID            string
	LeafIndex         string            // optional, "" = omit
	Meta              map[string]string // optional
}

// BuildVoucher assembles a payment voucher (mode 3 address transfer) from values
// the payer already holds after payOpen. Pure local assembly — no chain/endpoint call.
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
			ReceivingAddress:  params.ReceivingAddress,
			PaymentCommitment: params.PaymentCommitment,
			TokenAddress:      params.TokenAddress,
			Operator:          params.Operator,
			Amount:            params.Amount.String(),
			Nonce:             params.Nonce,
			NullifierKey:      params.NullifierKey,
			TreeID:            params.TreeID,
			LeafIndex:         params.LeafIndex,
		},
	}, nil
}
