package wallet

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
)

// DeliveryDetectionItem contains only public delivery data; native code retains detection keys.
type DeliveryDetectionItem struct {
	ID                            string
	Owner                         common.Address
	ChainID                       uint64
	ProtocolAddress, TokenAddress common.Address
	OutputCommitment              common.Hash
	Role                          string
	Clue, Ciphertext              []byte
}
type DeliveryDetectionResult struct {
	Owner   common.Address
	Results []DeliveryDetectionDecision
}
type DeliveryDetectionDecision struct{ ID, Status string }
type BatchDeliveryDetector interface {
	Detect(context.Context, []DeliveryDetectionItem) (DeliveryDetectionResult, error)
}
