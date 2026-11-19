package protocol

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

// NoteOperationID correlates canonical receipt facts independently of proofs,
// signatures, permits and funding nonces. Shield uses empty sources/nullifiers
// and zero withdrawal fields in p.
func NoteOperationID(c TransactionContext, kind uint8, p CrossTreeTransactParams, deliveries []Delivery, funding FundingParams) (common.Hash, error) {
	if kind > 1 || c.ChainID == nil || c.ChainID.Sign() < 0 || c.ChainID.BitLen() > 256 {
		return common.Hash{}, fmt.Errorf("invalid Note operation context or kind")
	}
	var err error
	p.FeeAmount, err = normalizeFee(p.FeeRecipient, p.FeeAmount)
	if err != nil {
		return common.Hash{}, err
	}
	if p.WithdrawalAmount == nil {
		p.WithdrawalAmount = new(big.Int)
	}
	if funding.Amount == nil {
		funding.Amount = new(big.Int)
	}
	dh, err := DeliveriesHash(deliveries)
	if err != nil {
		return common.Hash{}, err
	}
	method, err := transactMethod(false)
	if err != nil {
		return common.Hash{}, err
	}
	args := abi.Arguments{{Type: mustType("bytes32")}, {Type: mustType("uint256")}, {Type: mustType("address")}, {Type: mustType("uint8")}, {Type: *method.Inputs[1].Type.TupleElems[0]}, {Type: mustType("address")}, {Type: mustType("uint256")}, {Type: mustType("bytes32")}}
	return HashABI(args, crypto.Keccak256Hash([]byte("Invisi.NoteOperation.v1")), c.ChainID, c.ProtocolAddress, kind, p, funding.Payer, funding.Amount, dh)
}
