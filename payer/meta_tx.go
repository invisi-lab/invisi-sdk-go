package payer

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-labs/invisi-sdk-go/core"
)

// EIP-712 type definitions matching the InvisiProtocol contract.
var (
	payOpenWithSigTypes = apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"PayOpenWithSig": {
			{Name: "identityCommitment", Type: "bytes32"},
			{Name: "amount", Type: "uint256"},
			{Name: "nonce", Type: "bytes32"},
			{Name: "treeId", Type: "uint256"},
			{Name: "refundLink", Type: "bytes32"},
			{Name: "sigNonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
		},
	}

	payBlindWithSigTypes = apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"PayBlindWithSig": {
			{Name: "proof", Type: "bytes"},
			{Name: "paymentCommitment", Type: "bytes32"},
			{Name: "changeCommitment", Type: "bytes32"},
			{Name: "inputAmount", Type: "uint256"},
			{Name: "treeId", Type: "uint256"},
			{Name: "refundLink", Type: "bytes32"},
			{Name: "sigNonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
		},
	}

	depositCreditWithSigTypes = apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"DepositCreditWithSig": {
			{Name: "identityCommitment", Type: "bytes32"},
			{Name: "payerCommitment", Type: "bytes32"},
			{Name: "amount", Type: "uint256"},
			{Name: "nonce", Type: "bytes32"},
			{Name: "expiry", Type: "uint256"},
			{Name: "nullifierKeyHash", Type: "bytes32"},
			{Name: "treeId", Type: "uint256"},
			{Name: "sigNonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
		},
	}
)

// GetEIP712Domain returns the EIP-712 domain for the InvisibleProtocol contract.
func GetEIP712Domain(params EIP712DomainParams) apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:              "InvisibleProtocol",
		Version:           "1",
		ChainId:           (*math.HexOrDecimal256)(params.ChainID),
		VerifyingContract: params.VerifyingContract.Hex(),
	}
}

// BuildSignPayOpenTypedData builds EIP-712 typed data for payOpenWithSig.
// Computes paymentCommitment and generates a random nonce internally.
func BuildSignPayOpenTypedData(params SignPayOpenParams) (*SignPayOpenResult, error) {
	if params.ChainID == nil || params.Amount == nil || params.TreeID == nil {
		return nil, fmt.Errorf("payer: ChainID, Amount, TreeID are required")
	}
	if params.SigNonce == nil || params.Deadline == nil {
		return nil, fmt.Errorf("payer: SigNonce and Deadline are required")
	}

	// Generate nonce
	nonce, err := core.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("payer: failed to generate nonce: %w", err)
	}
	nonceHash := common.HexToHash(core.ElementToHex(nonce))

	// Compute paymentCommitment
	chainElem := core.BigIntToElement(params.ChainID)
	icElem, err := core.HexToElement(params.IdentityCommitment.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid identityCommitment: %w", err)
	}
	tokenHash := addressToTokenElement(params.TokenAddress)
	tokenElem, err := core.HexToElement(tokenHash.Hex())
	if err != nil {
		return nil, fmt.Errorf("payer: invalid tokenAddress: %w", err)
	}
	amountElem := core.BigIntToElement(params.Amount)

	commitment, err := core.ComputePaymentCommitment(chainElem, icElem, tokenElem, amountElem, nonce)
	if err != nil {
		return nil, fmt.Errorf("payer: failed to compute paymentCommitment: %w", err)
	}
	commitmentHash := common.HexToHash(core.ElementToHex(commitment))

	// RefundLink: default to zero
	refundLink := common.Hash{}
	if params.RefundLink != nil {
		refundLink = *params.RefundLink
	}

	domain := GetEIP712Domain(EIP712DomainParams{
		ChainID:           params.ChainID,
		VerifyingContract: params.VerifyingContract,
	})

	typedData := apitypes.TypedData{
		Types:       payOpenWithSigTypes,
		PrimaryType: "PayOpenWithSig",
		Domain:      domain,
		Message: apitypes.TypedDataMessage{
			"identityCommitment": params.IdentityCommitment.Hex(),
			"amount":             params.Amount.String(),
			"nonce":              nonceHash.Hex(),
			"treeId":             params.TreeID.String(),
			"refundLink":         refundLink.Hex(),
			"sigNonce":           params.SigNonce.String(),
			"deadline":           params.Deadline.String(),
		},
	}

	return &SignPayOpenResult{
		PaymentCommitment: commitmentHash,
		Nonce:             nonceHash,
		TypedData:         typedData,
	}, nil
}

// BuildSignPayBlindTypedData builds EIP-712 typed data for payBlindWithSig.
func BuildSignPayBlindTypedData(params SignPayBlindParams) (*SignPayBlindResult, error) {
	if params.ChainID == nil || params.InputAmount == nil || params.TreeID == nil {
		return nil, fmt.Errorf("payer: ChainID, InputAmount, TreeID are required")
	}
	if params.SigNonce == nil || params.Deadline == nil {
		return nil, fmt.Errorf("payer: SigNonce and Deadline are required")
	}

	refundLink := common.Hash{}
	if params.RefundLink != nil {
		refundLink = *params.RefundLink
	}

	domain := GetEIP712Domain(EIP712DomainParams{
		ChainID:           params.ChainID,
		VerifyingContract: params.VerifyingContract,
	})

	typedData := apitypes.TypedData{
		Types:       payBlindWithSigTypes,
		PrimaryType: "PayBlindWithSig",
		Domain:      domain,
		Message: apitypes.TypedDataMessage{
			"proof":             common.Bytes2Hex(params.Proof),
			"paymentCommitment": params.PaymentCommitment.Hex(),
			"changeCommitment":  params.ChangeCommitment.Hex(),
			"inputAmount":       params.InputAmount.String(),
			"treeId":            params.TreeID.String(),
			"refundLink":        refundLink.Hex(),
			"sigNonce":          params.SigNonce.String(),
			"deadline":          params.Deadline.String(),
		},
	}

	return &SignPayBlindResult{TypedData: typedData}, nil
}

// BuildSignDepositCreditTypedData builds EIP-712 typed data for depositCreditWithSig.
func BuildSignDepositCreditTypedData(params SignDepositCreditParams) (*SignDepositCreditResult, error) {
	if params.ChainID == nil || params.Amount == nil || params.TreeID == nil {
		return nil, fmt.Errorf("payer: ChainID, Amount, TreeID are required")
	}
	if params.Expiry == nil || params.SigNonce == nil || params.Deadline == nil {
		return nil, fmt.Errorf("payer: Expiry, SigNonce, Deadline are required")
	}

	domain := GetEIP712Domain(EIP712DomainParams{
		ChainID:           params.ChainID,
		VerifyingContract: params.VerifyingContract,
	})

	typedData := apitypes.TypedData{
		Types:       depositCreditWithSigTypes,
		PrimaryType: "DepositCreditWithSig",
		Domain:      domain,
		Message: apitypes.TypedDataMessage{
			"identityCommitment": params.IdentityCommitment.Hex(),
			"payerCommitment":    params.PayerCommitment.Hex(),
			"amount":             params.Amount.String(),
			"nonce":              params.Nonce.Hex(),
			"expiry":             params.Expiry.String(),
			"nullifierKeyHash":   params.NullifierKeyHash.Hex(),
			"treeId":             params.TreeID.String(),
			"sigNonce":           params.SigNonce.String(),
			"deadline":           params.Deadline.String(),
		},
	}

	return &SignDepositCreditResult{TypedData: typedData}, nil
}
