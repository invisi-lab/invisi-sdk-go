package protocol

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// CreditDrawCircuit describes the smallest current circuit that fits a draw.
type CreditDrawCircuit struct {
	Name                                 string
	MaxSize, PublicInputCount, TotalBits int
}

func SelectCreditDrawCircuit(count int) (CreditDrawCircuit, error) {
	if count < 1 || count > 120 {
		return CreditDrawCircuit{}, fmt.Errorf("draw requires 1..120 credits")
	}
	for _, n := range []int{1, 12, 50, 120} {
		if count <= n {
			return CreditDrawCircuit{Name: fmt.Sprintf("credit_draw_slots%d", n), MaxSize: n, PublicInputCount: 6*n + 14, TotalBits: 126 + bits.Len(uint(n-1))}, nil
		}
	}
	return CreditDrawCircuit{}, fmt.Errorf("unsupported Credit draw shape")
}

type NormalizedCreditDraw struct {
	Params  CreditDrawParams
	Profile CreditDrawCircuit
}

// ValidateCreditDrawRequest validates public request structure; it cannot establish
// ownership or conservation without the private witness and proof verification.
func ValidateCreditDrawRequest(c TransactionContext, p CreditDrawParams, cs []Delivery) (NormalizedCreditDraw, error) {
	r := NormalizedCreditDraw{Params: p}
	if err := ValidateContext(c); err != nil {
		return r, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return r, err
	}
	r.Profile = profile
	if err := CheckUint(c.Now, 64, "chain timestamp"); err != nil {
		return r, err
	}
	if err := ValidateNotes(p.Nullifiers); err != nil {
		return r, err
	}
	r.Params.Sources, err = NormalizeSources(p.Sources, len(p.Nullifiers), profile.MaxSize)
	if err != nil {
		return r, err
	}
	n := len(p.Nullifiers)
	if len(p.RemainderCommitments) != n || len(p.Expiries) != n || len(p.RecoveryAmounts) != n {
		return r, fmt.Errorf("Credit draw array count mismatch")
	}
	if len(cs) != 2*n {
		return r, fmt.Errorf("deliveries count mismatch")
	}
	seen := make(map[[32]byte]bool, n)
	for i, remainder := range p.RemainderCommitments {
		if err := Field(remainder); err != nil {
			return r, err
		}
		if err := Field(p.RecoveryAmounts[i]); err != nil {
			return r, err
		}
		if remainder != ([32]byte{}) {
			if seen[remainder] {
				return r, fmt.Errorf("duplicate remainder commitment")
			}
			seen[remainder] = true
		} else if p.RecoveryAmounts[i] != ([32]byte{}) {
			return r, fmt.Errorf("full draw requires zero recovery amount")
		}
		if err := CheckUint(p.Expiries[i], 64, "expiry"); err != nil {
			return r, err
		}
		if p.Expiries[i].Sign() == 0 || (c.Now != nil && p.Expiries[i].Cmp(c.Now) <= 0) {
			return r, fmt.Errorf("expiry must be future")
		}
		for j := 0; j < 2; j++ {
			if err := ValidateDelivery(cs[2*i+j], true, remainder == ([32]byte{})); err != nil {
				return r, err
			}
		}

	}
	if err := CheckUint(p.TotalAmount, profile.TotalBits, "totalAmount"); err != nil {
		return r, err
	}
	if p.Spender == (common.Address{}) || p.TotalAmount.Sign() == 0 {
		return r, fmt.Errorf("spender and total amount must be positive")
	}
	if err := CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	return r, nil
}

func CreditDrawBaseInputs(c TransactionContext, p CreditDrawParams, cs []Delivery) ([][32]byte, error) {
	r, err := ValidateCreditDrawRequest(c, p, cs)
	if err != nil {
		return nil, err
	}
	n := r.Profile.MaxSize
	ids, roots := make([]*big.Int, len(r.Params.Sources)), make([]*big.Int, len(r.Params.Sources))
	for i, s := range r.Params.Sources {
		ids[i] = s.TreeId
		roots[i] = creditNumber(s.MerkleRoot)
	}
	h, err := DeliveriesHash(cs)
	if err != nil {
		return nil, err
	}
	v := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), p.Spender.Big(), big.NewInt(int64(len(p.Nullifiers))), big.NewInt(int64(len(ids)))}
	v = append(v, creditPadded(ids, n)...)
	v = append(v, creditPadded(roots, n)...)
	v = append(v, creditPadded(creditNumbers(p.Nullifiers), n)...)
	v = append(v, creditPadded(creditNumbers(p.RemainderCommitments), n)...)
	v = append(v, creditPadded(p.Expiries, n)...)
	v = append(v, p.TotalAmount, p.Deadline, new(big.Int).SetBytes(h[:16]), new(big.Int).SetBytes(h[16:]))
	v = append(v, creditPadded(creditNumbers(p.RecoveryAmounts), n)...)
	periods := core.CreditPeriods(c.Now.Uint64())
	v = append(v, new(big.Int).SetUint64(periods.CurrentDay), new(big.Int).SetUint64(periods.CurrentMonth))
	return creditWords(v), nil
}

func CreditDrawPublicInputs(c TransactionContext, p CreditDrawParams, cs []Delivery) ([][32]byte, error) {
	base, err := CreditDrawBaseInputs(c, p, cs)
	if err != nil {
		return nil, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return nil, err
	}
	return AppendWalletAuthorization(c, core.WalletCreditDraw, profile.MaxSize, base)
}

func EncodeCreditDrawTypedData(c TransactionContext, p CreditDrawParams, cs []Delivery) (apitypes.TypedData, error) {
	base, err := CreditDrawBaseInputs(c, p, cs)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	profile, err := SelectCreditDrawCircuit(len(p.Nullifiers))
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, core.WalletCreditDraw, profile.MaxSize, base)
}

func BuildCreditDrawCalldata(p BuildCreditDrawCalldataParams) (*Calldata, error) {
	r, err := ValidateCreditDrawRequest(p.Context, p.Params, p.Deliveries)
	if err != nil {
		return nil, err
	}
	proof, err := EncodeCreditProof(p.Proof)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata("creditDraw", proof, r.Params, p.Deliveries)
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, ApproveAmount: new(big.Int), Capacity: r.Profile.MaxSize}, nil
}
