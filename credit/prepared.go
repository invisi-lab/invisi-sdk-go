package credit

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type Intent struct {
	Version           int
	Context           Context
	Create            *CreateParams
	Draw              *DrawParams
	Update            *UpdateParams
	Revoke            *RevokeParams
	Deliveries        []protocol.Delivery
	CreatedState      *State
	Inputs            []LocatedState
	UpdateChange      *Change
	OutputIdentity    fr.Element
	OutputNote        *primitives.Note
	NoteRecipientKeys *primitives.ReceivingKeys
	Broadcaster       *BroadcasterBinding
}

// Prepared keeps authorization and local witnesses outside portable recovery plans.
type Prepared struct {
	broadcaster.QuoteTarget `json:"-"`
	mu                      sync.Mutex
	intent                  Intent
	authority               *Authority
	proof                   *prover.ProofBundle
	revokeSignature         []byte
	submission              Submission
}

func (p *Prepared) String() string   { return "PreparedCredit([redacted])" }
func (p *Prepared) GoString() string { return p.String() }
func (p *Prepared) MarshalJSON() ([]byte, error) {
	return []byte(`{"preparedCredit":"[redacted]"}`), nil
}
func freeze[T any](v T) (T, error) {
	var result T
	raw, err := json.Marshal(v)
	if err != nil {
		return result, err
	}
	defer clear(raw)
	err = json.Unmarshal(raw, &result)
	return result, err
}
func (i Intent) statement() ([]common.Hash, apitypes.TypedData, error) {
	switch {
	case i.Create != nil:
		return i.Create.Statement(i.Context, i.Deliveries)
	case i.Draw != nil:
		return i.Draw.Statement(i.Context, i.Deliveries)
	case i.Update != nil:
		return i.Update.Statement(i.Context, i.Deliveries)
	case i.Revoke != nil:
		t, err := i.Revoke.TypedData(i.Context)
		return nil, t, err
	}
	return nil, apitypes.TypedData{}, fmt.Errorf("Credit operation required")
}
func prepare(i Intent) (*Prepared, error) {
	count := 0
	for _, present := range []bool{i.Create != nil, i.Draw != nil, i.Update != nil, i.Revoke != nil} {
		if present {
			count++
		}
	}
	if count != 1 || i.Version != 2 {
		return nil, fmt.Errorf("invalid Credit intent version or operation")
	}
	frozen, err := freeze(i)
	if err != nil {
		return nil, err
	}
	if _, _, err = frozen.statement(); err != nil {
		return nil, err
	}
	if frozen.Update != nil {
		if len(frozen.Inputs) != 1 || frozen.UpdateChange == nil {
			return nil, fmt.Errorf("update requires one Credit input and change")
		}
		next, err := frozen.Inputs[0].State.Update(*frozen.UpdateChange, frozen.Update.Deadline.Uint64())
		if err != nil {
			return nil, err
		}
		if common.Hash(next.Nullifier.Bytes()) != frozen.Update.Nullifier || common.Hash(next.StateCommitment.Bytes()) != frozen.Update.CreditCommitment {
			return nil, fmt.Errorf("update change differs from frozen transition")
		}
	}
	return (&Prepared{intent: frozen}).bindBroadcasterQuote(), nil
}
func PrepareCreate(c Context, p CreateParams, ds []protocol.Delivery, state State) (*Prepared, error) {
	return prepare(Intent{Version: 2, Context: c, Create: &p, Deliveries: ds, CreatedState: &state})
}
func PrepareUpdate(c Context, p UpdateParams, ds []protocol.Delivery, input LocatedState, change Change) (*Prepared, error) {
	return prepare(Intent{Version: 2, Context: c, Update: &p, Deliveries: ds, Inputs: []LocatedState{input}, UpdateChange: &change})
}
func PrepareDraw(c Context, p DrawParams, ds []protocol.Delivery, inputs []LocatedState, outputIdentity fr.Element) (*Prepared, error) {
	return prepare(Intent{Version: 2, Context: c, Draw: &p, Deliveries: ds, Inputs: inputs, OutputIdentity: outputIdentity})
}
func PrepareRevoke(c Context, p RevokeParams) (*Prepared, error) {
	return prepare(Intent{Version: 2, Context: c, Revoke: &p})
}
func (p *Prepared) SigningRequest() (apitypes.TypedData, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, typed, err := p.intent.statement()
	return typed, err
}

