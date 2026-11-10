package settlement

import "github.com/ethereum/go-ethereum/common"

// Operation contains public chain facts only. Empty secret amounts remain absent.
type Operation struct {
	ChainID             string           `json:"chainId"`
	ProtocolAddress     common.Address   `json:"protocolAddress"`
	BlockNumber         string           `json:"blockNumber"`
	BlockHash           common.Hash      `json:"blockHash"`
	ParentHash          common.Hash      `json:"parentHash"`
	Timestamp           string           `json:"timestamp"`
	TxHash              common.Hash      `json:"txHash"`
	OperationIndex      string           `json:"operationIndex"`
	MemberIndex         *string          `json:"memberIndex,omitempty"`
	LogIndex            string           `json:"logIndex"`
	Kind                string           `json:"kind"`
	TokenAddress        common.Address   `json:"tokenAddress"`
	MainCircuitCapacity string           `json:"mainCircuitCapacity"`
	PublicInputs        []common.Hash    `json:"publicInputs"`
	NoteNullifiers      []common.Hash    `json:"noteNullifiers"`
	CreditNullifiers    []common.Hash    `json:"creditNullifiers"`
	NoteOutputs         []Output         `json:"noteOutputs"`
	CreditOutputs       []CreditOutput   `json:"creditOutputs"`
	GrantRefs           []GrantRef       `json:"grantRefs"`
	Funding             []FundingReceipt `json:"funding"`
	WithdrawalAmount    string           `json:"withdrawalAmount"`
	FeeAmount           string           `json:"feeAmount"`
	FeeRecipient        common.Address   `json:"feeRecipient"`
	Success             bool             `json:"success"`
	OperationDigest     common.Hash      `json:"operationDigest"`
	UnsupportedReason   string           `json:"unsupportedReason,omitempty"`
}
type Output struct {
	NoteCommitment common.Hash `json:"noteCommitment"`
	TreeID         string      `json:"treeId"`
	LeafIndex      string      `json:"leafIndex"`
	Amount         string      `json:"amount,omitempty"`
}
type GrantRef struct {
	PayerIndex string `json:"payerIndex"`
	Shard      string `json:"shard"`
	Slot       string `json:"slot"`
}
type FundingReceipt struct {
	Payer        common.Address `json:"payer"`
	TokenAddress common.Address `json:"tokenAddress"`
	Amount       string         `json:"amount"`
}

type CreditOutput struct {
	CreditCommitment common.Hash `json:"creditCommitment"`
	TreeID           string      `json:"treeId"`
	LeafIndex        string      `json:"leafIndex"`
}
