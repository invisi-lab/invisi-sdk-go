package payer

import (
	"fmt"
)

// BuildApproveCalldata encodes an ERC20 approve(spender, amount) call.
// Returns the token contract address as `To` and the encoded calldata.
func BuildApproveCalldata(params ApproveParams) (*ApproveCalldata, error) {
	if params.Amount == nil {
		return nil, fmt.Errorf("payer: Amount is required")
	}

	data, err := packCalldata(
		erc20ABI,
		"approve",
		params.ProtocolAddress, // address spender
		params.Amount,          // uint256 amount
	)
	if err != nil {
		return nil, err
	}

	return &ApproveCalldata{
		To:       params.Token,
		CallData: data,
	}, nil
}
