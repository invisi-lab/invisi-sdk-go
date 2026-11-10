package protocol

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func SelectTransactCapacity(inputs, outputs int) (int, error) {
	if inputs < 0 || outputs < 0 || inputs > 190 || outputs > 190 || inputs+outputs < 1 || inputs+outputs > 190 || (inputs == 0 && outputs < 2) {
		return 0, fmt.Errorf("unsupported transact shape")
	}
	for _, n := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190} {
		if inputs+outputs <= n {
			return n, nil
		}
	}
	return 0, fmt.Errorf("unsupported transact shape")
}

type NormalizedTransact struct {
	Capacity                              int
	Params                                CrossTreeTransactParams
	Funding                               FundingParams
	InputTreeIDs, MerkleRoots, SlotValues []*big.Int
	DeliveriesHash                        common.Hash
}

func NormalizeTransactRequest(c TransactionContext, p CrossTreeTransactParams, cs []Delivery, f FundingParams) (NormalizedTransact, error) {
	r := NormalizedTransact{Params: p}
	if p.Token != c.TokenAddress {
		return r, fmt.Errorf("request asset context mismatch")
	}
	var feeErr error
	r.Params.FeeAmount, feeErr = normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return r, feeErr
	}
	if err := ValidateContext(c); err != nil {
		return r, err
	}
	capacity, err := SelectTransactCapacity(len(p.Nullifiers), len(p.NoteCommitments))
	if err != nil {
		return r, err
	}
	r.Capacity = capacity
	if err := ValidateNotes(p.Nullifiers); err != nil {
		return r, err
	}
	if err := ValidateNotes(p.NoteCommitments); err != nil {
		return r, err
	}
	r.Params.Sources, err = NormalizeSources(p.Sources, len(p.Nullifiers), capacity)
	if err != nil {
		return r, err
	}
	if len(cs) != len(p.NoteCommitments) {
		return r, fmt.Errorf("deliveries count mismatch")
	}
	for _, v := range cs {
		if err := ValidateDelivery(v, false, false); err != nil {
			return r, err
		}
	}
	if err := CheckUint(p.WithdrawalAmount, 128, "withdrawal amount"); err != nil {
		return r, err
	}
	if (p.WithdrawalAmount.Sign() == 0) != (p.WithdrawalAddress == common.Address{}) || (len(p.NoteCommitments) == 0 && p.WithdrawalAmount.Sign() == 0 && r.Params.FeeAmount.Sign() == 0) {
		return r, fmt.Errorf("invalid withdrawal")
	}
	if err := CheckDeadline(c, p.Deadline); err != nil {
		return r, err
	}
	r.Funding, err = NormalizeFunding(f)
	if err != nil {
		return r, err
	}
	if len(p.Nullifiers) == 0 && (r.Funding.Amount.Sign() == 0 || p.WithdrawalAmount.Sign() != 0) {
		return r, fmt.Errorf("zero inputs require positive funding and zero withdrawal")
	}
	ids, roots := make([]*big.Int, len(p.Sources)), make([]*big.Int, len(p.Sources))
	for i, s := range r.Params.Sources {
		ids[i] = s.TreeId
		roots[i] = number(s.MerkleRoot)
	}
	slots := make([]*big.Int, 0, len(p.Nullifiers)+len(p.NoteCommitments))
	for _, v := range p.Nullifiers {
		slots = append(slots, number(v))
	}
	for _, v := range p.NoteCommitments {
		slots = append(slots, number(v))
	}
	r.InputTreeIDs = padded(ids, capacity)
	r.MerkleRoots = padded(roots, capacity)
	r.SlotValues = padded(slots, capacity)
	r.DeliveriesHash, err = DeliveriesHash(cs)
	if err != nil {
		return r, err
	}
	return r, nil
}
func transactWords(c TransactionContext, r NormalizedTransact) []*big.Int {
	high, low := splitHash(r.DeliveriesHash)
	p := r.Params
	out := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), big.NewInt(int64(len(p.Nullifiers))), big.NewInt(int64(len(p.NoteCommitments))), big.NewInt(int64(len(p.Sources)))}
	out = append(out, r.InputTreeIDs...)
	out = append(out, r.MerkleRoots...)
	out = append(out, r.SlotValues...)
	return append(out, high, low, p.WithdrawalAddress.Big(), p.WithdrawalAmount, p.Deadline, r.Funding.Amount, r.Funding.Payer.Big(), p.FeeRecipient.Big(), p.FeeAmount)
}
func TransactPublicInputs(c TransactionContext, p CrossTreeTransactParams, cs []Delivery, f FundingParams) ([][32]byte, error) {
	r, err := NormalizeTransactRequest(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	return words(transactWords(c, r)), nil
}

func EncodeTransactTypedData(c TransactionContext, p CrossTreeTransactParams, cs []Delivery, f FundingParams) (apitypes.TypedData, error) {
	r, err := NormalizeTransactRequest(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeWalletAuthorizationTypedData(c, primitives.WalletTransact, r.Capacity, words(transactWords(c, r)))
}
func TransactUserPublicInputs(c TransactionContext, p CrossTreeTransactParams, cs []Delivery, f FundingParams) ([][32]byte, error) {
	inputs, err := TransactPublicInputs(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	td, err := EncodeTransactTypedData(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	digest, err := TypedDataHash(td)
	if err != nil {
		return nil, err
	}
	high, low := splitHash(digest)
	return append(inputs, Word(high), Word(low)), nil
}
func transactMethod(aggregate bool) (abi.Method, error) {
	a, err := ProtocolABI()
	if err != nil {
		return abi.Method{}, err
	}
	for _, m := range a.Methods {
		if m.RawName == "transact" && len(m.Inputs) == 2 && ((!aggregate && m.Inputs[1].Type.T == abi.TupleTy) || (aggregate && m.Inputs[1].Type.T == abi.SliceTy)) {
			return m, nil
		}
	}
	return abi.Method{}, fmt.Errorf("current transact overload unavailable")
}
func TransactFundingRequestHash(proof []byte, p CrossTreeTransactParams, cs []Delivery) (common.Hash, error) {
	var feeErr error
	p.FeeAmount, feeErr = normalizeFee(p.FeeRecipient, p.FeeAmount)
	if feeErr != nil {
		return common.Hash{}, feeErr
	}
	if err := CheckUint(p.WithdrawalAmount, 128, "withdrawal amount"); err != nil {
		return common.Hash{}, err
	}
	if err := CheckUint(p.Deadline, 64, "deadline"); err != nil {
		return common.Hash{}, err
	}
	if _, err := NormalizeSources(p.Sources, len(p.Nullifiers), 190); err != nil {
		return common.Hash{}, err
	}

	deliveryHash, err := DeliveriesHash(cs)
	if err != nil {
		return common.Hash{}, err
	}

	method, err := transactMethod(false)
	if err != nil {
		return common.Hash{}, err
	}
	if len(p.Sources) > 1 {
		return HashABI(abi.Arguments{{Type: mustType("bytes32")}, {Type: mustType("bytes32")}, {Type: *method.Inputs[1].Type.TupleElems[0]}, {Type: mustType("bytes32")}}, crypto.Keccak256Hash([]byte("InvisiCrossTreeTransact")), crypto.Keccak256Hash(proof), p, deliveryHash)
	}
	components := []abi.ArgumentMarshaling{{Name: "token", Type: "address"}, {Name: "merkleRoot", Type: "bytes32"}, {Name: "nullifiers", Type: "bytes32[]"}, {Name: "noteCommitments", Type: "bytes32[]"}, {Name: "inputTreeId", Type: "uint256"}, {Name: "withdrawalAddress", Type: "address"}, {Name: "withdrawalAmount", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "feeRecipient", Type: "address"}, {Name: "feeAmount", Type: "uint256"}}
	singleType, err := abi.NewType("tuple", "", components)
	if err != nil {
		return common.Hash{}, err
	}
	single := struct {
		Token                       common.Address
		MerkleRoot                  [32]byte
		Nullifiers, NoteCommitments [][32]byte
		InputTreeId                 *big.Int
		WithdrawalAddress           common.Address
		WithdrawalAmount, Deadline  *big.Int
		FeeRecipient                common.Address
		FeeAmount                   *big.Int
	}{
		Token: p.Token, Nullifiers: p.Nullifiers, NoteCommitments: p.NoteCommitments, InputTreeId: new(big.Int), WithdrawalAddress: p.WithdrawalAddress, WithdrawalAmount: p.WithdrawalAmount, Deadline: p.Deadline, FeeRecipient: p.FeeRecipient, FeeAmount: p.FeeAmount}
	if len(p.Sources) == 1 {
		single.MerkleRoot = p.Sources[0].MerkleRoot
		single.InputTreeId = p.Sources[0].TreeId
	}
	return HashABI(abi.Arguments{{Type: mustType("bytes32")}, {Type: singleType}, {Type: mustType("bytes32")}}, crypto.Keccak256Hash(proof), single, deliveryHash)
}
func EncodeTransactFundingTypedData(c TransactionContext, proof []byte, p CrossTreeTransactParams, cs []Delivery, f FundingParams) (apitypes.TypedData, error) {
	r, err := NormalizeTransactRequest(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	hash, err := TransactFundingRequestHash(proof, p, cs)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeFundingTypedData(c, "FundTransact", hash, r.Funding)
}
func BuildTransactCalldata(p BuildTransactCalldataParams) (*Calldata, error) {
	r, err := NormalizeTransactRequest(p.Context, p.Params, p.Deliveries, p.Funding)
	if err != nil {
		return nil, err
	}
	if err := ValidateFundingSubmission(r.Funding, p.Caller, p.Context.Now); err != nil {
		return nil, err
	}
	method, err := transactMethod(false)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata(method.Name, p.Proof, TransactRequest{Params: r.Params, Deliveries: p.Deliveries, Funding: r.Funding})
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, Capacity: r.Capacity, ApproveAmount: new(big.Int).Set(r.Funding.Amount)}, nil
}
func AggregateMemberRequestHash(c TransactionContext, r TransactRequest) (common.Hash, error) {
	c.TokenAddress = r.Params.Token
	normalized, err := NormalizeTransactRequest(c, r.Params, r.Deliveries, r.Funding)
	if err != nil {
		return common.Hash{}, err
	}
	if normalized.Capacity > 10 {
		return common.Hash{}, fmt.Errorf("aggregate supports native capacities1 through10")
	}
	return HashABI(args("bytes32", "uint256", "bytes32[]"), crypto.Keccak256Hash([]byte("InvisiAtomicBatchMemberV1")), big.NewInt(int64(normalized.Capacity)), words(transactWords(c, normalized)))
}
func validAggregateCount(n int) bool { return n >= 1 && n <= 5 }

// Zero identifies a mixed outer, never a member capacity.
func SelectAggregateRoute(capacities []int) (int, error) {
	if !validAggregateCount(len(capacities)) {
		return 0, fmt.Errorf("aggregate requires one through five members")
	}
	first, homogeneous, small := capacities[0], true, true
	for _, c := range capacities {
		if c < 1 || c > 10 {
			return 0, fmt.Errorf("unsupported aggregate capacity")
		}
		homogeneous = homogeneous && c == first
		small = small && c <= 5
	}
	if homogeneous && (len(capacities) == 1 || len(capacities) == 2 || len(capacities) == 5) {
		return first, nil
	}
	if small && len(capacities) >= 2 {
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported aggregate route")
}
func aggregateRequestRoute(requests []TransactRequest) (int, error) {
	capacities := []int{}
	for _, r := range requests {
		c, e := SelectTransactCapacity(len(r.Params.Nullifiers), len(r.Params.NoteCommitments))
		if e != nil {
			return 0, e
		}
		capacities = append(capacities, c)
	}
	return SelectAggregateRoute(capacities)
}

// AggregatePublicInputs returns router business words, not aggregate circuit public inputs.
func AggregatePublicInputs(c TransactionContext, requests []TransactRequest) ([][32]byte, error) {
	if !validAggregateCount(len(requests)) {
		return nil, fmt.Errorf("aggregate requires one through five requests")
	}
	if _, err := aggregateRequestRoute(requests); err != nil {
		return nil, err
	}
	var out [][32]byte
	for _, r := range requests {
		c.TokenAddress = r.Params.Token
		if _, err := AggregateMemberRequestHash(c, r); err != nil {
			return nil, err
		}
		v, err := TransactPublicInputs(c, r.Params, r.Deliveries, r.Funding)
		if err != nil {
			return nil, err
		}
		capacity, _ := SelectTransactCapacity(len(r.Params.Nullifiers), len(r.Params.NoteCommitments))
		out = append(out, Word(big.NewInt(int64(capacity))))
		out = append(out, v...)
	}
	return out, nil
}
func EncodeAggregateFundingTypedData(c TransactionContext, r TransactRequest) (apitypes.TypedData, error) {
	c.TokenAddress = r.Params.Token
	hash, err := AggregateMemberRequestHash(c, r)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	f, err := NormalizeFunding(r.Funding)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return EncodeFundingTypedData(c, "FundTransact", hash, f)
}
func BuildTransactAggregateCalldata(p BuildTransactAggregateCalldataParams) (*Calldata, error) {
	if !validAggregateCount(len(p.Requests)) {
		return nil, fmt.Errorf("aggregate requires one through five requests")
	}
	route, err := aggregateRequestRoute(p.Requests)
	if err != nil {
		return nil, err
	}
	requests := make([]TransactRequest, len(p.Requests))
	total := new(big.Int)
	seenN, seenO := map[[32]byte]bool{}, map[[32]byte]bool{}
	for i, r := range p.Requests {
		context := p.Context
		context.TokenAddress = r.Params.Token
		n, err := NormalizeTransactRequest(context, r.Params, r.Deliveries, r.Funding)
		if err != nil {
			return nil, err
		}
		if n.Capacity > 10 {
			return nil, fmt.Errorf("aggregate supports native capacities1 through10")
		}
		if err := ValidateFundingSubmission(n.Funding, p.Caller, p.Context.Now); err != nil {
			return nil, err
		}
		for _, v := range n.Params.Nullifiers {
			if seenN[v] {
				return nil, fmt.Errorf("duplicate aggregate nullifier")
			}
			seenN[v] = true
		}
		for _, v := range n.Params.NoteCommitments {
			if seenO[v] {
				return nil, fmt.Errorf("duplicate aggregate commitment")
			}
			seenO[v] = true
		}
		requests[i] = TransactRequest{Params: n.Params, Deliveries: r.Deliveries, Funding: n.Funding}
		total.Add(total, n.Funding.Amount)
	}
	method, err := transactMethod(true)
	if err != nil {
		return nil, err
	}
	data, err := PackProtocolCalldata(method.Name, p.Proof, requests)
	if err != nil {
		return nil, err
	}
	// A scalar approval cannot represent multiple assets; callers inspect each explicit funding request.
	for _, r := range p.Requests {
		if r.Params.Token != p.Requests[0].Params.Token {
			total = new(big.Int)
			break
		}
	}
	return &Calldata{Data: data, Capacity: route, ApproveAmount: total}, nil
}
