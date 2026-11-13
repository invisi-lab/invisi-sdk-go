package settlement

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// SubmissionID is the default server-side deduplication algorithm.
// Clients must use the ID returned by their Broadcaster, which may allocate IDs differently.
func SubmissionID(chainID *big.Int, protocol common.Address, data []byte) (common.Hash, error) {
	hash, err := RequestHash(data)
	if err != nil {
		return common.Hash{}, err
	}
	return SubmissionIDFromHash(chainID, protocol, hash)
}

// SubmissionIDFromHash implements the default server strategy for an already bound semantic hash.
func SubmissionIDFromHash(chainID *big.Int, protocol common.Address, hash common.Hash) (common.Hash, error) {
	if chainID == nil || chainID.Sign() <= 0 || chainID.BitLen() > 256 || protocol == (common.Address{}) {
		return common.Hash{}, errors.New("invalid submission scope")
	}
	stringType, _ := abi.NewType("string", "", nil)
	uintType, _ := abi.NewType("uint256", "", nil)
	addressType, _ := abi.NewType("address", "", nil)
	bytesType, _ := abi.NewType("bytes32", "", nil)
	encoded, err := (abi.Arguments{{Type: stringType}, {Type: uintType}, {Type: addressType}, {Type: bytesType}}).Pack("InvisiBroadcasterSubmissionV1", chainID, protocol, hash)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(encoded), nil
}
