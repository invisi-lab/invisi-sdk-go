package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type Execution struct {
	mu              sync.Mutex
	Runtime         *Runtime
	Plan            Plan
	Status          string
	Authorization   *core.WalletAuthorizationWitness
	Bundle          *prover.ProofBundle
	Data            []byte
	TransactionHash common.Hash
	WorkerID        string
}

func (o *Execution) ID() string { return o.Plan.ID }

func (o *Execution) String() string { return "Operation(" + o.Plan.ID + ",[private plan redacted])" }

func (o *Execution) GoString() string { return o.String() }

func (o *Execution) MarshalJSON() ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return json.Marshal(struct {
		ID     string `json:"id"`
		Kind   string `json:"kind"`
		Status string `json:"status"`
	}{o.Plan.ID, PublicOperationKind(o.Plan), o.Status})
}

func HashCode(code []byte) common.Hash { return crypto.Keccak256Hash(code) }

func ClonePlan(plan Plan) (Plan, error) {
	raw, e := json.Marshal(plan)
	if e != nil {
		return Plan{}, e
	}
	var result Plan
	e = json.Unmarshal(raw, &result)
	return result, e
}

func (o *Execution) Deliveries() []protocol.Delivery {
	if o.Plan.Kind != "transact" && o.Plan.Kind != "shield" {
		return o.Plan.Deliveries
	}
	result := []protocol.Delivery{}
	for _, out := range o.Plan.Outputs {
		result = append(result, out.Delivery)
	}
	return result
}

func (o *Execution) Words() ([][32]byte, error) {
	switch o.Plan.Kind {
	case "transact":
		return protocol.TransactUserPublicInputs(o.Plan.Context, o.Plan.Params, o.Deliveries(), o.Plan.Funding)
	case "create-credit":
		return protocol.CreateCreditPublicInputs(o.Plan.Context, *o.Plan.CreateCredit, o.Deliveries(), o.Plan.Funding)
	case "credit-draw":
		return protocol.CreditDrawPublicInputs(o.Plan.Context, *o.Plan.DrawCredit, o.Deliveries())
	case "cancel-credit":
		return protocol.CancelCreditPublicInputs(o.Plan.Context, *o.Plan.CancelCredit)
	case "top-up-credit":
		return protocol.TopUpCreditPublicInputs(o.Plan.Context, *o.Plan.TopUpCredit, o.Deliveries(), o.Plan.Funding)
	}
	return nil, errors.New("operation has no private proof")
}

func (o *Execution) AuthorizeInternal(ctx context.Context) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if e := o.CheckDrawPeriod(ctx); e != nil {
		return nil, e
	}
	if o.Status != "prepared" {
		return nil, errors.New("operation must be prepared before authorization")
	}
	if !o.Plan.NeedsAuthorization {
		return o, nil
	}
	if o.Runtime.Authorization == nil {
		return nil, errors.New("wallet authorization permission required")
	}
	words, e := o.Words()
	if e != nil {
		return nil, e
	}
	base := words[:len(words)-2]
	r, e := protocol.WalletAuthorizationRequest(o.Plan.Context, core.WalletOperation(o.Plan.Kind), o.Plan.Capacity, base)
	if e != nil {
		return nil, e
	}
	requestCopy := r
	requestCopy.Words = append([]fr.Element(nil), r.Words...)
	w, e := o.Runtime.Authorization.Authorize(ctx, requestCopy)
	if e != nil {
		return nil, e
	}
	if e = VerifyAuthorizationWitness(o.Runtime.Wallet.identity, r, w); e != nil {
		return nil, e
	}
	o.Authorization = &w
	o.Status = "authorized"
	if e = o.Persist(ctx, o.Status); e != nil {
		return nil, e
	}
	return o, nil
}

func PadStrings(values []string, size int) []string {
	for len(values) < size {
		values = append(values, "0")
	}
	return values
}

func ZeroPath() []string { return PadStrings(nil, core.TreeDepth) }

func (o *Execution) PrivateInputs() (map[string]any, error) {
	authority := core.WalletAuthorizationWitness{}
	if o.Authorization != nil {
		authority = *o.Authorization
	}
	private := map[string]any{"public_key_x": authority.PublicKeyX, "public_key_y": authority.PublicKeyY, "signature": authority.Signature}
	if o.Plan.Kind != "transact" {
		return o.CreditPrivateInputs(private)
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
	case "create-credit":
		return "create_credit"
	case "credit-draw":
		return fmt.Sprintf("credit_draw_slots%d", o.Plan.Capacity)
	case "cancel-credit":
		return "credit_cancel"
	case "top-up-credit":
		return "credit_top_up"
	}
	return ""
}

