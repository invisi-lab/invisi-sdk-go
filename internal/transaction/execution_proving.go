package transaction

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

func PadStrings(values []string, size int) []string {
	for len(values) < size {
		values = append(values, "0")
	}
	return values
}

func ZeroPath() []string { return PadStrings(nil, primitives.TreeDepth) }

func (o *Execution) PrivateInputs() (map[string]any, error) {
	authority := primitives.WalletAuthorizationWitness{}
	if o.Authorization != nil {
		authority = *o.Authorization
	}
	private := map[string]any{"public_key_x": authority.PublicKeyX, "public_key_y": authority.PublicKeyY, "signature": authority.Signature}
	if o.Plan.Kind != "transact" {
		return nil, errors.New("Note proving supports only transact; Credit uses its independent lifecycle")
	}
	private["input_owner_address"] = "0"
	if len(o.Plan.Inputs) > 0 {
		private["input_owner_address"] = o.Plan.Inputs[0].Note.OwnerAddress.String()
	}
	identities, amounts, nonces, indices, sources := []string{}, []string{}, []string{}, []string{}, []string{}
	paths := [][]string{}
	for i, input := range o.Plan.Inputs {
		identities = append(identities, "0")
		amounts = append(amounts, input.Note.Amount.String())
		nonces = append(nonces, input.Note.Nonce.String())
		indices = append(indices, new(big.Int).SetUint64(uint64(input.LeafIndex)).String())
		p := []string{}
		for _, hash := range o.Plan.Memberships[i].Path {
			p = append(p, hash.Big().String())
		}
		paths = append(paths, p)
		index := -1
		for j, source := range o.Plan.Params.Sources {
			if source.TreeId.Uint64() == input.TreeID {
				index = j
				break
			}
		}
		if index < 0 {
			return nil, errors.New("missing input source")
		}
		sources = append(sources, big.NewInt(int64(index)).String())
	}
	for _, out := range o.Plan.Outputs {
		identities = append(identities, out.Identity.Big().String())
		amounts = append(amounts, out.Amount.String())
		nonces = append(nonces, "0")
	}
	for len(paths) < o.Plan.Capacity {
		paths = append(paths, ZeroPath())
	}
	private["output_identities"] = PadStrings(identities, o.Plan.Capacity)
	private["amounts"] = PadStrings(amounts, o.Plan.Capacity)
	private["nonces"] = PadStrings(nonces, o.Plan.Capacity)
	private["leaf_indices"] = PadStrings(indices, o.Plan.Capacity)
	private["merkle_paths"] = paths
	private["input_source_indices"] = PadStrings(sources, o.Plan.Capacity)
	return private, nil
}

func (o *Execution) CircuitName() string {
	switch o.Plan.Kind {
	case "transact":
		return fmt.Sprintf("transact_user_slots%d", o.Plan.Capacity)

	}
	return ""
}

func (o *Execution) ProveInternal(ctx context.Context, options ...TransactionProveOptions) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if len(options) > 1 {
		return nil, errors.New("expected at most one prover option")
	}
	opt := TransactionProveOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	if o.Plan.Kind == "shield" {
		return nil, errors.New("shield requires Ready, not a proof")
	}
	if o.Plan.NeedsAuthorization && o.Authorization == nil {
		return nil, errors.New("authorize the Frozen operation before proving")
	}
	if o.Status != "prepared" && o.Status != "authorized" {
		return nil, errors.New("operation is not ready for proving")
	}
	if o.Runtime.Options.Prover == nil {
		return nil, errors.New("local prover required")
	}
	execution := opt.Execution
	if execution == "" {
		execution = "single"
	}
	if execution != "single" && execution != "aggregate" {
		return nil, errors.New("execution must be single or aggregate")
	}
	if execution == "aggregate" && o.Plan.Broadcaster != nil {
		return nil, errors.New("broadcaster operations cannot target aggregate proofs")
	}
	if execution == "aggregate" && (o.Plan.Kind != "transact" || (o.Plan.Capacity < 1 || o.Plan.Capacity > 10)) {
		return nil, errors.New("aggregate execution requires a native C1 through C10 transfer")
	}
	o.Plan.Target = "evm"
	if execution == "aggregate" {
		o.Plan.Target = "noir-recursive"
	}
	words, e := o.Words()
	if e != nil {
		return nil, e
	}
	public := make([]common.Hash, len(words))
	for i, w := range words {
		public[i] = w
	}
	private, e := o.PrivateInputs()
	if e != nil {
		return nil, e
	}
	request, e := prover.Request(o.CircuitName(), o.Plan.Target, private, public)
	if e != nil {
		return nil, e
	}
	expected := prover.CloneStatement(request.ProofStatement)
	bundle, e := o.Runtime.Options.Prover.Prove(ctx, request, opt.Options)
	if e != nil {
		return nil, e
	}
	bundle = prover.CloneBundle(bundle)
	if e = prover.AssertProofBundle(bundle, expected); e != nil {
		return nil, e
	}
	valid, e := o.Runtime.Options.Prover.Verify(ctx, prover.CloneBundle(bundle), prover.CloneStatement(expected), opt.Options)
	if e != nil {
		return nil, e
	}
	if !valid {
		return nil, errors.New("local proof verification failed")
	}
	o.Bundle = &bundle
	o.Authorization = nil
	o.Status = "proved"
	if o.Plan.FundingMode != "signature" {
		if o.Plan.Target != "noir-recursive" {
			if e = o.Encode(); e != nil {
				return nil, e
			}
		}
		o.Status = "ready"
	}
	if e = o.Persist(ctx, o.Status); e != nil {
		return nil, e
	}
	return o, nil
}

// Convert public statement words without reducing malformed values modulo the field.
func WitnessField(hash common.Hash) (fr.Element, error) { return primitives.HexToElement(hash.Hex()) }
