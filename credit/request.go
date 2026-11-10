package credit

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

type Context struct {
	primitives.ProtocolContext
	TokenAddress common.Address
	Timestamp    uint64
}
type CreateParams struct {
	Token, Owner     common.Address
	RevocationShard  uint64
	RevocationSlot   uint16
	CreditCommitment common.Hash
	Deadline         *big.Int
	FeeRecipient     common.Address
	FeeAmount        *big.Int
}
type UpdateParams struct {
	Token, Owner                common.Address
	Source                      protocol.InputSource
	Nullifier, CreditCommitment common.Hash
	RevocationShard             uint64
	RevocationSlot              uint16
	Deadline                    *big.Int
	FeeRecipient                common.Address
	FeeAmount                   *big.Int
}
type Payer struct {
	Payer  common.Address
	Amount *big.Int
}
type GrantRef struct {
	PayerIndex *big.Int
	Shard      uint64
	Slot       uint16
}
type DrawParams struct {
	Token                                          common.Address
	Sources                                        []protocol.InputSource
	Payers                                         []Payer
	GrantRefs                                      []GrantRef
	Nullifiers, CreditCommitments, RecoveryAmounts []common.Hash
	NoteCommitment                                 common.Hash
	TotalAmount, Deadline                          *big.Int
	FeeRecipient                                   common.Address
	FeeAmount                                      *big.Int
}

