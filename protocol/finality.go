package protocol

import "github.com/ethereum/go-ethereum/common"

type FinalityTag string

const (
	FinalityLatest    FinalityTag = "latest"
	FinalitySafe      FinalityTag = "safe"
	FinalityFinalized FinalityTag = "finalized"
)

func (t FinalityTag) Valid() bool {
	return t == FinalityLatest || t == FinalitySafe || t == FinalityFinalized
}

type FinalityStatus string

const (
	FinalityAvailable    FinalityStatus = "available"
	FinalityUnsupported  FinalityStatus = "unsupported"
	FinalityUnavailable  FinalityStatus = "unavailable"
	FinalityInconsistent FinalityStatus = "inconsistent"
)

func (s FinalityStatus) Valid() bool {
	return s == FinalityAvailable || s == FinalityUnsupported || s == FinalityUnavailable || s == FinalityInconsistent
}

type CanonicalBlock struct {
	Number     uint64      `json:"number"`
	Hash       common.Hash `json:"hash"`
	ParentHash common.Hash `json:"parentHash"`
	Timestamp  uint64      `json:"timestamp"`
}
type FinalityHead struct {
	Status FinalityStatus  `json:"status"`
	Block  *CanonicalBlock `json:"block"`
}
type ChainFinality struct {
	Latest    FinalityHead `json:"latest"`
	Safe      FinalityHead `json:"safe"`
	Finalized FinalityHead `json:"finalized"`
}

func (f ChainFinality) Head(tag FinalityTag) FinalityHead {
	switch tag {
	case FinalityLatest:
		return f.Latest
	case FinalitySafe:
		return f.Safe
	case FinalityFinalized:
		return f.Finalized
	}
	return FinalityHead{Status: FinalityUnavailable}
}

type TransactionConfirmation struct {
	Level  string          `json:"level"`
	Status FinalityStatus  `json:"status"`
	Block  *CanonicalBlock `json:"block"`
}
