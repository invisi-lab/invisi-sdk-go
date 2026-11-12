package credit

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type Authority struct{ PublicKey, Signature []byte }

func (a Authority) String() string               { return "CreditAuthority([redacted])" }
func (a Authority) GoString() string             { return a.String() }
func (a Authority) MarshalJSON() ([]byte, error) { return []byte(`{"authority":"[redacted]"}`), nil }
func authWitness(identity fr.Element, typed apitypes.TypedData, auth Authority) (map[string]any, error) {
	if len(auth.PublicKey) != 65 || len(auth.Signature) != 64 {
		return nil, fmt.Errorf("Credit authority requires uncompressed public key and raw64 signature")
	}
	public, err := crypto.UnmarshalPubkey(auth.PublicKey)
	if err != nil {
		return nil, err
	}
	if crypto.PubkeyToAddress(*public).Big().Cmp(primitives.ElementToBigInt(identity)) != 0 {
		return nil, fmt.Errorf("Credit authority identity mismatch")
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		return nil, err
	}
	if !crypto.VerifySignature(auth.PublicKey, digest, auth.Signature) {
		return nil, fmt.Errorf("invalid Credit wallet authorization")
	}
	asNumbers := func(raw []byte) []int {
		out := make([]int, len(raw))
		for i, b := range raw {
			out[i] = int(b)
		}
		return out
	}
	return map[string]any{"public_key_x": asNumbers(auth.PublicKey[1:33]), "public_key_y": asNumbers(auth.PublicKey[33:65]), "signature": asNumbers(auth.Signature)}, nil
}
func stringFields(values []fr.Element) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = v.String()
	}
	return out
}
func makeWitness(private map[string]any, names []string, inputs []common.Hash) (*prover.LocalWitness, error) {
	if len(names) != len(inputs) {
		return nil, fmt.Errorf("Credit witness public width mismatch")
	}
	for i, name := range names {
		private[name] = inputs[i].Big().String()
	}
	return prover.NewLocalWitness(private)
}

var createNames = []string{"chain_id", "protocol_address", "token_address", "owner_address", "revocation_shard", "revocation_slot", "credit_commitment", "deadline", "delivery_hash_high", "delivery_hash_low", "fee_recipient", "fee_amount", "auth_digest_high", "auth_digest_low"}
var updateNames = []string{"chain_id", "protocol_address", "token_address", "owner_address", "input_tree_id", "merkle_root", "nullifier", "credit_commitment", "revocation_shard", "revocation_slot", "deadline", "delivery_hash_high", "delivery_hash_low", "fee_recipient", "fee_amount", "auth_digest_high", "auth_digest_low"}

func (p CreateParams) Witness(c Context, deliveries []protocol.Delivery, state State, auth Authority) (*prover.LocalWitness, error) {
	inputs, typed, err := p.Statement(c, deliveries)
	if err != nil {
		return nil, err
	}
	commitment, err := state.Commitment()
	if err != nil {
		return nil, err
	}
	if state.ProtocolContext != c.ProtocolContext || primitives.ElementToBigInt(state.TokenAddress).Cmp(c.TokenAddress.Big()) != 0 || primitives.ElementToBigInt(state.OwnerWalletIdentity).Cmp(p.Owner.Big()) != 0 || primitives.ElementToBigInt(state.RevocationShard).Uint64() != p.RevocationShard || primitives.ElementToBigInt(state.RevocationSlot).Uint64() != uint64(p.RevocationSlot) || common.Hash(commitment.Bytes()) != p.CreditCommitment {
		return nil, fmt.Errorf("created Credit opening mismatch")
	}
	for _, v := range state.Usage.Fields() {
		if !v.IsZero() {
			return nil, fmt.Errorf("created grant must have zero usage")
		}
	}
	if primitives.ElementToBigInt(state.Expiry).Cmp(p.Deadline) < 0 {
		return nil, fmt.Errorf("grant expires before deadline")
	}
	private, err := authWitness(state.OwnerWalletIdentity, typed, auth)
	if err != nil {
		return nil, err
	}
	private["spender_wallet_identity"] = state.SpenderWalletIdentity.String()
	private["credit_nonce"] = state.Nonce.String()
	private["expiry"] = state.Expiry.String()
	private["credit_shared_key"] = state.CreditSharedKey.String()
	private["limits"] = stringFields(state.Policy.Fields())
	private["available_credit"] = state.AvailableCredit.String()
	private["unlimited"] = boolString(state.Unlimited)
	return makeWitness(private, createNames, inputs)
}

type LocatedState struct {
	State     State
	TreeID    uint64
	LeafIndex uint16
	Path      [16]common.Hash
	Debit     fr.Element
}