func DrawCapacity(count int) (int, error) {
	if count < 1 || count > 80 {
		return 0, fmt.Errorf("Credit draw requires 1..80 grants")
	}
	if count <= 5 {
		return count, nil
	}
	return (count + 9) / 10 * 10, nil
}
func (c Context) validate(token common.Address, deadline *big.Int) error {
	if err := uintField(c.ChainID, 64, "chain ID"); err != nil {
		return err
	}
	if c.ProtocolAddress.IsZero() {
		return fmt.Errorf("protocol required")
	}
	if err := uintField(c.ProtocolAddress, 160, "protocol"); err != nil {
		return err
	}
	if token != c.TokenAddress {
		return fmt.Errorf("token context mismatch")
	}
	if deadline == nil || deadline.Sign() < 0 || !deadline.IsUint64() || deadline.Uint64() < c.Timestamp {
		return fmt.Errorf("expired or invalid deadline")
	}
	return nil
}
func fieldHash(h common.Hash, nonzero bool) error {
	if _, err := primitives.HexToElement(h.Hex()); err != nil {
		return err
	}
	if nonzero && h == (common.Hash{}) {
		return fmt.Errorf("zero field")
	}
	return nil
}
func deliveryHash(ds []protocol.Delivery, required int, note bool) (*big.Int, error) {
	if len(ds) != required {
		return nil, fmt.Errorf("Credit delivery count mismatch")
	}
	for i, d := range ds {
		if err := protocol.ValidateDelivery(d, !(note && i == required-1), false); err != nil {
			return nil, err
		}
	}
	h, err := protocol.DeliveriesHash(ds)
	return h.Big(), err
}
func contextWords(c Context) []*big.Int {
	return []*big.Int{primitives.ElementToBigInt(c.ChainID), primitives.ElementToBigInt(c.ProtocolAddress), c.TokenAddress.Big()}
}
func split(h *big.Int) (*big.Int, *big.Int) {
	return new(big.Int).Rsh(new(big.Int).Set(h), 128), new(big.Int).And(new(big.Int).Set(h), new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
}
func authorize(c Context, name string, capacity int, words []*big.Int) ([]common.Hash, apitypes.TypedData, error) {
	operation := primitives.WalletOperation("")
	switch name {
	case "CreateCredit":
		operation = primitives.WalletCreateCredit
	case "CreditDraw":
		operation = primitives.WalletCreditDraw
	case "UpdateCredit":
		operation = primitives.WalletUpdateCredit
	default:
		return nil, apitypes.TypedData{}, fmt.Errorf("unsupported Credit operation")
	}
	hashes := make([]common.Hash, len(words))
	fields := make([]fr.Element, len(words))
	for i, w := range words {
		if w == nil || w.Sign() < 0 || w.BitLen() > 254 {
			return nil, apitypes.TypedData{}, fmt.Errorf("invalid public field")
		}
		hashes[i] = common.BigToHash(w)
		if err := fieldHash(hashes[i], false); err != nil {
			return nil, apitypes.TypedData{}, err
		}
		fields[i] = primitives.BigIntToElement(w)
	}
	typed, err := primitives.WalletAuthorizationTypedData(primitives.WalletAuthorizationRequest{Context: c.ProtocolContext, Operation: operation, Capacity: capacity, Words: fields})
	if err != nil {
		return nil, typed, err
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return nil, typed, err
	}
	high, low := split(new(big.Int).SetBytes(digest))
	return append(hashes, common.BigToHash(high), common.BigToHash(low)), typed, nil
}

func (p CreateParams) Statement(c Context, deliveries []protocol.Delivery) ([]common.Hash, apitypes.TypedData, error) {
	fee, feeErr := normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return nil, apitypes.TypedData{}, feeErr
	}
	if err := c.validate(p.Token, p.Deadline); err != nil {
		return nil, apitypes.TypedData{}, err
	}
	if p.Owner == (common.Address{}) || p.RevocationShard == 0 {
		return nil, apitypes.TypedData{}, fmt.Errorf("invalid Credit owner or allocation")
	}
	if err := fieldHash(p.CreditCommitment, true); err != nil {
		return nil, apitypes.TypedData{}, err
	}
	h, err := deliveryHash(deliveries, 2, false)
	if err != nil {
		return nil, apitypes.TypedData{}, err
	}
	high, low := split(h)
	words := append(contextWords(c), p.Owner.Big(), new(big.Int).SetUint64(p.RevocationShard), new(big.Int).SetUint64(uint64(p.RevocationSlot)), p.CreditCommitment.Big(), p.Deadline, high, low)
	words = append(words, p.FeeRecipient.Big(), fee)
	return authorize(c, "CreateCredit", 1, words)
}
func (p UpdateParams) Statement(c Context, deliveries []protocol.Delivery) ([]common.Hash, apitypes.TypedData, error) {
	fee, feeErr := normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return nil, apitypes.TypedData{}, feeErr
	}
	if err := c.validate(p.Token, p.Deadline); err != nil {
		return nil, apitypes.TypedData{}, err
	}
	if p.Owner == (common.Address{}) || p.RevocationShard == 0 || p.Source.TreeId == nil || p.Source.TreeId.Sign() <= 0 || !p.Source.TreeId.IsUint64() {
		return nil, apitypes.TypedData{}, fmt.Errorf("invalid Credit owner, allocation or tree")
	}
	for _, h := range []common.Hash{p.Source.MerkleRoot, p.Nullifier, p.CreditCommitment} {
		if err := fieldHash(h, true); err != nil {
			return nil, apitypes.TypedData{}, err
		}
	}
	h, err := deliveryHash(deliveries, 2, false)
	if err != nil {
		return nil, apitypes.TypedData{}, err
	}
	high, low := split(h)
	words := append(contextWords(c), p.Owner.Big(), p.Source.TreeId, common.Hash(p.Source.MerkleRoot).Big(), p.Nullifier.Big(), p.CreditCommitment.Big(), new(big.Int).SetUint64(p.RevocationShard), new(big.Int).SetUint64(uint64(p.RevocationSlot)), p.Deadline, high, low)
	words = append(words, p.FeeRecipient.Big(), fee)
	return authorize(c, "UpdateCredit", 1, words)
}
func (p DrawParams) Statement(c Context, deliveries []protocol.Delivery) ([]common.Hash, apitypes.TypedData, error) {
	fee, feeErr := normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return nil, apitypes.TypedData{}, feeErr
	}
	if err := c.validate(p.Token, p.Deadline); err != nil {
		return nil, apitypes.TypedData{}, err
	}
	n := len(p.Nullifiers)
	capacity, err := DrawCapacity(n)
	if err != nil {
		return nil, apitypes.TypedData{}, err
	}
	if len(p.CreditCommitments) != n || len(p.RecoveryAmounts) != n || len(p.GrantRefs) != n || len(p.Payers) < 1 || len(p.Payers) > n {
		return nil, apitypes.TypedData{}, fmt.Errorf("Credit group count mismatch")
	}
	sources, err := protocol.NormalizeSources(p.Sources, n, capacity)
	if err != nil {
		return nil, apitypes.TypedData{}, err
	}
	seenN, seenO := map[common.Hash]bool{}, map[common.Hash]bool{}
	for i, nf := range p.Nullifiers {
		if err = fieldHash(nf, true); err != nil {
			return nil, apitypes.TypedData{}, err
		}
		if seenN[nf] {
			return nil, apitypes.TypedData{}, fmt.Errorf("duplicate nullifier")
		}
		seenN[nf] = true
		h := p.CreditCommitments[i]
		if err = fieldHash(h, true); err != nil {
			return nil, apitypes.TypedData{}, err
		}
		if seenO[h] {
			return nil, apitypes.TypedData{}, fmt.Errorf("duplicate state commitment")
		}
		seenO[h] = true
		if err = fieldHash(p.RecoveryAmounts[i], false); err != nil {
			return nil, apitypes.TypedData{}, err
		}
	}
	if err = fieldHash(p.NoteCommitment, true); err != nil {
		return nil, apitypes.TypedData{}, err
	}
	if seenO[p.NoteCommitment] {
		return nil, apitypes.TypedData{}, fmt.Errorf("duplicate Note commitment")
	}
	total := new(big.Int)
	previous := new(big.Int)
	for _, payer := range p.Payers {
		if payer.Payer.Big().Cmp(previous) <= 0 || payer.Amount == nil || payer.Amount.Sign() <= 0 || payer.Amount.BitLen() > 128 {
			return nil, apitypes.TypedData{}, fmt.Errorf("payers must be sorted distinct with positive uint128 receipts")
		}
		previous = payer.Payer.Big()
		total.Add(total, payer.Amount)
	}
	if p.TotalAmount == nil || p.TotalAmount.Sign() <= 0 || p.TotalAmount.BitLen() > 128 || total.Cmp(p.TotalAmount) != 0 {
		return nil, apitypes.TypedData{}, fmt.Errorf("payer receipts do not equal total")
	}
	for i, r := range p.GrantRefs {
		if r.PayerIndex == nil || r.PayerIndex.Sign() < 0 || !r.PayerIndex.IsUint64() || r.PayerIndex.Uint64() >= uint64(len(p.Payers)) || r.Shard == 0 {
			return nil, apitypes.TypedData{}, fmt.Errorf("invalid grant reference")
		}
		if i > 0 {
			before := p.GrantRefs[i-1]
			cmp := r.PayerIndex.Cmp(before.PayerIndex)
			if cmp < 0 || cmp == 0 && (r.Shard < before.Shard || r.Shard == before.Shard && r.Slot <= before.Slot) {
				return nil, apitypes.TypedData{}, fmt.Errorf("grant references must be sorted distinct")
			}
		}
	}
	h, err := deliveryHash(deliveries, 2*n+1, true)
	if err != nil {
		return nil, apitypes.TypedData{}, err
	}
	high, low := split(h)
	words := append(contextWords(c), big.NewInt(int64(n)), big.NewInt(int64(len(sources))), big.NewInt(int64(len(p.Payers))))
	arrays := make([][]*big.Int, 10)
	for _, s := range sources {
		arrays[0] = append(arrays[0], s.TreeId)
		arrays[1] = append(arrays[1], common.Hash(s.MerkleRoot).Big())
	}
	for _, p := range p.Payers {
		arrays[2] = append(arrays[2], p.Payer.Big())
		arrays[3] = append(arrays[3], p.Amount)
	}
	for i, r := range p.GrantRefs {
		arrays[4] = append(arrays[4], r.PayerIndex)
		arrays[5] = append(arrays[5], new(big.Int).SetUint64(r.Shard))
		arrays[6] = append(arrays[6], new(big.Int).SetUint64(uint64(r.Slot)))
		arrays[7] = append(arrays[7], p.Nullifiers[i].Big())
		arrays[8] = append(arrays[8], p.CreditCommitments[i].Big())
		arrays[9] = append(arrays[9], p.RecoveryAmounts[i].Big())
	}
	for _, a := range arrays {
		for i := 0; i < capacity; i++ {
			v := new(big.Int)
			if i < len(a) {
				v = a[i]
			}
			words = append(words, v)
		}
	}
	periods := Periods(c.Timestamp)
	words = append(words, p.NoteCommitment.Big(), p.TotalAmount, p.Deadline, high, low, new(big.Int).SetUint64(periods.CurrentDay), new(big.Int).SetUint64(periods.CurrentMonth))
	if fee.Cmp(p.TotalAmount) >= 0 {
		return nil, apitypes.TypedData{}, fmt.Errorf("Credit fee must leave a positive ordinary Note")
	}
	words = append(words, p.FeeRecipient.Big(), fee)
	return authorize(c, "CreditDraw", capacity, words)
}

func normalizeFee(recipient common.Address, amount *big.Int) (*big.Int, error) {
	fee := new(big.Int)
	if amount != nil {
		fee.Set(amount)
	}
	if fee.Sign() < 0 || fee.BitLen() > 128 || (recipient == (common.Address{})) != (fee.Sign() == 0) {
		return nil, fmt.Errorf("Credit fee requires a nonzero recipient and positive uint128 amount, or an empty zero pair")
	}
	return fee, nil
}
