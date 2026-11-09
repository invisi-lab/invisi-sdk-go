package credit

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestCurrentCreditNativeLifecycle(t *testing.T) {
	bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH")
	if bb == "" || noir == "" {
		t.Skip("current native prover tools are not configured")
	}
	backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4, Timeout: 3 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
	defer cancel()
	ownerKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{1}, 32))
	spenderKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{2}, 32))
	owner, spender := crypto.PubkeyToAddress(ownerKey.PublicKey), crypto.PubkeyToAddress(spenderKey.PublicKey)
	ownerWallet, _ := core.DeriveWallet(crypto.FromECDSA(ownerKey))
	spenderWallet, _ := core.DeriveWallet(crypto.FromECDSA(spenderKey))
	state := State{AvailableCredit: core.ElementFromUint64(400), ProtocolContext: core.ProtocolContext{ChainID: core.ElementFromUint64(31337), ProtocolAddress: core.ElementFromUint64(17)}, TokenAddress: core.ElementFromUint64(34), SpenderWalletIdentity: core.BigIntToElement(spender.Big()), OwnerWalletIdentity: core.BigIntToElement(owner.Big()), RevocationShard: core.ElementFromUint64(1), Nonce: core.ElementFromUint64(55), Expiry: core.ElementFromUint64(2100000000), SharedKey: core.ElementFromUint64(66), Policy: Policy{SingleLimit: core.ElementFromUint64(100), DailyLimit: core.ElementFromUint64(200), MonthlyLimit: core.ElementFromUint64(300)}}
	c := Context{ProtocolContext: state.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	deadline := big.NewInt(2000000300)
	collector := common.HexToAddress("0x1234")
	fee := big.NewInt(1)
	deliveryState := DeliveryState{State: state, OwnerReceivingKeys: core.ReceivingKeys{ViewingPublicKey: ownerWallet.ViewingPublicKey, ClueKey: ownerWallet.ClueKey}, SpenderReceivingKeys: core.ReceivingKeys{ViewingPublicKey: spenderWallet.ViewingPublicKey, ClueKey: spenderWallet.ClueKey}}
	deliveries := func(s State) []protocol.Delivery {
		t.Helper()
		d := deliveryState
		d.State = s
		commitment, e := s.Commitment()
		if e != nil {
			t.Fatal(e)
		}
		result := []protocol.Delivery{}
		for _, role := range []Role{Owner, Spender} {
			identity := s.OwnerWalletIdentity
			keys := d.OwnerReceivingKeys
			if role == Spender {
				identity = s.SpenderWalletIdentity
				keys = d.SpenderReceivingKeys
			}
			cipher, e := Encrypt(d, DeliveryContext{ProtocolContext: s.ProtocolContext, TokenAddress: s.TokenAddress, Commitment: commitment, RecipientIdentity: identity}, role)
			if e != nil {
				t.Fatal(e)
			}
			clue, e := core.CreateClue(keys.ClueKey)
			if e != nil {
				t.Fatal(e)
			}
			result = append(result, protocol.Delivery{Ciphertext: cipher, Clue: clue})
		}
		return result
	}
	quoteKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, c, quoteKey, collector)
	prove := func(name string, p *Prepared, key *ecdsa.PrivateKey) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			if e := p.WithBroadcaster(ctx, client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(2)}); e != nil {
				t.Fatal(e)
			}
			typed, e := p.SigningRequest()
			if e != nil {
				t.Fatal(e)
			}
			digest, _, e := apitypes.TypedDataAndHash(typed)
			if e != nil {
				t.Fatal(e)
			}
			sig, e := crypto.Sign(digest, key)
			if e != nil {
				t.Fatal(e)
			}
			if e = p.WithAuthority(Authority{PublicKey: crypto.FromECDSAPub(&key.PublicKey), Signature: sig[:64]}); e != nil {
				t.Fatal(e)
			}
			if e = p.Prove(ctx, backend, prover.Options{}); e != nil {
				t.Fatal(e)
			}
			data, e := p.Calldata(owner)
			if e != nil || len(data) < 4 {
				t.Fatal("current Credit calldata", e)
			}
			status, e := p.SubmitBroadcaster(ctx, client, "native-"+name)
			if e != nil || status.State != "submitted" {
				t.Fatal("current Credit broadcaster submission", e)
			}
		})
	}
	commitment, _ := state.Commitment()
	created, e := PrepareCreate(c, CreateParams{Token: c.TokenAddress, Owner: owner, RevocationShard: 1, CreditCommitment: common.Hash(commitment.Bytes()), Deadline: deadline, FeeRecipient: collector, FeeAmount: fee}, deliveries(state), state)
	if e != nil {
		t.Fatal(e)
	}
	prove("create", created, ownerKey)
	debit := core.ElementFromUint64(7)
	drawn, e := state.Draw(debit, c.Timestamp)
	if e != nil {
		t.Fatal(e)
	}
	if drawn.State.AvailableCredit != core.ElementFromUint64(393) || drawn.State.TotalSpent != core.ElementFromUint64(7) {
		t.Fatal("gross fee-inclusive draw quota mismatch")
	}
	input := LocatedState{State: state, TreeID: 1, Debit: debit}
	root, e := input.root()
	if e != nil {
		t.Fatal(e)
	}
	outputIdentity, e := core.ComputeNoteIdentity(core.ElementFromUint64(99), core.ElementFromUint64(100))
	if e != nil {
		t.Fatal(e)
	}
	noteCommitment, e := core.ComputeNoteCommitment(state.ChainID, state.ProtocolAddress, outputIdentity, state.TokenAddress, core.ElementFromUint64(6))
	if e != nil {
		t.Fatal(e)
	}
	drawParams := DrawParams{Token: c.TokenAddress, Sources: []protocol.InputSource{{TreeId: big.NewInt(1), MerkleRoot: root}}, Payers: []Payer{{Payer: owner, Amount: big.NewInt(7)}}, GrantRefs: []GrantRef{{PayerIndex: big.NewInt(0), Shard: 1}}, Nullifiers: []common.Hash{common.Hash(drawn.Nullifier.Bytes())}, CreditCommitments: []common.Hash{common.Hash(drawn.StateCommitment.Bytes())}, RecoveryAmounts: []common.Hash{common.Hash(drawn.RecoveryAmount.Bytes())}, NoteCommitment: common.Hash(noteCommitment.Bytes()), TotalAmount: big.NewInt(7), Deadline: deadline, FeeRecipient: collector, FeeAmount: fee}
	draw, e := PrepareDraw(c, drawParams, append(deliveries(drawn.State), protocol.Delivery{}), []LocatedState{input}, outputIdentity)
	if e != nil {
		t.Fatal(e)
	}
	prove("draw", draw, spenderKey)
	changedPolicy := Policy{SingleLimit: core.ElementFromUint64(10), DailyLimit: core.ElementFromUint64(40), MonthlyLimit: core.ElementFromUint64(60)}
	updateChange := Change{Policy: changedPolicy, Expiry: core.ElementFromUint64(2200000000), IncreaseAvailableCredit: core.ElementFromUint64(100)}
	updated, e := drawn.State.Update(updateChange, deadline.Uint64())
	if e != nil {
		t.Fatal(e)
	}
	updateInput := LocatedState{State: drawn.State, TreeID: 1}
	updateRoot, e := updateInput.root()
	if e != nil {
		t.Fatal(e)
	}
	update, e := PrepareUpdate(c, UpdateParams{Token: c.TokenAddress, Owner: owner, Source: protocol.InputSource{TreeId: big.NewInt(1), MerkleRoot: updateRoot}, Nullifier: common.Hash(updated.Nullifier.Bytes()), CreditCommitment: common.Hash(updated.StateCommitment.Bytes()), RevocationShard: 1, Deadline: deadline, FeeRecipient: collector, FeeAmount: fee}, deliveries(updated.State), updateInput, updateChange)
	if e != nil {
		t.Fatal(e)
	}
	if updated.State.AvailableCredit != core.ElementFromUint64(493) || updated.State.TotalSpent != core.ElementFromUint64(7) {
		t.Fatal("paid update changed historical usage or failed to increase quota")
	}
	prove("update", update, ownerKey)
}

