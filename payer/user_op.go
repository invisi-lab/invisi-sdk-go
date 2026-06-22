package payer

import (
	"fmt"
)

// BuildPayOpenUserOp builds an EIP-4337 UserOperation for payOpen.
// Internally calls BuildPayOpenCalldata and wraps the result.
func BuildPayOpenUserOp(params PayOpenUserOpParams) (*PayOpenUserOp, error) {
	if params.ChainID == nil || params.Amount == nil || params.TreeID == nil {
		return nil, fmt.Errorf("payer: ChainID, Amount, TreeID are required")
	}
	if params.Nonce == nil {
		return nil, fmt.Errorf("payer: Nonce is required")
	}

	calldataResult, err := BuildPayOpenCalldata(PayOpenParams{
		ChainID:            params.ChainID,
		IdentityCommitment: params.IdentityCommitment,
		TokenAddress:       params.TokenAddress,
		Amount:             params.Amount,
		TreeID:             params.TreeID,
		RefundLink:         params.RefundLink,
		Nonce:              params.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("payer: BuildPayOpenUserOp: %w", err)
	}

	return &PayOpenUserOp{
		BaseUserOp: BaseUserOp{
			To:                   params.ProtocolAddress,
			Sender:               params.Sender,
			CallData:             calldataResult.Data,
			CallGasLimit:         GasEstimatePayOpen.CallGasLimit,
			VerificationGasLimit: GasEstimatePayOpen.VerificationGasLimit,
			PreVerificationGas:   GasEstimatePayOpen.PreVerificationGas,
			Paymaster:            params.Paymaster,
		},
		PaymentCommitment: calldataResult.PaymentCommitment,
		Nonce:             calldataResult.Nonce,
	}, nil
}

// BuildPayBlindUserOp builds an EIP-4337 UserOperation for payBlind.
// Internally calls BuildPayBlindCalldata and wraps the result.
func BuildPayBlindUserOp(params PayBlindUserOpParams) (*BaseUserOp, error) {
	if params.InputAmount == nil || params.TreeID == nil {
		return nil, fmt.Errorf("payer: InputAmount, TreeID are required")
	}

	calldataResult, err := BuildPayBlindCalldata(PayBlindParams{
		Proof:             params.Proof,
		PaymentCommitment: params.PaymentCommitment,
		ChangeCommitment:  params.ChangeCommitment,
		InputAmount:       params.InputAmount,
		TreeID:            params.TreeID,
		RefundLink:        params.RefundLink,
	})
	if err != nil {
		return nil, fmt.Errorf("payer: BuildPayBlindUserOp: %w", err)
	}

	return &BaseUserOp{
		To:                   params.ProtocolAddress,
		Sender:               params.Sender,
		CallData:             calldataResult.Data,
		CallGasLimit:         GasEstimatePayBlind.CallGasLimit,
		VerificationGasLimit: GasEstimatePayBlind.VerificationGasLimit,
		PreVerificationGas:   GasEstimatePayBlind.PreVerificationGas,
		Paymaster:            params.Paymaster,
	}, nil
}

// BuildDepositCreditUserOp builds an EIP-4337 UserOperation for depositCredit.
// Directly encodes calldata (no cryptographic computation needed).
func BuildDepositCreditUserOp(params DepositCreditUserOpParams) (*BaseUserOp, error) {
	if params.Amount == nil || params.TreeID == nil {
		return nil, fmt.Errorf("payer: Amount, TreeID are required")
	}
	if params.Expiry == nil {
		return nil, fmt.Errorf("payer: Expiry is required")
	}

	// ABI encode depositCredit directly
	data, err := packCalldata(
		invisiProtocolABI,
		"depositCredit",
		params.IdentityCommitment, // bytes32
		params.PayerCommitment,    // bytes32
		params.Amount,             // uint256
		params.Nonce,              // bytes32
		params.Expiry,             // uint256
		params.NullifierKeyHash,   // bytes32
		params.TreeID,             // uint256
	)
	if err != nil {
		return nil, fmt.Errorf("payer: BuildDepositCreditUserOp: %w", err)
	}

	return &BaseUserOp{
		To:                   params.ProtocolAddress,
		Sender:               params.Sender,
		CallData:             data,
		CallGasLimit:         GasEstimateDepositCredit.CallGasLimit,
		VerificationGasLimit: GasEstimateDepositCredit.VerificationGasLimit,
		PreVerificationGas:   GasEstimateDepositCredit.PreVerificationGas,
		Paymaster:            params.Paymaster,
	}, nil
}
