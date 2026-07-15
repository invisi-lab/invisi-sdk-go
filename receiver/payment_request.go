package receiver

import (
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
)

// CreatePaymentRequest generates an invoice internally and builds an IPPRequest
// envelope for the given mode.
//
// Modes:
//   - "single" (1a): fixed amount, single-use — nonce/paymentCommitment are
//     delivered later via propose, not prefilled here.
//   - "multi" (1b/2): many payers or many uses; Amount may be set (1b) or
//     omitted (2, interactive — Endpoint required).
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
			PaymentTypes:       params.PaymentTypes,
			Mode:               params.Mode,
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

	return request, nil
}