func TestCurrentRevocationBroadcasterWithoutZK(t *testing.T) {
	s, _ := fixture(t)
	ownerKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{1}, 32))
	owner := crypto.PubkeyToAddress(ownerKey.PublicKey)
	c := Context{ProtocolContext: s.ProtocolContext, TokenAddress: common.BigToAddress(big.NewInt(34)), Timestamp: 2000000000}
	collector := common.HexToAddress("0x1234")
	p, err := PrepareRevoke(c, RevokeParams{Token: c.TokenAddress, Owner: owner, Shard: 1, Slot: 0, Deadline: big.NewInt(2000000300)})
	if err != nil {
		t.Fatal(err)
	}
	quoteKey, _ := crypto.ToECDSA(common.LeftPadBytes([]byte{9}, 32))
	client := creditBroadcaster(t, c, quoteKey, collector)
	if err = p.WithBroadcaster(context.Background(), client, WithBroadcasterOptions{MaxFeeAmount: big.NewInt(2)}); err != nil {
		t.Fatal(err)
	}
	typed, err := p.SigningRequest()
	if err != nil {
		t.Fatal(err)
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := crypto.Sign(digest, ownerKey)
	if err != nil {
		t.Fatal(err)
	}
	verify := func(_ context.Context, expected common.Address, data apitypes.TypedData, sig []byte) (bool, error) {
		h, _, err := apitypes.TypedDataAndHash(data)
		if err != nil {
			return false, err
		}
		pub, err := crypto.SigToPub(h, sig)
		if err != nil {
			return false, err
		}
		return crypto.PubkeyToAddress(*pub) == expected, nil
	}
	if err = p.WithRevocationSignature(context.Background(), signature, verify); err != nil {
		t.Fatal(err)
	}
	status, err := p.SubmitBroadcaster(context.Background(), client, "revoke-no-zk")
	if err != nil || status.State != "submitted" {
		t.Fatal(status, err)
	}
	if _, err = p.Proof(); err == nil {
		t.Fatal("revocation unexpectedly required a ZK proof")
	}
}
