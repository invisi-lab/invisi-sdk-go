package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type noteProvider struct {
	headers                         map[uint64]*types.Header
	latest                          uint64
	spent, unknownRoot, oldRootOnly bool
	spentReads                      int
	spentAt                         uint64
	badCode                         bool
}

func (p *noteProvider) ChainID(context.Context) (*big.Int, error) { return big.NewInt(97), nil }
func (p *noteProvider) HeaderByNumber(_ context.Context, n *big.Int) (*types.Header, error) {
	number := p.latest
	if n != nil {
		number = n.Uint64()
	}
	if h := p.headers[number]; h != nil {
		copy := types.CopyHeader(h)
		return copy, nil
	}
	return nil, errors.New("missing test block")
}
func (p *noteProvider) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	if p.badCode {
		return []byte{1}, nil
	}
	return []byte{0x60, 0}, nil
}
func (p *noteProvider) CallContract(_ context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	abi, _ := protocol.ProtocolABI()
	method, e := abi.MethodById(msg.Data[:4])
	if e != nil {
		return nil, e
	}
	switch method.RawName {
	case "VERSION":
		return method.Outputs.Pack(big.NewInt(1))
	case "rootHistory":
		known := !p.unknownRoot
		if p.oldRootOnly && block != nil && block.Uint64() == 20 {
			known = false
		}
		return method.Outputs.Pack(known)
	case "isNullifierSpent":
		p.spentReads++
		if block != nil {
			p.spentAt = block.Uint64()
		}
		return method.Outputs.Pack(p.spent)
	}
	return nil, errors.New("unexpected chain method")
}

type noteFixture struct {
	note             *PaymentNote
	pool             *Protocol
	wallet           *Wallet
	provider         *noteProvider
	options          PaymentNoteVerificationOptions
	missing, corrupt bool
}

