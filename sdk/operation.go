package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type Operation struct {
	mu              sync.Mutex
	session         *WalletSession
	plan            operationPlan
	status          string
	authorization   *core.WalletAuthorizationWitness
	bundle          *prover.ProofBundle
	data            []byte
	transactionHash common.Hash
	workerID        string
}

func (o *Operation) ID() string       { return o.plan.ID }
func (o *Operation) String() string   { return "Operation(" + o.plan.ID + ",[private plan redacted])" }
func (o *Operation) GoString() string { return o.String() }
func (o *Operation) MarshalJSON() ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return json.Marshal(struct {
		ID     string `json:"id"`
		Kind   string `json:"kind"`
		Status string `json:"status"`
	}{o.plan.ID, publicOperationKind(o.plan), o.status})
}
func hashCode(code []byte) common.Hash { return crypto.Keccak256Hash(code) }
func clonePlan(plan operationPlan) (operationPlan, error) {
	raw, e := json.Marshal(plan)
	if e != nil {
		return operationPlan{}, e
	}
	var result operationPlan
	e = json.Unmarshal(raw, &result)
	return result, e
}
func (s *WalletSession) persistPlan(ctx context.Context, plan operationPlan) (*Operation, error) {

	if e := s.canonical(ctx, plan.Block); e != nil {
		return nil, e
	}
	id, e := randomOperationID()
	if e != nil {
		return nil, e
	}
	plan.ID = id
	plan.Version = 1
	plan.WalletIdentity = s.wallet.identity
	plan.Namespace = s.scope.Namespace
	plan, e = clonePlan(plan)
	if e != nil {
		return nil, e
	}
	operation := &Operation{session: s, plan: plan, status: "prepared", workerID: id}
	encoded, e := json.Marshal(plan)
	if e != nil {
		return nil, e
	}
	if s.explicitInventory != nil {
		return operation, nil
	}
	e = s.options.Store.Update(ctx, s.scope, func(state *WalletState) error {
		if state.Invalidated {
			return errors.New("wallet requires reset refresh")
		}
		for _, n := range plan.Inputs {
			current, ok := state.Notes[n.Commitment]
			if !ok || current.Spent || current.ReservedBy != "" {
				return errors.New("input already spent or reserved")
			}
			if !current.Note.Equal(n.Note) || current.TreeID != n.TreeID || current.LeafIndex != n.LeafIndex {
				return errors.New("input changed during preparation")
			}
		}
		for _, c := range plan.CreditInputs {
			for _, existing := range state.Credits {
				if existing.Commitment == c.Commitment && (existing.Spent || existing.ReservedBy != "") {
					return errors.New("Credit already spent or reserved")
				}
			}
			current, ok := state.Credits[creditKey(c.Commitment, c.Role)]
			if !ok || current.Spent || current.ReservedBy != "" {
				return errors.New("Credit already spent or reserved")
			}
		}
		record := OperationRecord{ID: id, Kind: publicOperationKind(plan), Status: "prepared", Deadline: plan.Deadline, Plan: encoded, WorkerID: id}
		for _, n := range plan.Inputs {
			current := state.Notes[n.Commitment]
			current.ReservedBy = id
			state.Notes[n.Commitment] = current
			record.Inputs = append(record.Inputs, n.Commitment)
		}
		for _, c := range plan.CreditInputs {
			for key, current := range state.Credits {
				if current.Commitment == c.Commitment {
					current.ReservedBy = id
					state.Credits[key] = current
				}
			}
		}
		state.Operations[id] = record
		return nil
	})
	return operation, e
}
func (o *Operation) persist(ctx context.Context, status string) error {
	if o.session.explicitInventory != nil {
		return ctx.Err()
	}
	encoded, e := json.Marshal(o.plan)
	if e != nil {
		return e
	}
	return o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		record, ok := state.Operations[o.plan.ID]
		if !ok {
			return errors.New("operation record missing")
		}
		if record.WorkerID != o.workerID {
			return errors.New("operation worker was replaced; resume operation")
		}
		record.Status = status
		if status == "uncertain" {
			record.Externalized = true
		}
		record.TransactionHash = o.transactionHash
		record.CallData = append([]byte(nil), o.data...)
		record.Plan = encoded
		if o.bundle != nil {
			raw, e := json.Marshal(o.bundle)
			if e != nil {
				return e
			}
			record.Proof = raw
		}
		state.Operations[o.plan.ID] = record
		return nil
	})
}
func (o *Operation) deliveries() []protocol.Delivery {
	if o.plan.Kind != "transact" && o.plan.Kind != "shield" {
		return o.plan.Deliveries
	}
	result := []protocol.Delivery{}
	for _, out := range o.plan.Outputs {
		result = append(result, out.Delivery)
	}
	return result
}
func (o *Operation) words() ([][32]byte, error) {
	switch o.plan.Kind {
	case "transact":
		return protocol.TransactUserPublicInputs(o.plan.Context, o.plan.Params, o.deliveries(), o.plan.Funding)
	case "create-credit":
		return protocol.CreateCreditPublicInputs(o.plan.Context, *o.plan.CreateCredit, o.deliveries(), o.plan.Funding)
	case "credit-draw":
		return protocol.CreditDrawPublicInputs(o.plan.Context, *o.plan.DrawCredit, o.deliveries())
	case "cancel-credit":
		return protocol.CancelCreditPublicInputs(o.plan.Context, *o.plan.CancelCredit)
	case "top-up-credit":
		return protocol.TopUpCreditPublicInputs(o.plan.Context, *o.plan.TopUpCredit, o.deliveries(), o.plan.Funding)
	}
	return nil, errors.New("operation has no private proof")
}
func (o *Operation) authorizeInternal(ctx context.Context) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return nil, e
	}
	if e := o.checkDrawPeriod(ctx); e != nil {
		return nil, e
	}
	if o.status != "prepared" {
		return nil, errors.New("operation must be prepared before authorization")
	}
	if !o.plan.NeedsAuthorization {
		return o, nil
	}
	if o.session.authorization == nil {
		return nil, errors.New("wallet authorization permission required")
	}
	words, e := o.words()
	if e != nil {
		return nil, e
	}
	base := words[:len(words)-2]
	r, e := protocol.WalletAuthorizationRequest(o.plan.Context, core.WalletOperation(o.plan.Kind), o.plan.Capacity, base)
	if e != nil {
		return nil, e
	}
	requestCopy := r
	requestCopy.Words = append([]fr.Element(nil), r.Words...)
	w, e := o.session.authorization.Authorize(ctx, requestCopy)
	if e != nil {
		return nil, e
	}
	if e = verifyAuthorizationWitness(o.session.wallet.identity, r, w); e != nil {
		return nil, e
	}
	o.authorization = &w
	o.status = "authorized"
	if e = o.persist(ctx, o.status); e != nil {
		return nil, e
	}
	return o, nil
}
func padStrings(values []string, size int) []string {
	for len(values) < size {
		values = append(values, "0")
	}
	return values
}
func zeroPath() []string { return padStrings(nil, core.TreeDepth) }
func (o *Operation) privateInputs() (map[string]any, error) {
	authority := core.WalletAuthorizationWitness{}
	if o.authorization != nil {
		authority = *o.authorization
	}
	private := map[string]any{"public_key_x": authority.PublicKeyX, "public_key_y": authority.PublicKeyY, "signature": authority.Signature}
	if o.plan.Kind != "transact" {
		return o.creditPrivateInputs(private)
	}
	private["wallet_identity"] = "0"
	if len(o.plan.Inputs) > 0 {
		private["wallet_identity"] = o.plan.Inputs[0].Note.WalletIdentity.String()
	}
	identities, amounts, nonces, indices, sources := []string{}, []string{}, []string{}, []string{}, []string{}
	paths := [][]string{}
	for i, input := range o.plan.Inputs {
		identities = append(identities, "0")
		amounts = append(amounts, input.Note.Amount.String())
		nonces = append(nonces, input.Note.Nonce.String())
		indices = append(indices, new(big.Int).SetUint64(uint64(input.LeafIndex)).String())
		p := []string{}
		for _, hash := range o.plan.Memberships[i].Path {
			p = append(p, hash.Big().String())
		}
		paths = append(paths, p)
		index := -1
		for j, source := range o.plan.Params.Sources {
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
	for _, out := range o.plan.Outputs {
		identities = append(identities, out.Note.WalletIdentity.String())
		amounts = append(amounts, out.Note.Amount.String())
		nonces = append(nonces, out.Note.Nonce.String())
	}
	for len(paths) < o.plan.Capacity {
		paths = append(paths, zeroPath())
	}
	private["output_wallet_identities"] = padStrings(identities, o.plan.Capacity)
	private["amounts"] = padStrings(amounts, o.plan.Capacity)
	private["nonces"] = padStrings(nonces, o.plan.Capacity)
	private["leaf_indices"] = padStrings(indices, o.plan.Capacity)
	private["merkle_paths"] = paths
	private["input_source_indices"] = padStrings(sources, o.plan.Capacity)
	return private, nil
}
func (o *Operation) circuitName() string {
	switch o.plan.Kind {
	case "transact":
		return fmt.Sprintf("transact_user_slots%d", o.plan.Capacity)
	case "create-credit":
		return "create_credit"
	case "credit-draw":
		return fmt.Sprintf("credit_draw_slots%d", o.plan.Capacity)
	case "cancel-credit":
		return "credit_cancel"
	case "top-up-credit":
		return "credit_top_up"
	}
	return ""
}
func (o *Operation) proveInternal(ctx context.Context, options ...TransactionProveOptions) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return nil, e
	}
	if e := o.checkDrawPeriod(ctx); e != nil {
		return nil, e
	}
	if len(options) > 1 {
		return nil, errors.New("expected at most one prover option")
	}
	opt := TransactionProveOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	if o.plan.Kind == "shield" {
		return nil, errors.New("shield requires Ready, not a proof")
	}
	if o.plan.NeedsAuthorization && o.authorization == nil {
		return nil, errors.New("authorize the frozen operation before proving")
	}
	if o.status != "prepared" && o.status != "authorized" {
		return nil, errors.New("operation is not ready for proving")
	}
	if o.session.options.Prover == nil {
		return nil, errors.New("local prover required")
	}
	execution := opt.Execution
	if execution == "" {
		execution = "single"
	}
	if execution != "single" && execution != "aggregate" {
		return nil, errors.New("execution must be single or aggregate")
	}
	if execution == "aggregate" && (o.plan.Kind != "transact" || o.plan.Capacity != 14) {
		return nil, errors.New("aggregate execution requires a 14-slot transfer")
	}
	o.plan.Target = "evm"
	if execution == "aggregate" {
		o.plan.Target = "noir-recursive"
	}
	words, e := o.words()
	if e != nil {
		return nil, e
	}
	public := make([]common.Hash, len(words))
	for i, w := range words {
		public[i] = w
	}
	private, e := o.privateInputs()
	if e != nil {
		return nil, e
	}
	request, e := prover.Request(o.circuitName(), o.plan.Target, private, public)
	if e != nil {
		return nil, e
	}
	expected := prover.CloneStatement(request.ProofStatement)
	bundle, e := o.session.options.Prover.Prove(ctx, request, opt.Options)
	if e != nil {
		return nil, e
	}
	bundle = prover.CloneBundle(bundle)
	if e = prover.AssertProofBundle(bundle, expected); e != nil {
		return nil, e
	}
	valid, e := o.session.options.Prover.Verify(ctx, prover.CloneBundle(bundle), prover.CloneStatement(expected), opt.Options)
	if e != nil {
		return nil, e
	}
	if !valid {
		return nil, errors.New("local proof verification failed")
	}
	o.bundle = &bundle
	o.authorization = nil
	o.status = "proved"
	if o.plan.FundingMode != "signature" {
		if o.plan.Target != "noir-recursive" {
			if e = o.encode(); e != nil {
				return nil, e
			}
		}
		o.status = "ready"
	}
	if e = o.persist(ctx, o.status); e != nil {
		return nil, e
	}
	return o, nil
}
func (o *Operation) readyInternal(ctx context.Context) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return nil, e
	}
	if o.plan.Kind != "shield" || o.status != "prepared" {
		return nil, errors.New("Ready is for prepared shield operations")
	}
	if o.plan.FundingMode == "signature" {
		return nil, errors.New("authorize shield funding before readiness")
	}
	if e := o.encode(); e != nil {
		return nil, e
	}
	o.status = "ready"
	return o, o.persist(ctx, o.status)
}
func (o *Operation) fundingTypedData() (apitypes.TypedData, error) {
	switch o.plan.Kind {
	case "shield":
		return protocol.EncodeShieldFundingTypedData(o.shieldParams())
	case "transact":
		if o.plan.Target == "noir-recursive" {
			return protocol.EncodeAggregateFundingTypedData(o.plan.Context, protocol.TransactRequest{Params: o.plan.Params, Deliveries: o.deliveries(), Funding: o.plan.Funding})
		}

		return protocol.EncodeTransactFundingTypedData(o.plan.Context, o.bundle.Proof, o.plan.Params, o.deliveries(), o.plan.Funding)
	case "create-credit":
		return protocol.EncodeCreateCreditFundingTypedData(o.plan.Context, *o.plan.CreateCredit, o.deliveries(), o.plan.Funding)
	case "top-up-credit":
		return protocol.EncodeTopUpCreditFundingTypedData(o.plan.Context, *o.plan.TopUpCredit, o.deliveries(), o.plan.Funding)
	}
	return apitypes.TypedData{}, errors.New("operation has no public funding")
}
func (o *Operation) authorizeFundingInternal(ctx context.Context) (*Operation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	typed, err := o.fundingRequest(ctx)
	if err != nil {
		return nil, err
	}
	signer := o.session.options.FundingSigner
	if signer == nil {
		return nil, &MissingCapabilityError{Capability: "EVM funding signer"}
	}
	if signer.Address() != o.plan.Funding.Payer {
		return nil, errors.New("funding signer differs from payer")
	}
	copied, err := frozen(typed)
	if err != nil {
		return nil, err
	}
	signature, err := signer.SignTypedData(ctx, copied)
	if err != nil {
		return nil, err
	}
	if err = o.attachFunding(ctx, signature); err != nil {
		return nil, err
	}
	return o, nil
}
func (o *Operation) shieldParams() protocol.BuildShieldCalldataParams {
	outputs := []protocol.ShieldOutput{}
	for _, out := range o.plan.Outputs {
		outputs = append(outputs, protocol.ShieldOutput{Identity: out.Identity, Amount: core.ElementToBigInt(out.Note.Amount)})
	}
	var caller *common.Address
	if o.session.options.Signer != nil {
		address := o.session.options.Signer.From
		caller = &address
	}
	return protocol.BuildShieldCalldataParams{Context: o.plan.Context, Params: o.plan.Shield, Outputs: outputs, Deliveries: o.deliveries(), Funding: o.plan.Funding, Caller: caller}
}
func (o *Operation) encode() error {
	var call *protocol.Calldata
	var e error
	var caller *common.Address
	if o.session.options.Signer != nil {
		address := o.session.options.Signer.From
		caller = &address
	}
	proof := []byte(nil)
	if o.bundle != nil {
		proof = o.bundle.Proof
	}
	switch o.plan.Kind {
	case "shield":
		call, e = protocol.BuildShieldCalldata(o.shieldParams())
	case "transact":
		call, e = protocol.BuildTransactCalldata(protocol.BuildTransactCalldataParams{Context: o.plan.Context, Proof: proof, Params: o.plan.Params, Deliveries: o.deliveries(), Funding: o.plan.Funding, Caller: caller})
	default:
		call, e = o.encodeCredit(proof, caller)
	}
	if e != nil {
		return e
	}
	o.data = append([]byte(nil), call.Data...)
	return nil
}