// ExportIntent contains private grant openings and must only be stored in authenticated encrypted storage.
func (p *Prepared) ExportIntent() (Intent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return freeze(p.intent)
}
func (p *Prepared) witness(auth Authority) (*prover.LocalWitness, error) {
	i := p.intent
	switch {
	case i.Create != nil:
		if i.CreatedState == nil {
			return nil, fmt.Errorf("created Credit opening required")
		}
		return i.Create.Witness(i.Context, i.Deliveries, *i.CreatedState, auth)
	case i.Update != nil:
		if len(i.Inputs) != 1 || i.UpdateChange == nil {
			return nil, fmt.Errorf("update requires one Credit input")
		}
		return i.Update.Witness(i.Context, i.Deliveries, i.Inputs[0], *i.UpdateChange, auth)
	case i.Draw != nil:
		return i.Draw.Witness(i.Context, i.Deliveries, i.Inputs, i.OutputIdentity, auth)
	}
	return nil, fmt.Errorf("revocation does not use a ZK witness")
}
func (p *Prepared) WithAuthority(auth Authority) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	frozen := Authority{PublicKey: append([]byte(nil), auth.PublicKey...), Signature: append([]byte(nil), auth.Signature...)}
	if _, err := p.witness(frozen); err != nil {
		return err
	}
	p.authority = &frozen
	p.proof = nil
	return nil
}
func (p *Prepared) expected() (prover.ProofStatement, error) {
	inputs, _, err := p.intent.statement()
	if err != nil {
		return prover.ProofStatement{}, err
	}
	name := ""
	switch {
	case p.intent.Create != nil:
		name = "create_credit"
	case p.intent.Update != nil:
		name = "credit_update"
	case p.intent.Draw != nil:
		capacity, err := DrawCapacity(len(p.intent.Draw.Nullifiers))
		if err != nil {
			return prover.ProofStatement{}, err
		}
		name = fmt.Sprintf("credit_draw_slots%d", capacity)
	default:
		return prover.ProofStatement{}, fmt.Errorf("revocation does not use a proof")
	}
	return prover.Statement(name, "evm", inputs)
}
func (p *Prepared) Prove(ctx context.Context, backend prover.Prover, options prover.Options) error {
	if backend == nil {
		return fmt.Errorf("local prover required")
	}
	p.mu.Lock()
	if p.authority == nil {
		p.mu.Unlock()
		return fmt.Errorf("Credit wallet authorization required")
	}
	auth := Authority{PublicKey: append([]byte(nil), p.authority.PublicKey...), Signature: append([]byte(nil), p.authority.Signature...)}
	witness, err := p.witness(auth)
	if err != nil {
		p.mu.Unlock()
		return err
	}
	statement, err := p.expected()
	p.mu.Unlock()
	if err != nil {
		return err
	}
	bundle, err := backend.Prove(ctx, prover.ProvingRequest{ProofStatement: statement, Witness: witness}, options)
	if err != nil {
		return err
	}
	bundle = prover.CloneBundle(bundle)
	if err = prover.AssertProofBundle(bundle, statement); err != nil {
		return err
	}
	valid, err := backend.Verify(ctx, prover.CloneBundle(bundle), prover.CloneStatement(statement), options)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("Credit proof independent verification failed")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proof = &bundle
	return nil
}
func (p *Prepared) Proof() (*prover.ProofBundle, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.proof == nil {
		return nil, fmt.Errorf("Credit proof unavailable")
	}
	copy := prover.CloneBundle(*p.proof)
	return &copy, nil
}

