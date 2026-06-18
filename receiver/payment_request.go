package receiver

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// CreatePaymentRequest generates an invoice internally, builds an IPPRequest
// envelope, and optionally pre-computes a nonce + paymentCommitment for
// mode 1a (single-payer with known amount).
//
// Three modes:
//   - Mode 1a: Amount set, MultiPayer=false → nonce + paymentCommitment included
//   - Mode 1b: Amount set, MultiPayer=true → no nonce (assigned by propose)
//   - Mode 2:  No Amount → interactive mode, endpoint required
func CreatePaymentRequest(params CreatePaymentRequestParams) (*ipp.IPPRequest, error) {
	// Derive keys
	nullifierKey, err := core.DeriveNullifierKey(params.MasterKey, params.InvoiceID)
	if err != nil {
		return nil, err
	}

	recvAddrElem := addressToElement(params.ReceivingAddress)
	identityCommitment, err := core.ComputeIdentityCommitment(nullifierKey, recvAddrElem)
	if err != nil {
		return nil, err
	}

	icHex := core.ElementToHex(identityCommitment)

	request := &ipp.IPPRequest{
		V:       ipp.IPPCurrentVersion,
		Network: ipp.ChainIDToCAIP2(params.ChainID),
		Type:    ipp.MessageTypeRequest,
		Payload: ipp.IPPPayloadV1{
			IdentityCommitment: icHex,
			TokenAddress:       params.TokenAddress.Hex(),
			Operator:           params.Operator.Hex(),
			TreeID:             params.TreeID.String(),
			PaymentTypes:       params.PaymentTypes,
		},
	}

	if params.Amount != nil {
		amountStr := params.Amount.String()
		request.Payload.Amount = &amountStr
	}

	if params.Endpoint != nil {
		endpoint := *params.Endpoint
		request.Payload.Endpoint = &endpoint
	}

	if params.Meta != nil {
		request.Meta = params.Meta
	}

	// Mode 1a: amount set, not multi-payer → pre-compute nonce + paymentCommitment
	if params.Amount != nil && !params.MultiPayer {
		proposeResult, err := ComputePaymentCommitmentForPropose(ProposeParams{
			IdentityCommitment: common.HexToHash(icHex),
			TokenAddress:       params.TokenAddress,
			Amount:             params.Amount,
			ChainID:            params.ChainID,
		})
		if err != nil {
			return nil, err
		}

		nonceHex := proposeResult.Nonce.Hex()
		pcHex := proposeResult.PaymentCommitment.Hex()
		request.Payload.Nonce = &nonceHex
		request.Payload.PaymentCommitment = &pcHex
	}

	return request, nil
}