type operationConfirmation struct {
	Status          string
	TransactionHash common.Hash
}

type ReconcileOptions struct{ TransactionHash common.Hash }

func (o *Operation) reconcileInternal(ctx context.Context, options ...ReconcileOptions) (operationConfirmation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return operationConfirmation{}, e
	}
	if len(options) > 1 {
		return operationConfirmation{}, errors.New("expected at most one reconciliation option")
	}
	if len(options) == 1 && options[0].TransactionHash != (common.Hash{}) {
		o.transactionHash = options[0].TransactionHash
	}

	if e := o.session.protocol.ValidateDeployment(ctx, o.session.options.Provider); e != nil {
		return operationConfirmation{}, e
	}

	result := operationConfirmation{Status: o.status, TransactionHash: o.transactionHash}
	if o.transactionHash == (common.Hash{}) {
		return o.reconcileSpendState(ctx)
	}
	receipt, e := o.session.options.Provider.TransactionReceipt(ctx, o.transactionHash)
	if errors.Is(e, ethereum.NotFound) {
		return o.reconcileSpendState(ctx)
	}
	if e != nil {
		return result, e
	}
	if receipt == nil {
		return o.reconcileSpendState(ctx)
	}
	if receipt.TxHash != o.transactionHash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		return result, errors.New("receipt does not identify the submitted transaction and block")
	}
	header, e := o.session.options.Provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return result, e
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return result, errors.New("transaction receipt is not canonical")
	}
	reader, ok := o.session.options.Provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return result, errors.New("provider must authenticate submitted transaction calldata")
	}
	tx, pending, e := reader.TransactionByHash(ctx, o.transactionHash)
	if e != nil {
		return result, e
	}
	if pending || tx == nil || tx.Hash() != o.transactionHash || tx.To() == nil || *tx.To() != o.plan.Context.ProtocolAddress || !o.matchesTransactionData(tx.Data()) {
		return result, errors.New("submitted transaction differs from frozen operation")
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return o.reconcileSpendState(ctx)
	}
	receiptBlock := protocolrpc.CanonicalBlock{Number: header.Number.Uint64(), Hash: header.Hash(), ParentHash: header.ParentHash, Timestamp: header.Time}
	latest, e := o.session.options.Provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return result, e
	}
	if latest == nil || latest.Number == nil || !latest.Number.IsUint64() || latest.Number.Cmp(header.Number) < 0 {
		return result, errors.New("latest block predates receipt")
	}
	block := protocolrpc.CanonicalBlock{Number: latest.Number.Uint64(), Hash: latest.Hash(), ParentHash: latest.ParentHash, Timestamp: latest.Time}
	spent, e := o.spentAt(ctx, latest.Number)
	if e != nil {
		return result, e
	}
	for _, value := range spent {
		if !value {
			return result, errors.New("successful transaction has unspent input")
		}
	}
	if e = o.session.canonical(ctx, receiptBlock); e != nil {
		return result, e
	}
	if e = o.session.canonical(ctx, block); e != nil {
		return result, e
	}
	e = o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		record, ok := state.Operations[o.plan.ID]
		if !ok || state.Invalidated || record.WorkerID != o.workerID {
			return errors.New("stale operation worker during confirmation")
		}
		if e := checkObservation(state, block); e != nil {
			return e
		}
		for _, input := range o.plan.Inputs {
			if note, ok := state.Notes[input.Commitment]; ok {
				if note.ReservedBy != "" && note.ReservedBy != o.plan.ID {
					return errors.New("input reserved by another operation")
				}
				note.Spent = true
				note.ReservedBy = ""
				state.Notes[input.Commitment] = note
			}
		}
		for _, input := range o.plan.CreditInputs {
			for key, credit := range state.Credits {
				if credit.Commitment == input.Commitment {
					if credit.ReservedBy != "" && credit.ReservedBy != o.plan.ID {
						return errors.New("Credit reserved by another operation")
					}
					credit.Spent = true
					credit.ReservedBy = ""
					state.Credits[key] = credit
				}
			}
		}
		record.Status = "confirmed"
		record.TransactionHash = o.transactionHash
		record.ResolvedAt = &block
		state.Operations[o.plan.ID] = record
		state.CheckedThrough = &block
		return nil
	})
	if e != nil {
		return result, e
	}
	o.status = "confirmed"
	result.Status = "confirmed"
	return result, nil
}