// WithRevocationSignature verifies an EOA or ERC-1271 owner signature through the caller's verifier.
func (p *Prepared) WithRevocationSignature(ctx context.Context, signature []byte, verify func(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error)) error {
	p.mu.Lock()
	if p.intent.Revoke == nil {
		p.mu.Unlock()
		return fmt.Errorf("not a revocation")
	}
	owner := p.intent.Revoke.Owner
	_, typed, err := p.intent.statement()
	p.mu.Unlock()
	if err != nil {
		return err
	}
	if len(signature) == 0 || verify == nil {
		return fmt.Errorf("owner signature and verifier required")
	}
	frozen := append([]byte(nil), signature...)
	valid, err := verify(ctx, owner, typed, append([]byte(nil), frozen...))
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid Credit owner revocation signature")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.revokeSignature = frozen
	return nil
}
func (p *Prepared) Calldata(caller common.Address) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	i := p.intent
	if i.Revoke != nil {
		return i.Revoke.Calldata(i.Context, p.revokeSignature, caller)
	}
	if p.proof == nil {
		return nil, fmt.Errorf("independently verified Credit proof required")
	}
	statement, err := p.expected()
	if err != nil {
		return nil, err
	}
	if err = prover.AssertProofBundle(*p.proof, statement); err != nil {
		return nil, err
	}
	switch {
	case i.Create != nil:
		return i.Create.Calldata(i.Context, p.proof.Proof, i.Deliveries)
	case i.Draw != nil:
		return i.Draw.Calldata(i.Context, p.proof.Proof, i.Deliveries)
	case i.Update != nil:
		return i.Update.Calldata(i.Context, p.proof.Proof, i.Deliveries)
	}
	return nil, fmt.Errorf("Credit operation required")
}

// Restore reopens an encrypted recovery intent; authorization signatures are never restored.
// A saved proof is independently verified before it can authorize calldata export.
func Restore(ctx context.Context, intent Intent, bundle *prover.ProofBundle, backend prover.Prover) (*Prepared, error) {
	prepared, err := prepare(intent)
	if err != nil {
		return nil, err
	}
	if bundle == nil {
		return prepared, nil
	}
	if backend == nil {
		return nil, fmt.Errorf("saved Credit proof requires independent verifier")
	}
	statement, err := prepared.expected()
	if err != nil {
		return nil, err
	}
	frozen := prover.CloneBundle(*bundle)
	if err = prover.AssertProofBundle(frozen, statement); err != nil {
		return nil, err
	}
	valid, err := backend.Verify(ctx, prover.CloneBundle(frozen), prover.CloneStatement(statement), prover.Options{})
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("saved Credit proof verification failed")
	}
	prepared.proof = &frozen
	return prepared, nil
}

func (p *Prepared) AuthorizationRequest() (primitives.WalletAuthorizationRequest, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	inputs, _, err := p.intent.statement()
	if err != nil {
		return primitives.WalletAuthorizationRequest{}, err
	}
	operation := primitives.WalletOperation("")
	capacity := 1
	switch {
	case p.intent.Create != nil:
		operation = primitives.WalletCreateCredit
	case p.intent.Update != nil:
		operation = primitives.WalletUpdateCredit
	case p.intent.Draw != nil:
		operation = primitives.WalletCreditDraw
		capacity, _ = DrawCapacity(len(p.intent.Draw.Nullifiers))
	default:
		return primitives.WalletAuthorizationRequest{}, fmt.Errorf("revocation uses owner EIP712 authorization")
	}
	words := make([]fr.Element, len(inputs)-2)
	for i := range words {
		words[i] = primitives.BigIntToElement(inputs[i].Big())
	}
	return primitives.WalletAuthorizationRequest{Context: p.intent.Context.ProtocolContext, Operation: operation, Capacity: capacity, Words: words}, nil
}

