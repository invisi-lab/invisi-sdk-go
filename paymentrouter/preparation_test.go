package paymentrouter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-labs/invisi-sdk-go/rpc"
	"github.com/invisi-labs/invisi-sdk-go/sdk"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

type preparationProvider struct{ *chain }

func (preparationProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return nil, fmt.Errorf("no receipt")
}
func (p preparationProvider) CallContract(ctx context.Context, m ethereum.CallMsg, n *big.Int) ([]byte, error) {
	if *m.To == addr(1) {
		a, _ := protocol.ProtocolABI()
		method, e := a.MethodById(m.Data[:4])
		if e == nil && method.RawName == "VERSION" {
			return method.Outputs.Pack(bi(1))
		}
		return nil, fmt.Errorf("unexpected core request")
	}
	return p.chain.CallContract(ctx, m, n)
}

type proofBackend struct{}

func (proofBackend) Prove(_ context.Context, r prover.ProvingRequest, _ prover.Options) (prover.ProofBundle, error) {
	proof := make([]byte, 32)
	proof[31] = 1
	hash := sha256.Sum256(proof)
	return prover.ProofBundle{FormatVersion: 1, CircuitName: r.CircuitName, Target: r.Target, BBVersion: "6.0.0", ArtifactSHA256: r.ArtifactSHA256, VKSHA256: r.VKSHA256, Proof: proof, ProofSHA256: hex.EncodeToString(hash[:]), PublicInputs: append([]common.Hash(nil), r.ExpectedPublicInputs...)}, nil
}
func (proofBackend) Verify(context.Context, prover.ProofBundle, prover.ProofStatement, prover.Options) (bool, error) {
	return true, nil
}
func TestActualSDKPreparationAndProofLifecycle(t *testing.T) {
	ctx := context.Background()
	rpc := newChain()
	dep := protocolrpc.ProtocolDeployment{ChainID: "31337", Protocol: protocolrpc.ContractDeployment{Address: addr(1), RuntimeCodeHash: crypto.Keccak256Hash([]byte{0x60, 0}), FromBlock: 1}}
	pool, e := sdk.NewProtocol(dep)
	if e != nil {
		t.Fatal(e)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64
			Method string
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "invisi_getDeployment" {
			t.Errorf("unexpected protocol RPC %s", req.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": dep}})
	}))
	defer server.Close()
	membership, e := protocolrpc.NewProtocolRpcClient(server.URL + "/v1")
	if e != nil {
		t.Fatal(e)
	}
	key := make([]byte, 32)
	key[31] = 1
	material, e := primitives.DeriveWallet(key)
	if e != nil {
		t.Fatal(e)
	}
	address, e := primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: material.OwnerAddress, ViewingPublicKey: material.ViewingPublicKey, ClueKey: material.ClueKey})
	if e != nil {
		t.Fatal(e)
	}
	wallet, e := sdk.WalletFromAddress(address)
	if e != nil {
		t.Fatal(e)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: preparationProvider{chain: rpc}, ProtocolRPC: membership, Prover: proofBackend{}})
	c, _ := NewClient(deployment(), rpc)
	outputs := []sdk.NoteOutput{}
	for i, n := range []int64{3, 5} {
		o, e := sdk.NewNoteOutput(sdk.PaymentRequest{ChainID: bi(31337), ProtocolAddress: addr(1), TokenAddress: addr(2), Amount: bi(n), Recipient: sdk.IdentityRecipient{Identity: common.BigToHash(bi(int64(i + 100)))}})
		if e != nil {
			t.Fatal(e)
		}
		outputs = append(outputs, o)
	}
	fee := &sdk.Fee{Recipient: addr(10), Amount: bi(2)}
	opts := PreparePaymentOptions{Invisi: invisi, Wallet: wallet, Protocol: pool, TokenAddress: addr(2), Outputs: outputs, Fee: fee, Deadline: 1500}
	rpc.hook = func() { fee.Amount.SetInt64(9); outputs[0], outputs[1] = outputs[1], outputs[0] }
	prepared, e := c.PreparePayment(ctx, opts)
	if e != nil {
		t.Fatal(e)
	}
	preview := prepared.Preview()
	if preview.InputCount != 0 || preview.FundingPayer != addr(3) || preview.FundingAmount.Cmp(bi(10)) != 0 || preview.FeeAmount.Cmp(bi(2)) != 0 {
		t.Fatal("wrong SDK funding preview", preview)
	}
	ready, e := prepared.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	call, e := ready.ExportTransaction(ctx)
	if e != nil {
		t.Fatal(e)
	}
	_, r, e := InspectCorePayment(call, deployment(), bi(1000))
	if e != nil {
		t.Fatal(e)
	}
	if r.Params.NoteCommitments[0] != outputs[1].Commitment() || r.Params.NoteCommitments[1] != outputs[0].Commitment() || r.Funding.Amount.Cmp(bi(10)) != 0 || r.Params.FeeAmount.Cmp(bi(2)) != 0 {
		t.Fatal("caller mutation changed proved order or amount")
	}
	a, _ := NewUniswapV3Adapter(rpc, addr(9), []V3Route{{Tokens: []common.Address{addr(4), addr(2)}, Fees: []uint32{3000}}})
	quote, e := c.Quote(ctx, a, parameters())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.BuildPaymentTransaction(ctx, ready, quote, addr(10)); e != nil {
		t.Fatal(e)
	}
	opts.Outputs = outputs[:1]
	if _, e = c.PreparePayment(ctx, opts); e == nil {
		t.Fatal("accepted one pure funding output")
	}
}
