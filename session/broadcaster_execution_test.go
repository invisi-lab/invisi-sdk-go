package session

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestQuotedAggregateDefaultsToRecursiveAndRetainsFrozenRequest(t *testing.T) {
	ctx := context.Background()
	s, token, _ := testSession(t, uint64(time.Now().Unix()))
	if bb, noir := os.Getenv("BB_BINARY_PATH"), os.Getenv("NOIR_EXECUTE_PATH"); bb != "" && noir != "" {
		backend, err := native.NewNativeProver(native.NativeOptions{BBBinary: bb, NoirBinary: noir, Threads: 4})
		if err != nil {
			t.Fatal(err)
		}
		s.options.Prover = backend
	}
	address, _ := s.wallet.GetAddress()
	op, e := s.PrepareTransfer(ctx, sdk.TransferPreparation{TokenAddress: token, Outputs: testOutputs(s, []testPaymentOutput{{To: address, Amount: big.NewInt(30)}})})
	if e != nil {
		t.Fatal(e)
	}
	var sent []json.RawMessage
	statuses := 0
	c, _ := testBroadcasterExecution(t, op, []int64{5}, false, nil, "aggregate", func(method string, raw json.RawMessage) {
		if method == "broadcaster_submitTransaction" {
			sent = append(sent, append([]byte(nil), raw...))
		}
		if method == "broadcaster_getTransactionStatus" {
			statuses++
		}
	})
	e = c.Quote(ctx, op, sdk.BroadcasterQuoteOptions{MaxFeeAmount: big.NewInt(10)})
	if e != nil {
		t.Fatal(e)
	}
	op, e = authorizeForTest(op, ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = op.Prove(ctx, sdk.TransactionProveOptions{Execution: "single"}); e == nil {
		t.Fatal("signed aggregate accepted single proving")
	}
	op, e = op.Prove(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if op.plan.Target != "noir-recursive" {
		t.Fatal(op.plan.Target)
	}
	if _, e = op.ExportTransaction(ctx); e == nil {
		t.Fatal("recursive member exported as EVM transaction")
	}

	lostClient, err := broadcaster.NewClient(c.Configuration(), broadcaster.Options{HTTPClient: &http.Client{Transport: &lostBroadcasterResponse{base: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = op.SubmitBroadcaster(ctx, lostClient); err == nil {
		t.Fatal("lost response did not surface")
	}
	if op.plan.Broadcaster.Request == nil || op.plan.Broadcaster.RequestID != "" {
		t.Fatal("request was not frozen before transport")
	}
	op, e = s.Resume(ctx, op.ID())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = op.SubmitBroadcaster(ctx, c); e != nil {
		t.Fatal(e)
	}
	if len(sent) != 2 || !bytes.Equal(sent[0], sent[1]) {
		t.Fatal("lost response retry changed request")
	}
	var wire broadcaster.Submission
	if e = json.Unmarshal(sent[0], &wire); e != nil {
		t.Fatal(e)
	}
	if wire.MemberProof == nil || wire.MemberProof.Target != "noir-recursive" || wire.MemberProof.ProofHex != "0x"+common.Bytes2Hex(op.bundle.Proof) {
		t.Fatal("public recursive member missing")
	}
	op, e = s.Resume(ctx, op.ID())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = op.SubmitBroadcaster(ctx, c); e != nil {
		t.Fatal(e)
	}
	if len(sent) != 2 || statuses != 1 {
		t.Fatalf("send=%d status=%d", len(sent), statuses)
	}
	op.plan.Broadcaster.RequestID = "not-found"
	if _, err := op.SubmitBroadcaster(ctx, c); err == nil {
		t.Fatal("missing known ID accepted")
	}
	if len(sent) != 2 {
		t.Fatal("missing known ID resubmitted")
	}
	for _, state := range []string{"queued", "proving", "rejected", "expired"} {
		op.plan.Broadcaster.RequestID = state
		status, err := op.SubmitBroadcaster(ctx, c)
		if err != nil || status.State != state {
			t.Fatal(status, err)
		}
		snapshot, err := s.options.Store.ReadSnapshot(ctx, s.scope)
		if err != nil {
			t.Fatal(err)
		}
		for _, input := range op.plan.Inputs {
			if snapshot.Notes[input.Commitment].ReservedBy != op.ID() {
				t.Fatal("operator status released reservation", state)
			}
		}
	}
	a, e := protocol.ProtocolABI()
	if e != nil {
		t.Fatal(e)
	}
	aggregate := ""
	for name, m := range a.Methods {
		if m.RawName == "transact" && m.Inputs[1].Type.T == abi.SliceTy {
			aggregate = name
		}
	}
	request := protocol.TransactRequest{Params: op.plan.Params, Deliveries: op.deliveries(), Funding: op.plan.Funding}
	wrong := request
	wrong.Params.FeeAmount = new(big.Int).Add(request.Params.FeeAmount, big.NewInt(1))
	for _, tc := range []struct {
		members []protocol.TransactRequest
		want    bool
	}{{[]protocol.TransactRequest{request}, true}, {[]protocol.TransactRequest{request, request}, false}, {[]protocol.TransactRequest{}, false}, {[]protocol.TransactRequest{wrong}, false}} {
		data, e := a.Pack(aggregate, []byte{1}, tc.members)
		if e != nil {
			t.Fatal(e)
		}
		if op.matchesTransactionData(data) != tc.want {
			t.Fatal("canonical aggregate membership check")
		}
	}
	op.plan.Broadcaster.Request.Data = "0x00"
	if err := op.validateRecovered(); err == nil {
		t.Fatal("restored mutated public request accepted")
	}

}

// The relay accepts the request but the connection loses its response body.
type lostBroadcasterResponse struct{ base http.RoundTripper }

func (t *lostBroadcasterResponse) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := request.GetBody()
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(body)
	body.Close()
	if err != nil {
		return nil, err
	}
	response, err := t.base.RoundTrip(request)
	if err == nil && strings.Contains(string(raw), "broadcaster_submitTransaction") {
		response.Body.Close()
		return nil, errors.New("submission response lost")
	}
	return response, err
}