func newNoteFixture(t *testing.T) *noteFixture {
	t.Helper()
	f := &noteFixture{}
	wallet, e := WalletFromPrivateKey("0x" + strings.Repeat("01", 32))
	if e != nil {
		t.Fatal(e)
	}
	f.wallet = wallet
	output, e := NewNoteOutput(PaymentRequest{ChainID: big.NewInt(97), ProtocolAddress: common.HexToAddress("11"), TokenAddress: common.HexToAddress("22"), Amount: big.NewInt(70), Recipient: EVMRecipient{Address: wallet.OwnerAddress(), Nonce: big.NewInt(0)}})
	if e != nil {
		t.Fatal(e)
	}
	ipn, e := output.ExportPaymentNote()
	if e != nil {
		t.Fatal(e)
	}
	f.note, e = PaymentNoteFromIPN(ipn)
	if e != nil {
		t.Fatal(e)
	}
	f.provider = &noteProvider{headers: map[uint64]*types.Header{10: {Number: big.NewInt(10), Time: 100}, 20: {Number: big.NewInt(20), Time: 200}}, latest: 10}
	h := f.provider.headers[10]
	block := protocolrpc.CanonicalBlock{Number: 10, Hash: h.Hash(), Timestamp: h.Time}
	dep := protocolrpc.ProtocolDeployment{ChainID: "97", Protocol: protocolrpc.ContractDeployment{Address: f.note.ProtocolAddress(), RuntimeCodeHash: ethcrypto.Keccak256Hash([]byte{0x60, 0}), FromBlock: 1}}
	root := core.BigIntToElement(output.Commitment().Big())
	path := make([]common.Hash, core.TreeDepth)
	for range path {
		root, e = core.Poseidon2Hash(root, core.ElementFromUint64(0))
		if e != nil {
			t.Fatal(e)
		}
	}
	anchor := map[string]any{"treeId": 1, "leafCount": 1, "root": FieldHash(root), "block": block}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		var result any
		proof := map[string]any{"anchor": anchor, "leafIndex": 0, "commitment": output.Commitment(), "siblings": path}
		switch req.Method {
		case "invisi_getDeployment":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep}
		case "invisi_getStatus":
			result = map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep, "indexedHead": block, "checkedBits": 8, "detectionEnabled": true}
		case "invisi_getRoots":
			result = []any{anchor}
		case "invisi_getMembership":
			if f.corrupt {
				bad := append([]common.Hash(nil), path...)
				bad[0] = common.HexToHash("1")
				proof["siblings"] = bad
			}
			result = proof
		case "invisi_getNote":
			if !f.missing {
				result = map[string]any{"output": map[string]any{"block": block, "treeId": 1, "leafIndex": 0, "commitment": output.Commitment(), "token": f.note.TokenAddress(), "ciphertext": "0x", "clue": "0x", "transactionHash": common.HexToHash("33"), "logIndex": 0, "kind": "Shield"}, "membership": proof}
			}
		default:
			t.Errorf("unexpected RPC %s", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	t.Cleanup(server.Close)
	rpc, e := protocolrpc.NewProtocolRpcClient(server.URL+"/v1", protocolrpc.ClientOptions{Deployment: &dep})
	if e != nil {
		t.Fatal(e)
	}
	f.pool, e = protocol.ProtocolFromRPC(context.Background(), rpc, f.provider)
	if e != nil {
		t.Fatal(e)
	}
	f.options = PaymentNoteVerificationOptions{ProtocolRPC: rpc, Provider: f.provider}
	return f
}
func TestPaymentNoteVerificationLevels(t *testing.T) {
	f := newNoteFixture(t)
	ctx := context.Background()
	f.provider.spent = true
	verified, e := f.note.Verify(ctx, f.pool)
	if e != nil {
		t.Fatal(e)
	}
	if verified.Membership().Leaf != f.note.Commitment() || f.provider.spentReads != 0 {
		t.Fatal("simple verification queried spent state")
	}
	if _, e = f.note.VerifySpendable(ctx, f.pool, f.wallet); e == nil {
		t.Fatal("spent note accepted")
	}
	f.provider.spent = false
	spendable, e := f.note.VerifySpendable(ctx, f.pool, f.wallet)
	if e != nil {
		t.Fatal(e)
	}
	input, e := spendable.ToInput()
	if e != nil || input.Commitment != f.note.Commitment() {
		t.Fatal(input, e)
	}
	input.Note.Amount = core.ElementFromUint64(1)
	copy, _ := spendable.ToInput()
	if core.ElementToBigInt(copy.Note.Amount).Cmp(big.NewInt(70)) != 0 {
		t.Fatal("input was mutable")
	}
	data := f.note.Data()
	data.Amount.SetInt64(1)
	if f.note.Amount().Cmp(big.NewInt(70)) != 0 {
		t.Fatal("note data mutable")
	}
	wrong, _ := WalletFromPrivateKey("0x" + strings.Repeat("02", 32))
	if _, e = f.note.VerifySpendable(ctx, f.pool, wrong); e == nil {
		t.Fatal("wrong owner accepted")
	}
	if _, e = f.note.VerifySpendable(ctx, f.pool, nil); e == nil {
		t.Fatal("nil wallet accepted")
	}
	forged := *wrong
	forged.PublicWallet = f.wallet.PublicWallet
	if _, e = f.note.VerifySpendable(ctx, f.pool, &forged); e == nil {
		t.Fatal("mismatched private material accepted")
	}
}
func TestPaymentNoteVerificationFailureAndCurrentState(t *testing.T) {
	f := newNoteFixture(t)
	ctx := context.Background()
	f.missing = true
	if _, e := f.note.Verify(ctx, f.pool); e == nil {
		t.Fatal("missing leaf accepted")
	}
	f.missing = false
	f.corrupt = true
	if _, e := f.note.Verify(ctx, f.pool); e == nil {
		t.Fatal("corrupt proof accepted")
	}
	f.corrupt = false
	f.provider.badCode = true
	if _, e := f.note.Verify(ctx, f.pool); e == nil {
		t.Fatal("wrong code accepted")
	}
	f.provider.badCode = false
	f.provider.unknownRoot = true
	if _, e := f.note.Verify(ctx, f.pool); e == nil {
		t.Fatal("unknown root accepted")
	}
	f.provider.unknownRoot = false
	f.provider.latest = 20
	verified, e := f.note.VerifySpendable(ctx, f.pool, f.wallet)
	if e != nil {
		t.Fatal(e)
	}
	if verified.CheckedAt().Number != 10 || verified.SpendabilityCheckedAt().Number != 20 || f.provider.spentAt != 20 {
		t.Fatal("spent state was not read at latest block")
	}
	f.provider.oldRootOnly = true
	if _, e = f.note.VerifySpendable(ctx, f.pool, f.wallet); e == nil {
		t.Fatal("expired current root accepted")
	}
	if _, e = f.note.Verify(ctx, f.pool, PaymentNoteVerificationOptions{Confirmations: 100}); e == nil {
		t.Fatal("insufficient confirmations accepted")
	}
	manual, e := protocol.NewProtocol(f.pool.Deployment())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.note.Verify(ctx, manual); e == nil {
		t.Fatal("unconnected Protocol accepted")
	}
	if _, e = f.note.Verify(ctx, manual, f.options); e != nil {
		t.Fatal(e)
	}
	f.provider.headers[10].Extra = []byte{1}
	if _, e = f.note.Verify(ctx, f.pool); e == nil {
		t.Fatal("reorganized origin accepted")
	}
}
