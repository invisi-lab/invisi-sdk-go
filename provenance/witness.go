package provenance

import (
	"errors"
	"fmt"
	"strconv"
)

func SupportedCapacity(k int) bool {
	return k == 10 || k == 20 || k == 40 || k == 80 || k == 120 || k == 160 || k == 190
}

type RecognizedTree struct{ leaves []string }

func (t *RecognizedTree) LeafCount() uint64 { return uint64(len(t.leaves)) }
func (t *RecognizedTree) Append(leaf string) (uint64, error) {
	n, e := Field(leaf)
	if e != nil {
		return 0, e
	}
	if n.Sign() == 0 || uint64(len(t.leaves)) >= 1<<32 {
		return 0, errors.New("invalid leaf or full tree")
	}
	t.leaves = append(t.leaves, n.String())
	return uint64(len(t.leaves) - 1), nil
}
func (t *RecognizedTree) AppendCommitment(s Scope, commitment string) (uint64, error) {
	scope, e := ScopeHash(s)
	if e != nil {
		return 0, e
	}
	leaf, e := RecognizedLeaf(scope, commitment)
	if e != nil {
		return 0, e
	}
	return t.Append(leaf)
}
func (t *RecognizedTree) Path(index uint64) (path []string, root string, err error) {
	if index >= uint64(len(t.leaves)) {
		return nil, "", errors.New("leaf index")
	}
	nodes := map[uint64]string{}
	for i, v := range t.leaves {
		nodes[uint64(i)] = v
	}
	zero := "0"
	for l := 0; l < 32; l++ {
		sibling, ok := nodes[index^1]
		if !ok {
			sibling = zero
		}
		path = append(path, sibling)
		next := map[uint64]string{}
		for i := range nodes {
			p := i / 2
			if _, ok := next[p]; !ok {
				a, ok := nodes[p*2]
				if !ok {
					a = zero
				}
				b, ok := nodes[p*2+1]
				if !ok {
					b = zero
				}
				next[p], err = hash(a, b)
				if err != nil {
					return nil, "", err
				}
			}
		}
		zero, err = hash(zero, zero)
		if err != nil {
			return nil, "", err
		}
		nodes = next
		index /= 2
	}
	root, ok := nodes[0]
	if !ok {
		root = zero
	}
	return path, root, nil
}

type WitnessInput struct {
	Opening
	SourceIndex int
	LeafIndex   uint64
	Path        []string
}
type WitnessRequest struct {
	Scope
	Capacity        int
	TokenAddress    string
	ChainGeneration string
	OperationDigest string
	Inputs          []WitnessInput
	Sources         []Source
}
type Witness map[string]any

