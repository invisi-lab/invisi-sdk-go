package provenance

import (
	"errors"
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
