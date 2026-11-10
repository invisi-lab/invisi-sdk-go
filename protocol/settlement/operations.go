package settlement

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"math/big"
)

type ShieldParams struct {
	Token           common.Address
	NoteCommitments [][32]byte
	Deadline        *big.Int
	FeeRecipient    common.Address
	FeeAmount       *big.Int
}
type ShieldOutput struct {
	Identity [32]byte
	Amount   *big.Int
}
type CreateCreditParams struct {
	Token            common.Address
	Owner            common.Address
	RevocationShard  uint64
	RevocationSlot   uint16
	CreditCommitment [32]byte
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}
type CreditPayer struct {
	Payer  common.Address
	Amount *big.Int
}
type CreditGrantRef struct {
	PayerIndex *big.Int
	Shard      uint64
	Slot       uint16
}
type CreditDrawParams struct {
	Token             common.Address
	Sources           []InputSource
	Payers            []CreditPayer
	GrantRefs         []CreditGrantRef
	Nullifiers        [][32]byte
	CreditCommitments [][32]byte
	RecoveryAmounts   [][32]byte
	NoteCommitment    [32]byte
	TotalAmount       *big.Int
	Deadline          *big.Int
	FeeRecipient      common.Address
	FeeAmount         *big.Int
}
type UpdateCreditParams struct {
	Token            common.Address
	Owner            common.Address
	Source           InputSource
	Nullifier        [32]byte
	CreditCommitment [32]byte
	RevocationShard  uint64
	RevocationSlot   uint16
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}
type RevokeCreditParams struct {
	Token        common.Address
	Owner        common.Address
	Shard        uint64
	Slot         uint16
	Deadline     *big.Int
	Signature    []byte
	FeeRecipient common.Address
	FeeAmount    *big.Int
}

// Credit fees are bound by owner/spender authorization; no separate public funding signature is used.
type Statement struct {
	Token        common.Address
	FeeRecipient common.Address
	FeeAmount    *big.Int
	Deadline     *big.Int
	Nullifiers   []common.Hash
}

func (r *Request) IsCredit() bool {
	return r.Method == "createCredit" || r.Method == "creditDraw" || r.Method == "updateCredit" || r.Method == "revokeCredit"
}
func (r *Request) Statement() Statement {
	s := Statement{FeeAmount: new(big.Int)}
	var nullifiers [][32]byte
	switch r.Method {
	case "transact":
		p := r.Params
		s = Statement{Token: p.Token, FeeRecipient: p.FeeRecipient, FeeAmount: p.FeeAmount, Deadline: p.Deadline}
		nullifiers = p.Nullifiers
	case "shield":
		p := r.Shield
		s = Statement{Token: p.Token, FeeRecipient: p.FeeRecipient, FeeAmount: p.FeeAmount, Deadline: p.Deadline}
	case "createCredit":
		s.FeeRecipient = r.Creation.FeeRecipient
		s.FeeAmount = r.Creation.FeeAmount
		s.Token = r.Creation.Token
		s.Deadline = r.Creation.Deadline
	case "creditDraw":
		s.FeeRecipient = r.Draw.FeeRecipient
		s.FeeAmount = r.Draw.FeeAmount
		s.Token = r.Draw.Token
		s.Deadline = r.Draw.Deadline
		nullifiers = r.Draw.Nullifiers
	case "updateCredit":
		s.FeeRecipient = r.Update.FeeRecipient
		s.FeeAmount = r.Update.FeeAmount
		s.Token = r.Update.Token
		s.Deadline = r.Update.Deadline
		nullifiers = [][32]byte{r.Update.Nullifier}
	case "revokeCredit":
		s.FeeRecipient = r.Revocation.FeeRecipient
		s.FeeAmount = r.Revocation.FeeAmount
		s.Token = r.Revocation.Token
		s.Deadline = r.Revocation.Deadline
	}
	for _, n := range nullifiers {
		s.Nullifiers = append(s.Nullifiers, common.Hash(n))
	}
	return s
}
func (r *Request) HasFunding() bool { return r.Method == "transact" || r.Method == "shield" }

// Grant reservations prevent duplicate pending create/revoke calls that have no nullifier or funding nonce.
func (r *Request) PublicReservations() []common.Hash {
	var owner common.Address
	var shard uint64
	var slot uint16
	switch r.Method {
	case "createCredit":
		owner = r.Creation.Owner
		shard = r.Creation.RevocationShard
		slot = r.Creation.RevocationSlot
	case "revokeCredit":
		owner = r.Revocation.Owner
		shard = r.Revocation.Shard
		slot = r.Revocation.Slot
	default:
		return nil
	}
	args := abi.Arguments{}
	for _, kind := range []string{"string", "string", "address", "address", "uint64", "uint16"} {
		typ, _ := abi.NewType(kind, "", nil)
		args = append(args, abi.Argument{Type: typ})
	}
	encoded, err := args.Pack("InvisiBroadcasterGrantReservationV1", r.Method, r.Statement().Token, owner, shard, slot)
	if err != nil {
		panic(err)
	}
	return []common.Hash{crypto.Keccak256Hash(encoded)}
}