func BuildWitness(s WitnessRequest) (Witness, error) {
	if !SupportedCapacity(s.Capacity) || len(s.Inputs) == 0 || len(s.Inputs) > s.Capacity || len(s.Sources) == 0 || len(s.Sources) > len(s.Inputs) {
		return nil, errors.New("invalid provenance shape")
	}
	scope, e := ScopeHash(s.Scope)
	if e != nil {
		return nil, e
	}
	sd, e := SourcesDigest(s.ChainGeneration, s.Sources)
	if e != nil {
		return nil, e
	}
	vectors := map[string][]string{}
	for _, k := range []string{"owners", "nonces", "amounts", "source_indices", "leaf_indices", "tree_ids", "roots", "leaf_counts"} {
		vectors[k] = make([]string, 0, s.Capacity)
	}
	for _, src := range s.Sources {
		vectors["tree_ids"] = append(vectors["tree_ids"], src.TreeID)
		vectors["roots"] = append(vectors["roots"], src.Root)
		vectors["leaf_counts"] = append(vectors["leaf_counts"], src.LeafCount)
	}
	paths := [][]string{}
	ns := []string{}
	used := map[int]bool{}
	seen := map[string]bool{}
	for _, i := range s.Inputs {
		if i.SourceIndex < 0 || i.SourceIndex >= len(s.Sources) || len(i.Path) != 32 || i.LeafIndex >= 1<<32 {
			return nil, errors.New("invalid input path")
		}
		src := s.Sources[i.SourceIndex]
		count, _ := Uint(src.LeafCount, 64)
		if i.LeafIndex >= count.Uint64() {
			return nil, errors.New("leaf outside certified prefix")
		}
		used[i.SourceIndex] = true
		c, n, e := NoteOpening(s.Scope, s.TokenAddress, i.Opening)
		if e != nil {
			return nil, e
		}
		node, e := RecognizedLeaf(scope, c)
		if e != nil {
			return nil, e
		}
		for level, p := range i.Path {
			if (i.LeafIndex>>level)&1 == 0 {
				node, e = hash(node, p)
			} else {
				node, e = hash(p, node)
			}
			if e != nil {
				return nil, e
			}
		}
		expected, _ := Field(src.Root)
		if node != expected.String() {
			return nil, errors.New("membership mismatch")
		}
		if seen[n] {
			return nil, errors.New("duplicate input")
		}
		seen[n] = true
		ns = append(ns, n)
		vectors["owners"] = append(vectors["owners"], i.Owner)
		vectors["nonces"] = append(vectors["nonces"], i.Nonce)
		vectors["amounts"] = append(vectors["amounts"], i.Amount)
		vectors["source_indices"] = append(vectors["source_indices"], strconv.Itoa(i.SourceIndex))
		vectors["leaf_indices"] = append(vectors["leaf_indices"], strconv.FormatUint(i.LeafIndex, 10))
		paths = append(paths, append([]string(nil), i.Path...))
	}
	if len(used) != len(s.Sources) {
		return nil, errors.New("unused source")
	}
	id, e := InputDigest(ns)
	if e != nil {
		return nil, e
	}
	binding, e := BindingDigest(scope, s.TokenAddress, s.ChainGeneration, strconv.Itoa(len(s.Sources)), sd, strconv.Itoa(len(s.Inputs)), s.OperationDigest, id)
	if e != nil {
		return nil, e
	}
	ohi, olo, e := limbs(s.OperationDigest)
	if e != nil {
		return nil, e
	}
	values := []string{s.ChainID, s.ProtocolAddress, s.TokenAddress, s.ChainGeneration, strconv.Itoa(len(s.Sources)), sd, strconv.Itoa(len(s.Inputs)), ohi, olo, id, binding}
	w := Witness{}
	for i, k := range publicInputNames {
		n, e := Field(values[i])
		if e != nil {
			return nil, e
		}
		w[k] = n.String()
	}
	for k, v := range vectors {
		for len(v) < s.Capacity {
			v = append(v, "0")
		}
		for i, value := range v {
			n, e := Field(value)
			if e != nil {
				return nil, fmt.Errorf("%s: %w", k, e)
			}
			v[i] = n.String()
		}
		w[k] = v
	}
	for len(paths) < s.Capacity {
		p := make([]string, 32)
		for i := range p {
			p[i] = "0"
		}
		paths = append(paths, p)
	}
	w["paths"] = paths
	return w, nil
}
func equalNumber(a, b string) bool {
	x, e := Uint(a, 256)
	if e != nil {
		return false
	}
	y, e := Uint(b, 256)
	return e == nil && x.Cmp(y) == 0
}

func (t *RecognizedTree) Root() (string, error) {
	if t.LeafCount() > 0 {
		_, root, e := t.Path(0)
		return root, e
	}
	zero := "0"
	for i := 0; i < 32; i++ {
		var e error
		zero, e = hash(zero, zero)
		if e != nil {
			return "", e
		}
	}
	return zero, nil
}

type Membership struct {
	LeafIndex uint64
	Path      []string
	Root      string
}

func (t *RecognizedTree) MembershipForCommitment(scope Scope, commitment string) (Membership, bool, error) {
	s, e := ScopeHash(scope)
	if e != nil {
		return Membership{}, false, e
	}
	expected, e := RecognizedLeaf(s, commitment)
	if e != nil {
		return Membership{}, false, e
	}
	for index, leaf := range t.leaves {
		if leaf == expected {
			path, root, e := t.Path(uint64(index))
			return Membership{uint64(index), path, root}, e == nil, e
		}
	}
	return Membership{}, false, nil
}
