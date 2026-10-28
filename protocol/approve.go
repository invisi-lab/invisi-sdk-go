package protocol

import ()

// BuildApproveCalldata encodes an ERC20 approve(spender, amount) call.
// Returns the token contract address as `To` and the encoded calldata.
func BuildApproveCalldata(params ApproveParams) (*ApproveCalldata, error) {
	if err := CheckUint(params.Amount, 256, "amount"); err != nil {
		return nil, err
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