// PrepareDrawToNote freezes the ordinary Note opening and its public delivery keys for fee quotation.
func PrepareDrawToNote(c Context, p DrawParams, stateDeliveries []protocol.Delivery, inputs []LocatedState, note primitives.Note, keys *primitives.ReceivingKeys) (*Prepared, error) {
	if note.ProtocolContext != c.ProtocolContext || primitives.ElementToBigInt(note.TokenAddress).Cmp(p.Token.Big()) != 0 {
		return nil, fmt.Errorf("ordinary Note context mismatch")
	}
	fee, err := normalizeFee(p.FeeRecipient, p.FeeAmount)
	if err != nil {
		return nil, err
	}
	if p.TotalAmount == nil || p.TotalAmount.Cmp(fee) <= 0 {
		return nil, fmt.Errorf("Credit fee must leave a positive Note")
	}
	net := new(big.Int).Sub(p.TotalAmount, fee)
	if primitives.ElementToBigInt(note.Amount).Cmp(net) != 0 {
		return nil, fmt.Errorf("ordinary Note amount must equal gross debit minus fee")
	}
	identity, err := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	if err != nil {
		return nil, err
	}
	commitment, err := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	if err != nil {
		return nil, err
	}
	if p.NoteCommitment != (common.Hash{}) && p.NoteCommitment != common.Hash(commitment.Bytes()) {
		return nil, fmt.Errorf("ordinary Note commitment mismatch")
	}
	p.NoteCommitment = common.Hash(commitment.Bytes())
	if len(stateDeliveries) != 2*len(inputs) {
		return nil, fmt.Errorf("draw requires two deliveries per successor grant")
	}
	deliveries := append([]protocol.Delivery(nil), stateDeliveries...)
	delivery, err := noteDelivery(note, keys)
	if err != nil {
		return nil, err
	}
	deliveries = append(deliveries, delivery)
	return prepare(Intent{Version: 2, Context: c, Draw: &p, Deliveries: deliveries, Inputs: inputs, OutputIdentity: identity, OutputNote: &note, NoteRecipientKeys: keys})
}
func noteDelivery(note primitives.Note, keys *primitives.ReceivingKeys) (protocol.Delivery, error) {
	if keys == nil {
		return protocol.Delivery{}, nil
	}
	identity, err := primitives.ComputeNoteIdentity(note.OwnerAddress, note.Nonce)
	if err != nil {
		return protocol.Delivery{}, err
	}
	commitment, err := primitives.ComputeNoteCommitment(note.ChainID, note.ProtocolAddress, identity, note.TokenAddress, note.Amount)
	if err != nil {
		return protocol.Delivery{}, err
	}
	cipher, err := primitives.EncryptPrivateTransferNote(keys.ViewingPublicKey, note, primitives.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: commitment})
	if err != nil {
		return protocol.Delivery{}, err
	}
	clue, err := primitives.CreateClue(keys.ClueKey)
	if err != nil {
		return protocol.Delivery{}, err
	}
	return protocol.Delivery{Ciphertext: cipher, Clue: clue}, nil
}

func (p *Prepared) ExportSigningRequest() (primitives.SigningRequest, error) {
	typed, err := p.SigningRequest()
	if err != nil {
		return primitives.SigningRequest{}, err
	}
	return primitives.NewSigningRequest(typed)
}
func (p *Prepared) WithWalletSignature(signature []byte) error {
	request, err := p.AuthorizationRequest()
	if err != nil {
		return err
	}
	p.mu.Lock()
	var identity fr.Element
	switch {
	case p.intent.Create != nil:
		identity = primitives.BigIntToElement(p.intent.Create.Owner.Big())
	case p.intent.Update != nil:
		identity = primitives.BigIntToElement(p.intent.Update.Owner.Big())
	case p.intent.Draw != nil && len(p.intent.Inputs) > 0:
		identity = p.intent.Inputs[0].State.SpenderWalletIdentity
	default:
		p.mu.Unlock()
		return fmt.Errorf("Credit wallet identity unavailable")
	}
	p.mu.Unlock()
	witness, err := primitives.BuildWalletAuthorizationWitnessFromSignature(identity, request, signature)
	if err != nil {
		return err
	}
	public := append([]byte{4}, witness.PublicKeyX[:]...)
	public = append(public, witness.PublicKeyY[:]...)
	return p.WithAuthority(Authority{PublicKey: public, Signature: witness.Signature[:]})
}
func (p *Prepared) CheckTimestamp(timestamp uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	i := p.intent
	var deadline *big.Int
	switch {
	case i.Create != nil:
		deadline = i.Create.Deadline
	case i.Update != nil:
		deadline = i.Update.Deadline
	case i.Draw != nil:
		deadline = i.Draw.Deadline
		if Periods(timestamp) != Periods(i.Context.Timestamp) {
			return fmt.Errorf("Credit draw crossed a UTC usage period; prepare again")
		}
	case i.Revoke != nil:
		deadline = i.Revoke.Deadline
	default:
		return fmt.Errorf("Credit operation required")
	}
	if deadline == nil || new(big.Int).SetUint64(timestamp).Cmp(deadline) > 0 {
		return fmt.Errorf("Credit operation expired")
	}
	return nil
}
