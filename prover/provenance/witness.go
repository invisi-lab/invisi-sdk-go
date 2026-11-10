package provenance

import (
	"errors"
	"fmt"
	"strconv"
)

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
