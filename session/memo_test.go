package session

import (
	"bytes"
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/invisi-lab/invisi-sdk-go/core"
	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestMemoPreparationCopiesAndPersists(t *testing.T) {
	session, token, _ := testSession(t)
	address, _ := session.wallet.GetAddress()
	memo := []byte{0xff, 0, 1}
	prepared, err := session.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30), Payload: memo}})})
	if err != nil {
		t.Fatal(err)
	}
	memo[0] = 0
	if !bytes.Equal(prepared.plan.Outputs[0].Note.Payload, []byte{0xff, 0, 1}) {
		t.Fatal("caller memo mutation changed prepared note")
	}
	resumed, err := session.Resume(context.Background(), prepared.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(resumed.plan.Outputs[0].Note.Payload, []byte{0xff, 0, 1}) {
		t.Fatal("memo lost on resume")
	}
}

func TestMemoStoreCloneAndFundsDisclosureBoundary(t *testing.T) {
	ctx := context.Background()
	session, token, _ := testSession(t)
	snapshot, err := session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	var stored StoredNote
	for _, current := range snapshot.Notes {
		stored = current
	}
	stored.Note.Payload = []byte{0xff, 0, 1}
	if err := session.options.Store.Update(ctx, session.scope, func(state *WalletState) error { state.Notes[stored.Commitment] = stored; return nil }); err != nil {
		t.Fatal(err)
	}
	stored.Note.Payload[0] = 1
	snapshot, err = session.options.Store.ReadSnapshot(ctx, session.scope)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Notes[stored.Commitment].Note.Payload[0] != 0xff {
		t.Fatal("store retained caller memo")
	}
	codec, err := NewEncryptedWalletCodec(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := codec.Seal(session.scope, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := codec.Open(session.scope, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Notes[stored.Commitment].Note.Payload[0] = 2
	note := restored.Notes[stored.Commitment].Note
	if note.Payload[0] != 0xff {
		t.Fatal("encrypted storage lost or aliased memo")
	}
	opening, err := sdk.DisclosureFromNote(note)
	if err != nil {
		t.Fatal(err)
	}
	output := sdk.DisclosureOutput{ProtocolContext: note.ProtocolContext, TokenAddress: core.BigIntToElement(token.Big()), OutputCommitment: stored.Commitment, Role: "funds"}
	result, err := opening.Verify(output)
	if err != nil || len(result.Note.Payload) != 0 || opening.GetScope().RevealsPayload {
		t.Fatal("opening claimed memo authenticity", err)
	}
	output.Ciphertext, err = core.EncryptPrivateTransferNote(session.viewing.GetPublicKey(), note, core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(stored.Commitment.Big())})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := session.viewing.DiscloseNote(output.Ciphertext, core.NoteDecryptionContext{NoteEncryptionContext: core.NoteEncryptionContext{ProtocolContext: note.ProtocolContext, OutputCommitment: core.BigIntToElement(stored.Commitment.Big())}, TokenAddress: note.TokenAddress})
	if err != nil {
		t.Fatal(err)
	}
	result, err = payload.Verify(output)
	if err != nil || !bytes.Equal(result.Note.Payload, note.Payload) || !payload.GetScope().RevealsPayload {
		t.Fatal("payload memo not authenticated", err)
	}
}

type unsupportedProtocolProvider struct{ sessionProvider }

func (p unsupportedProtocolProvider) CallContract(ctx context.Context, message ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if bytes.Equal(message.Data, crypto.Keccak256([]byte("VERSION()"))[:4]) {
		return common.LeftPadBytes([]byte{2}, 32), nil
	}
	return p.sessionProvider.CallContract(ctx, message, block)
}
func TestMemoProtocolGuardAcceptsVersionOneAndRejectsVersionTwo(t *testing.T) {
	base := sessionProvider{}
	pool, err := sdk.NewProtocol(protocolrpc.ProtocolDeployment{ChainID: "31337", Protocol: protocolrpc.ContractDeployment{Address: common.HexToAddress("0x1111111111111111111111111111111111111111"), RuntimeCodeHash: transaction.HashCode([]byte{0x60, 0}), FromBlock: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if err = pool.ValidateDeployment(context.Background(), base); err != nil {
		t.Fatal("Protocol VERSION 1 rejected", err)
	}
	if err = pool.ValidateDeployment(context.Background(), unsupportedProtocolProvider{base}); err == nil {
		t.Fatal("Protocol VERSION 2 accepted")
	}
}

type memoPreparationBarrier struct {
	sessionProvider
	entered, release chan struct{}
	once             sync.Once
}

func (p *memoPreparationBarrier) ChainID(ctx context.Context) (*big.Int, error) {
	p.once.Do(func() { close(p.entered) })
	select {
	case <-p.release:
		return p.sessionProvider.ChainID(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func TestMemoPreparationSnapshotsBeforeNetworkWait(t *testing.T) {
	session, token, _ := testSession(t)
	address, _ := session.wallet.GetAddress()
	memo := []byte("original")
	provider := &memoPreparationBarrier{sessionProvider: session.options.Provider.(sessionProvider), entered: make(chan struct{}), release: make(chan struct{})}
	session.options.Provider = provider
	session.rpc = session.rpc.WithProvider(provider)
	type preparedResult struct {
		operation *Operation
		err       error
	}
	result := make(chan preparedResult, 1)
	go func() {
		operation, err := session.PrepareTransfer(context.Background(), sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(session, []testPaymentOutput{{To: address, Amount: big.NewInt(30), Payload: memo}})})
		result <- preparedResult{operation, err}
	}()
	<-provider.entered
	memo[0] = 'X'
	close(provider.release)
	prepared := <-result
	if prepared.err != nil {
		t.Fatal(prepared.err)
	}
	if string(prepared.operation.plan.Outputs[0].Note.Payload) != "original" {
		t.Fatal("caller changed memo during network wait")
	}
}