func (s LocatedState) root() (common.Hash, error) {
	if s.TreeID == 0 {
		return common.Hash{}, fmt.Errorf("zero Credit tree")
	}
	leaf, err := s.State.Commitment()
	if err != nil {
		return common.Hash{}, err
	}
	for i, h := range s.Path {
		v, err := primitives.HexToElement(h.Hex())
		if err != nil {
			return common.Hash{}, err
		}
		if (s.LeafIndex>>i)&1 == 0 {
			leaf, err = primitives.Poseidon2Hash(leaf, v)
		} else {
			leaf, err = primitives.Poseidon2Hash(v, leaf)
		}
		if err != nil {
			return common.Hash{}, err
		}
	}
	return common.Hash(leaf.Bytes()), nil
}
func (s LocatedState) pathFields() []string {
	values := make([]string, 16)
	for i, v := range s.Path {
		values[i] = v.Big().String()
	}
	return values
}
func (p UpdateParams) Witness(c Context, deliveries []protocol.Delivery, input LocatedState, change Change, auth Authority) (*prover.LocalWitness, error) {
	inputs, typed, err := p.Statement(c, deliveries)
	if err != nil {
		return nil, err
	}
	s := input.State
	if s.ProtocolContext != c.ProtocolContext || primitives.ElementToBigInt(s.TokenAddress).Cmp(p.Token.Big()) != 0 || primitives.ElementToBigInt(s.OwnerWalletIdentity).Cmp(p.Owner.Big()) != 0 || primitives.ElementToBigInt(s.RevocationShard).Uint64() != p.RevocationShard || primitives.ElementToBigInt(s.RevocationSlot).Uint64() != uint64(p.RevocationSlot) {
		return nil, fmt.Errorf("updated grant namespace mismatch")
	}
	next, err := s.Update(change, p.Deadline.Uint64())
	if err != nil {
		return nil, err
	}
	if common.Hash(next.Nullifier.Bytes()) != p.Nullifier || common.Hash(next.StateCommitment.Bytes()) != p.CreditCommitment {
		return nil, fmt.Errorf("updated grant transition mismatch")
	}
	root, err := input.root()
	if err != nil {
		return nil, err
	}
	if p.Source.TreeId.Uint64() != input.TreeID || root != common.Hash(p.Source.MerkleRoot) {
		return nil, fmt.Errorf("invalid Credit membership")
	}
	private, err := authWitness(s.OwnerWalletIdentity, typed, auth)
	if err != nil {
		return nil, err
	}
	for k, v := range map[string]fr.Element{"spender_wallet_identity": s.SpenderWalletIdentity, "input_nonce": s.Nonce, "old_expiry": s.Expiry, "new_expiry": change.Expiry, "credit_shared_key": s.CreditSharedKey} {
		private[k] = v.String()
	}
	private["old_available_credit"] = s.AvailableCredit.String()
	private["old_unlimited"] = boolString(s.Unlimited)
	private["new_unlimited"] = boolString(next.State.Unlimited)
	private["increase_available_credit"] = change.IncreaseAvailableCredit.String()
	private["decrease_available_credit"] = change.DecreaseAvailableCredit.String()
	private["old_limits"] = stringFields(s.Policy.Fields())
	private["new_limits"] = stringFields(change.Policy.Fields())
	private["usage"] = stringFields(s.Usage.Fields())
	private["leaf_index"] = new(big.Int).SetUint64(uint64(input.LeafIndex)).String()
	private["merkle_path"] = input.pathFields()
	return makeWitness(private, updateNames, inputs)
}
func (p DrawParams) Witness(c Context, deliveries []protocol.Delivery, credits []LocatedState, outputIdentity fr.Element, auth Authority) (*prover.LocalWitness, error) {
	inputs, typed, err := p.Statement(c, deliveries)
	if err != nil {
		return nil, err
	}
	if len(credits) != len(p.Nullifiers) {
		return nil, fmt.Errorf("Credit input count mismatch")
	}
	capacity, _ := DrawCapacity(len(credits))
	identity := credits[0].State.SpenderWalletIdentity
	private, err := authWitness(identity, typed, auth)
	if err != nil {
		return nil, err
	}
	private["spender_wallet_identity"] = identity.String()
	arrays := map[string][]string{}
	for _, name := range []string{"input_nonces", "expiries", "credit_shared_keys", "leaf_indices", "input_source_indices", "draw_amounts", "available_credits"} {
		arrays[name] = make([]string, capacity)
		for i := range arrays[name] {
			arrays[name][i] = "0"
		}
	}
	unlimitedFlags := make([]string, capacity)
	for i := range unlimitedFlags {
		unlimitedFlags[i] = "0"
	}
	limits, usages, paths := make([][]string, capacity), make([][]string, capacity), make([][]string, capacity)
	for i := 0; i < capacity; i++ {
		limits[i] = []string{"0", "0", "0"}
		usages[i] = []string{"0", "0", "0", "0", "0"}
		paths[i] = make([]string, 16)
		for j := range paths[i] {
			paths[i][j] = "0"
		}
	}
	sources, payers := map[int]bool{}, map[int]bool{}
	totals := make([]*big.Int, len(p.Payers))
	for i := range totals {
		totals[i] = new(big.Int)
	}
	for i, input := range credits {
		s := input.State
		if s.ProtocolContext != c.ProtocolContext || primitives.ElementToBigInt(s.TokenAddress).Cmp(p.Token.Big()) != 0 || s.SpenderWalletIdentity != identity || primitives.ElementToBigInt(s.Expiry).Cmp(p.Deadline) < 0 {
			return nil, fmt.Errorf("Credit context, spender or expiry mismatch")
		}
		si := -1
		for j, source := range p.Sources {
			if source.TreeId.Uint64() == input.TreeID {
				si = j
				break
			}
		}
		if si < 0 {
			return nil, fmt.Errorf("Credit source missing")
		}
		pi := int(p.GrantRefs[i].PayerIndex.Uint64())
		if primitives.ElementToBigInt(s.OwnerWalletIdentity).Cmp(p.Payers[pi].Payer.Big()) != 0 || primitives.ElementToBigInt(s.RevocationShard).Uint64() != p.GrantRefs[i].Shard || primitives.ElementToBigInt(s.RevocationSlot).Uint64() != uint64(p.GrantRefs[i].Slot) {
			return nil, fmt.Errorf("Credit payer or namespace mismatch")
		}
		root, err := input.root()
		if err != nil {
			return nil, err
		}
		if root != common.Hash(p.Sources[si].MerkleRoot) {
			return nil, fmt.Errorf("invalid Credit membership")
		}
		next, err := s.Draw(input.Debit, c.Timestamp)
		if err != nil {
			return nil, err
		}
		if common.Hash(next.Nullifier.Bytes()) != p.Nullifiers[i] || common.Hash(next.StateCommitment.Bytes()) != p.CreditCommitments[i] || common.Hash(next.RecoveryAmount.Bytes()) != p.RecoveryAmounts[i] {
			return nil, fmt.Errorf("Credit draw transition mismatch")
		}
		arrays["input_nonces"][i] = s.Nonce.String()
		arrays["expiries"][i] = s.Expiry.String()
		arrays["credit_shared_keys"][i] = s.CreditSharedKey.String()
		arrays["leaf_indices"][i] = new(big.Int).SetUint64(uint64(input.LeafIndex)).String()
		arrays["input_source_indices"][i] = big.NewInt(int64(si)).String()
		arrays["draw_amounts"][i] = input.Debit.String()
		arrays["available_credits"][i] = s.AvailableCredit.String()
		unlimitedFlags[i] = boolString(s.Unlimited)
		limits[i] = stringFields(s.Policy.Fields())
		usages[i] = stringFields(s.Usage.Fields())
		paths[i] = input.pathFields()
		sources[si] = true
		payers[pi] = true
		totals[pi].Add(totals[pi], primitives.ElementToBigInt(input.Debit))
	}
	if len(sources) != len(p.Sources) || len(payers) != len(p.Payers) {
		return nil, fmt.Errorf("unused Credit public group")
	}
	for i, total := range totals {
		if total.Cmp(p.Payers[i].Amount) != 0 {
			return nil, fmt.Errorf("payer receipt mismatch")
		}
	}
	if outputIdentity.IsZero() {
		return nil, fmt.Errorf("zero output Note identity")
	}
	fee, _ := normalizeFee(p.FeeRecipient, p.FeeAmount)
	net := new(big.Int).Sub(p.TotalAmount, fee)
	out, err := primitives.Poseidon2Hash(primitives.ElementFromUint64(3), c.ChainID, c.ProtocolAddress, outputIdentity, primitives.BigIntToElement(p.Token.Big()), primitives.BigIntToElement(net))
	if err != nil {
		return nil, err
	}
	if common.Hash(out.Bytes()) != p.NoteCommitment {
		return nil, fmt.Errorf("ordinary Note output mismatch")
	}
	private["output_identity"] = outputIdentity.String()
	for k, a := range arrays {
		private[k] = a
	}
	private["unlimited_flags"] = unlimitedFlags
	private["limits"] = limits
	private["usages"] = usages
	private["merkle_paths"] = paths
	names := []string{"chain_id", "protocol_address", "token_address", "actual_count", "source_count", "payer_count"}
	for i, name := range names {
		private[name] = inputs[i].Big().String()
	}
	at := 6
	for _, name := range []string{"tree_ids", "roots", "payer_addresses", "payer_amounts", "grant_payer_indices", "grant_shards", "grant_slots", "nullifiers", "state_commitments", "recovery_amounts"} {
		row := make([]string, capacity)
		for i := range row {
			row[i] = inputs[at+i].Big().String()
		}
		private[name] = row
		at += capacity
	}
	return makeWitness(private, []string{"note_commitment", "total_amount", "deadline", "delivery_hash_high", "delivery_hash_low", "current_day", "current_month", "fee_recipient", "fee_amount", "auth_digest_high", "auth_digest_low"}, inputs[at:])
}

func boolString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