func (o *Operation) Abandon(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if e := o.checkLease(ctx); e != nil {
		return e
	}
	if o.status != "prepared" {
		return errors.New("only an unexternalized prepared operation can be abandoned")
	}
	e := o.session.options.Store.Update(ctx, o.session.scope, func(state *WalletState) error {
		if record, ok := state.Operations[o.plan.ID]; !ok || record.WorkerID != o.workerID {
			return errors.New("operation worker replaced")
		}
		for hash, n := range state.Notes {
			if n.ReservedBy == o.plan.ID {
				n.ReservedBy = ""
				state.Notes[hash] = n
			}
		}
		for key, c := range state.Credits {
			if c.ReservedBy == o.plan.ID {
				c.ReservedBy = ""
				state.Credits[key] = c
			}
		}
		record := state.Operations[o.plan.ID]
		record.Status = "failed"
		state.Operations[o.plan.ID] = record
		return nil
	})
	if e == nil {
		o.status = "failed"
	}
	return e
}
func (o *Operation) checkLease(ctx context.Context) error {
	if o.session.explicitInventory != nil {
		return ctx.Err()
	}
	state, e := o.session.options.Store.ReadSnapshot(ctx, o.session.scope)
	if e != nil {
		return e
	}
	if state.Invalidated {
		return errors.New("wallet requires reset refresh")
	}
	record, ok := state.Operations[o.plan.ID]
	if !ok || record.WorkerID != o.workerID {
		return errors.New("operation worker was replaced; resume operation")
	}
	return nil
}

// Convert public statement words without reducing malformed values modulo the field.
func witnessField(hash common.Hash) (fr.Element, error) { return core.HexToElement(hash.Hex()) }