func (o *Execution) ProveInternal(ctx context.Context, options ...TransactionProveOptions) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if e := o.CheckDrawPeriod(ctx); e != nil {
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
	if execution == "aggregate" && (o.Plan.Kind != "transact" || o.Plan.Capacity != 14) {
		return nil, errors.New("aggregate execution requires a 14-slot transfer")
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

func (o *Execution) ReadyInternal(ctx context.Context) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.CheckLease(ctx); e != nil {
		return nil, e
	}
	if o.Plan.Kind != "shield" || o.Status != "prepared" {
		return nil, errors.New("Ready is for prepared shield operations")
	}
	if o.Plan.FundingMode == "signature" {
		return nil, errors.New("authorize shield funding before readiness")
	}
	if e := o.Encode(); e != nil {
		return nil, e
	}
	o.Status = "ready"
	return o, o.Persist(ctx, o.Status)
}

func (o *Execution) FundingTypedData() (apitypes.TypedData, error) {
	switch o.Plan.Kind {
	case "shield":
		return protocol.EncodeShieldFundingTypedData(o.ShieldParams())
	case "transact":
		if o.Plan.Target == "noir-recursive" {
			return protocol.EncodeAggregateFundingTypedData(o.Plan.Context, protocol.TransactRequest{Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding})
		}

		return protocol.EncodeTransactFundingTypedData(o.Plan.Context, o.Bundle.Proof, o.Plan.Params, o.Deliveries(), o.Plan.Funding)
	case "create-credit":
		return protocol.EncodeCreateCreditFundingTypedData(o.Plan.Context, *o.Plan.CreateCredit, o.Deliveries(), o.Plan.Funding)
	case "top-up-credit":
		return protocol.EncodeTopUpCreditFundingTypedData(o.Plan.Context, *o.Plan.TopUpCredit, o.Deliveries(), o.Plan.Funding)
	}
	return apitypes.TypedData{}, errors.New("operation has no public funding")
}

func (o *Execution) AuthorizeFundingInternal(ctx context.Context) (*Execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	typed, err := o.FundingRequest(ctx)
	if err != nil {
		return nil, err
	}
	signer := o.Runtime.Options.FundingSigner
	if signer == nil {
		return nil, &MissingCapabilityError{Capability: "EVM funding signer"}
	}
	if signer.Address() != o.Plan.Funding.Payer {
		return nil, errors.New("funding signer differs from payer")
	}
	copied, err := Frozen(typed)
	if err != nil {
		return nil, err
	}
	signature, err := signer.SignTypedData(ctx, copied)
	if err != nil {
		return nil, err
	}
	if err = o.AttachFunding(ctx, signature); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *Execution) ShieldParams() protocol.BuildShieldCalldataParams {
	outputs := []protocol.ShieldOutput{}
	for _, out := range o.Plan.Outputs {
		outputs = append(outputs, protocol.ShieldOutput{Identity: out.Identity, Amount: new(big.Int).Set(out.Amount)})
	}
	var caller *common.Address
	if o.Runtime.Options.Signer != nil {
		address := o.Runtime.Options.Signer.From
		caller = &address
	}
	return protocol.BuildShieldCalldataParams{Context: o.Plan.Context, Params: o.Plan.Shield, Outputs: outputs, Deliveries: o.Deliveries(), Funding: o.Plan.Funding, Caller: caller}
}

func (o *Execution) Encode() error {
	var call *protocol.Calldata
	var e error
	var caller *common.Address
	if o.Runtime.Options.Signer != nil {
		address := o.Runtime.Options.Signer.From
		caller = &address
	}
	proof := []byte(nil)
	if o.Bundle != nil {
		proof = o.Bundle.Proof
	}
	switch o.Plan.Kind {
	case "shield":
		call, e = protocol.BuildShieldCalldata(o.ShieldParams())
	case "transact":
		call, e = protocol.BuildTransactCalldata(protocol.BuildTransactCalldataParams{Context: o.Plan.Context, Proof: proof, Params: o.Plan.Params, Deliveries: o.Deliveries(), Funding: o.Plan.Funding, Caller: caller})
	default:
		call, e = o.EncodeCredit(proof, caller)
	}
	if e != nil {
		return e
	}
	o.Data = append([]byte(nil), call.Data...)
	return nil
}

type Confirmation struct {
	Status          string
	TransactionHash common.Hash
}

// Convert public statement words without reducing malformed values modulo the field.
func WitnessField(hash common.Hash) (fr.Element, error) { return core.HexToElement(hash.Hex()) }
