package payer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
)

func SelectTransactCapacity(inputs, outputs int) (int, error) {
	if inputs < 0 || outputs < 0 || inputs > 168 || outputs > 168 || inputs+outputs < 1 || inputs+outputs > 168 || (inputs == 0 && outputs < 2) {
		return 0, fmt.Errorf("unsupported transact shape")
	}
	for _, n := range []int{14, 56, 112, 168} {
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
	CiphertextsHash                       common.Hash
}

func NormalizeTransactRequest(c TransactionContext, p CrossTreeTransactParams, cs [][]byte, f FundingParams) (NormalizedTransact, error) {
	r := NormalizedTransact{Params: p}
	if err := ValidateContext(c); err != nil {
		return r, err
	}
	capacity, err := SelectTransactCapacity(len(p.Nullifiers), len(p.OutputCommitments))
	if err != nil {
		return r, err
	}
	r.Capacity = capacity
	if err := ValidateNotes(p.Nullifiers); err != nil {
		return r, err
	}
	if err := ValidateNotes(p.OutputCommitments); err != nil {
		return r, err
	}
	r.Params.Sources, err = NormalizeSources(p.Sources, len(p.Nullifiers), capacity)
	if err != nil {
		return r, err
	}
	if len(cs) != len(p.OutputCommitments) {
		return r, fmt.Errorf("ciphertexts count mismatch")
	}
	for _, v := range cs {
		if err := ValidateCiphertext(v, 178, -1); err != nil {
			return r, err
		}
	}
	if err := CheckUint(p.WithdrawalAmount, 128, "withdrawal amount"); err != nil {
		return r, err
	}
	if (p.WithdrawalAmount.Sign() == 0) != (p.WithdrawalAddress == common.Address{}) || (len(p.OutputCommitments) == 0 && p.WithdrawalAmount.Sign() == 0) {
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
	slots := make([]*big.Int, 0, len(p.Nullifiers)+len(p.OutputCommitments))
	for _, v := range p.Nullifiers {
		slots = append(slots, number(v))
	}
	for _, v := range p.OutputCommitments {
		slots = append(slots, number(v))
	}
	r.InputTreeIDs = padded(ids, capacity)
	r.MerkleRoots = padded(roots, capacity)
	r.SlotValues = padded(slots, capacity)
	r.CiphertextsHash = CiphertextsHash(cs)
	return r, nil
}
func transactWords(c TransactionContext, r NormalizedTransact) []*big.Int {
	high, low := splitHash(r.CiphertextsHash)
	p := r.Params
	out := []*big.Int{c.ChainID, c.ProtocolAddress.Big(), c.TokenAddress.Big(), big.NewInt(int64(len(p.Nullifiers))), big.NewInt(int64(len(p.OutputCommitments))), big.NewInt(int64(len(p.Sources)))}
	out = append(out, r.InputTreeIDs...)
	out = append(out, r.MerkleRoots...)
	out = append(out, r.SlotValues...)
	return append(out, high, low, p.WithdrawalAddress.Big(), p.WithdrawalAmount, p.Deadline, r.Funding.Amount, r.Funding.Payer.Big())
}
func TransactPublicInputs(c TransactionContext, p CrossTreeTransactParams, cs [][]byte, f FundingParams) ([][32]byte, error) {
	r, err := NormalizeTransactRequest(c, p, cs, f)
	if err != nil {
		return nil, err
	}
	return words(transactWords(c, r)), nil
}

var slotFields = []apitypes.Type{
	{Name: "token", Type: "address"}, {Name: "inputCount", Type: "uint256"}, {Name: "outputCount", Type: "uint256"}, {Name: "sourceCount", Type: "uint256"},
	{Name: "inputTreeIdsHash", Type: "bytes32"}, {Name: "merkleRootsHash", Type: "bytes32"}, {Name: "slotValuesHash", Type: "bytes32"}, {Name: "ciphertextsHash", Type: "bytes32"},
	{Name: "withdrawalAddress", Type: "address"}, {Name: "withdrawalAmount", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "depositAmount", Type: "uint256"}, {Name: "depositPayer", Type: "address"},
}

func EncodeTransactTypedData(c TransactionContext, p CrossTreeTransactParams, cs [][]byte, f FundingParams) (apitypes.TypedData, error) {
	r, err := NormalizeTransactRequest(c, p, cs, f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return typedData(c, fmt.Sprintf("TransactSlots%d", r.Capacity), slotFields, map[string]any{
		"token": c.TokenAddress.Hex(), "inputCount": fmt.Sprint(len(p.Nullifiers)), "outputCount": fmt.Sprint(len(p.OutputCommitments)), "sourceCount": fmt.Sprint(len(p.Sources)),
		"inputTreeIdsHash": hashWords(r.InputTreeIDs).Hex(), "merkleRootsHash": hashWords(r.MerkleRoots).Hex(), "slotValuesHash": hashWords(r.SlotValues).Hex(), "ciphertextsHash": r.CiphertextsHash.Hex(),
		"withdrawalAddress": p.WithdrawalAddress.Hex(), "withdrawalAmount": p.WithdrawalAmount.String(), "deadline": p.Deadline.String(), "depositAmount": r.Funding.Amount.String(), "depositPayer": r.Funding.Payer.Hex(),
	}), nil
}
func TransactUserPublicInputs(c TransactionContext, p CrossTreeTransactParams, cs [][]byte, f FundingParams) ([][32]byte, error) {
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
func TransactFundingRequestHash(proof []byte, p CrossTreeTransactParams, cs [][]byte) (common.Hash, error) {
	if err := CheckUint(p.WithdrawalAmount, 128, "withdrawal amount"); err != nil {
		return common.Hash{}, err
	}
	if err := CheckUint(p.Deadline, 64, "deadline"); err != nil {
		return common.Hash{}, err
	}
	if _, err := NormalizeSources(p.Sources, len(p.Nullifiers), 168); err != nil {
		return common.Hash{}, err
	}

	method, err := transactMethod(false)
	if err != nil {
		return common.Hash{}, err
	}
	if len(p.Sources) > 1 {
		return HashABI(abi.Arguments{{Type: mustType("bytes32")}, {Type: mustType("bytes32")}, {Type: *method.Inputs[1].Type.TupleElems[0]}, {Type: mustType("bytes32")}}, crypto.Keccak256Hash([]byte("InvisiCrossTreeTransact")), crypto.Keccak256Hash(proof), p, CiphertextsHash(cs))
	}
	components := []abi.ArgumentMarshaling{{Name: "merkleRoot", Type: "bytes32"}, {Name: "nullifiers", Type: "bytes32[]"}, {Name: "outputCommitments", Type: "bytes32[]"}, {Name: "inputTreeId", Type: "uint256"}, {Name: "withdrawalAddress", Type: "address"}, {Name: "withdrawalAmount", Type: "uint256"}, {Name: "deadline", Type: "uint256"}}
	singleType, err := abi.NewType("tuple", "", components)
	if err != nil {
		return common.Hash{}, err
	}
	single := struct {
		MerkleRoot                    [32]byte
		Nullifiers, OutputCommitments [][32]byte
		InputTreeId                   *big.Int
		WithdrawalAddress             common.Address
		WithdrawalAmount, Deadline    *big.Int
	}{
		Nullifiers: p.Nullifiers, OutputCommitments: p.OutputCommitments, InputTreeId: new(big.Int), WithdrawalAddress: p.WithdrawalAddress, WithdrawalAmount: p.WithdrawalAmount, Deadline: p.Deadline}
	if len(p.Sources) == 1 {
		single.MerkleRoot = p.Sources[0].MerkleRoot
		single.InputTreeId = p.Sources[0].TreeId
	}
	return HashABI(abi.Arguments{{Type: mustType("bytes32")}, {Type: singleType}, {Type: mustType("bytes32")}}, crypto.Keccak256Hash(proof), single, CiphertextsHash(cs))
}
func EncodeTransactFundingTypedData(c TransactionContext, proof []byte, p CrossTreeTransactParams, cs [][]byte, f FundingParams) (apitypes.TypedData, error) {
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
	r, err := NormalizeTransactRequest(p.Context, p.Params, p.Ciphertexts, p.Funding)
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
	data, err := PackProtocolCalldata(method.Name, p.Proof, TransactRequest{Params: r.Params, Ciphertexts: p.Ciphertexts, Funding: r.Funding})
	if err != nil {
		return nil, err
	}
	return &Calldata{Data: data, Capacity: r.Capacity, ApproveAmount: new(big.Int).Set(r.Funding.Amount)}, nil
}
func AggregateMemberRequestHash(c TransactionContext, r TransactRequest) (common.Hash, error) {
	normalized, err := NormalizeTransactRequest(c, r.Params, r.Ciphertexts, r.Funding)
	if err != nil {
		return common.Hash{}, err
	}
	if normalized.Capacity != 14 {
		return common.Hash{}, fmt.Errorf("aggregate supports only capacity14")
	}
	return HashABI(args("bytes32", "uint256", "bytes32[]"), crypto.Keccak256Hash([]byte("InvisiAtomicBatchMemberV1")), big.NewInt(14), words(transactWords(c, normalized)))
}
func validAggregateCount(n int) bool { return n == 4 || n == 8 || n == 12 }

// AggregatePublicInputs returns router business words, not aggregate circuit public inputs.
func AggregatePublicInputs(c TransactionContext, requests []TransactRequest) ([][32]byte, error) {
	if !validAggregateCount(len(requests)) {
		return nil, fmt.Errorf("aggregate requires 4,8 or12 requests")
	}
	var out [][32]byte
	for _, r := range requests {
		if _, err := AggregateMemberRequestHash(c, r); err != nil {
			return nil, err
		}
		v, err := TransactPublicInputs(c, r.Params, r.Ciphertexts, r.Funding)
		if err != nil {
			return nil, err
		}
		out = append(out, Word(big.NewInt(14)))
		out = append(out, v...)
	}
	return out, nil
}
func EncodeAggregateFundingTypedData(c TransactionContext, r TransactRequest) (apitypes.TypedData, error) {
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
		return nil, fmt.Errorf("aggregate requires 4,8 or12 requests")
	}
	requests := make([]TransactRequest, len(p.Requests))
	total := new(big.Int)
	seenN, seenO := map[[32]byte]bool{}, map[[32]byte]bool{}
	for i, r := range p.Requests {
		n, err := NormalizeTransactRequest(p.Context, r.Params, r.Ciphertexts, r.Funding)
		if err != nil {
			return nil, err
		}
		if n.Capacity != 14 {
			return nil, fmt.Errorf("aggregate supports only capacity14")
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
		for _, v := range n.Params.OutputCommitments {
			if seenO[v] {
				return nil, fmt.Errorf("duplicate aggregate commitment")
			}
			seenO[v] = true
		}
		requests[i] = TransactRequest{Params: n.Params, Ciphertexts: r.Ciphertexts, Funding: n.Funding}
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
	return &Calldata{Data: data, Capacity: 14, ApproveAmount: total}, nil
}
