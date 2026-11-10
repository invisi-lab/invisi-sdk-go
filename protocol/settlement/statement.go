package settlement

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func packHash(kinds []string, values ...any) (common.Hash, error) {
	a := abi.Arguments{}
	for _, k := range kinds {
		t, e := abi.NewType(k, "", nil)
		if e != nil {
			return common.Hash{}, e
		}
		a = append(a, abi.Argument{Type: t})
	}
	b, e := a.Pack(values...)
	return crypto.Keccak256Hash(b), e
}

// TransactStatement reproduces TransactAuthorizationLib including the two authorization limbs.
// The native member capacity is retained even when an aggregate outer uses wider padding.
func TransactStatement(chainID *big.Int, protocol common.Address, r *Request) (uint32, [][32]byte, common.Hash, error) {
	bad := func() (uint32, [][32]byte, common.Hash, error) {
		return 0, nil, common.Hash{}, errors.New("invalid canonical transact statement")
	}
	if r == nil || r.Method != "transact" || chainID == nil || !chainID.IsUint64() || protocol == (common.Address{}) {
		return bad()
	}
	p, f := r.Params, r.Funding
	n, m := len(p.Nullifiers), len(p.NoteCommitments)
	c := n + m
	if c < 1 || c > 190 || n == 0 && m < 2 {
		return bad()
	}
	if c > 10 {
		c = (c + 9) / 10 * 10
	}
	if len(p.Sources) > n || n > 0 && len(p.Sources) == 0 || len(r.Deliveries) != m {
		return bad()
	}
	if p.WithdrawalAmount == nil || p.Deadline == nil || p.FeeAmount == nil || f.Amount == nil {
		return bad()
	}
	if n == 0 && (len(p.Sources) != 0 || p.WithdrawalAmount.Sign() != 0 || p.WithdrawalAddress != (common.Address{}) || f.Amount.Sign() <= 0 || f.Payer == (common.Address{})) {
		return bad()
	}
	for _, v := range []*big.Int{p.WithdrawalAmount, p.FeeAmount, f.Amount} {
		if v.Sign() < 0 || v.BitLen() > 128 {
			return bad()
		}
	}
	if p.Deadline.Sign() < 0 || p.Deadline.BitLen() > 64 {
		return bad()
	}
	if p.Token == (common.Address{}) || (f.Amount.Sign() == 0) != (f.Payer == (common.Address{})) || (p.WithdrawalAmount.Sign() == 0) != (p.WithdrawalAddress == (common.Address{})) || (p.FeeAmount.Sign() == 0) != (p.FeeRecipient == (common.Address{})) {
		return bad()
	}
	seen := map[[32]byte]bool{}
	for _, items := range [][][32]byte{p.Nullifiers, p.NoteCommitments} {
		for _, v := range items {
			if v == ([32]byte{}) || seen[v] {
				return bad()
			}
			seen[v] = true
		}
	}
	previous := new(big.Int)
	for _, source := range p.Sources {
		if source.TreeId == nil || source.TreeId.Cmp(previous) <= 0 || source.MerkleRoot == ([32]byte{}) {
			return bad()
		}
		previous = source.TreeId
	}
	for _, delivery := range r.Deliveries {
		if len(delivery.Ciphertext) == 0 && len(delivery.Clue) == 0 {
			continue
		}
		if len(delivery.Ciphertext) == 0 || len(delivery.Ciphertext) > 4096 || len(delivery.Clue) != 68 || delivery.Clue[64] != 24 {
			return bad()
		}
	}
	tuple, e := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{{Name: "ciphertext", Type: "bytes"}, {Name: "clue", Type: "bytes"}})
	if e != nil {
		return bad()
	}
	b, e := (abi.Arguments{{Type: tuple}}).Pack(r.Deliveries)
	if e != nil {
		return bad()
	}
	dh := crypto.Keccak256Hash(b)
	w := make([][32]byte, 15+3*c)
	put := func(i int, v *big.Int) { w[i] = common.BigToHash(v) }
	put(0, chainID)
	put(1, protocol.Big())
	put(2, p.Token.Big())
	put(3, big.NewInt(int64(n)))
	put(4, big.NewInt(int64(m)))
	put(5, big.NewInt(int64(len(p.Sources))))
	for i, s := range p.Sources {
		if s.TreeId == nil || s.TreeId.Sign() < 0 || !s.TreeId.IsUint64() {
			return bad()
		}
		put(6+i, s.TreeId)
		w[6+c+i] = s.MerkleRoot
	}
	for i, v := range p.Nullifiers {
		w[6+2*c+i] = v
	}
	for i, v := range p.NoteCommitments {
		w[6+2*c+n+i] = v
	}
	off := 6 + 3*c
	put(off, new(big.Int).SetBytes(dh[:16]))
	put(off+1, new(big.Int).SetBytes(dh[16:]))
	put(off+2, p.WithdrawalAddress.Big())
	put(off+3, p.WithdrawalAmount)
	put(off+4, p.Deadline)
	put(off+5, f.Amount)
	put(off+6, f.Payer.Big())
	put(off+7, p.FeeRecipient.Big())
	put(off+8, p.FeeAmount)
	field, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	for _, v := range w {
		if new(big.Int).SetBytes(v[:]).Cmp(field) >= 0 {
			return bad()
		}
	}
	domain, e := packHash([]string{"bytes32", "bytes32", "bytes32", "uint256", "address"}, crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")), crypto.Keccak256Hash([]byte("InvisibleProtocol")), crypto.Keccak256Hash([]byte("1")), chainID, protocol)
	if e != nil {
		return bad()
	}
	statement, e := packHash([]string{"uint256", "bytes32[]"}, big.NewInt(int64(c)), w)
	if e != nil {
		return bad()
	}
	auth, e := packHash([]string{"bytes32", "bytes32", "bytes32"}, crypto.Keccak256Hash([]byte("WalletAuthorization(bytes32 operation,bytes32 statementHash)")), crypto.Keccak256Hash([]byte("Invisi.Transact.v1")), statement)
	if e != nil {
		return bad()
	}
	digest := crypto.Keccak256Hash([]byte{0x19, 0x01}, domain[:], auth[:])
	w = append(w, common.BigToHash(new(big.Int).SetBytes(digest[:16])), common.BigToHash(new(big.Int).SetBytes(digest[16:])))
	operation, e := packHash([]string{"string", "uint64", "address", "uint8", "uint32", "bytes32[]"}, "Invisi.Provenance.Operation.v1", chainID.Uint64(), protocol, uint8(2), uint32(c), w)
	return uint32(c), w, operation, e
}

// AggregateMemberHash authenticates the BatchRequest.statementHash without outer padding.
func AggregateMemberHash(chainID *big.Int, protocol common.Address, r *Request) (common.Hash, error) {
	c, w, _, err := TransactStatement(chainID, protocol, r)
	if err != nil {
		return common.Hash{}, err
	}
	if c > 10 {
		return common.Hash{}, ErrCalldata
	}
	return packHash([]string{"bytes32", "uint256", "bytes32[]"}, crypto.Keccak256Hash([]byte("InvisiAtomicBatchMemberV1")), new(big.Int).SetUint64(uint64(c)), w[:len(w)-2])
}
