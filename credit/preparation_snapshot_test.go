package credit

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

func TestPreparationSnapshotsRequestsBeforeFrameReads(t *testing.T) {
	for _, kind := range []string{"create", "update", "draw", "revoke"} {
		t.Run(kind, func(t *testing.T) {
			builder, s, _ := requestBuilder(t)
			entered, release := make(chan struct{}), make(chan struct{})
			frame := builder.Frame
			builder.Frame = func(ctx context.Context, token common.Address, deadline uint64) (Context, uint64, error) {
				close(entered)
				<-release
				return frame(ctx, token, deadline)
			}
			input := builder.Input
			builder.Input = func(ctx context.Context, c Context, h common.Hash, r Role) (PreparationInput, error) {
				if h != common.HexToHash("0x1") {
					return PreparationInput{}, fmt.Errorf("Credit reference mutated")
				}
				return input(ctx, c, h, r)
			}

			token := common.BigToAddress(primitives.ElementToBigInt(s.TokenAddress))
			fee := &Fee{Recipient: common.HexToAddress("0x1234"), Amount: big.NewInt(2)}
			payload := []byte("before")
			amount := big.NewInt(10)
			limit := primitives.ElementFromUint64(90)
			expiry := primitives.ElementFromUint64(2100000000)
			members := []DrawMember{{Credit: common.HexToHash("0x1"), Amount: amount, Payload: payload}}
			type result struct {
				prepared *Prepared
				err      error
			}
			done := make(chan result, 1)
			go func() {
				var p *Prepared
				var err error
				switch kind {
				case "create":
					p, err = builder.Create(context.Background(), CreatePreparation{TokenAddress: token, Spender: "spender", Expiry: 2100000000, AvailableCredit: amount, Payload: payload, Fee: fee})
				case "update":
					p, err = builder.Update(context.Background(), UpdatePreparation{TokenAddress: token, Credit: common.HexToHash("0x1"), UpdatePatch: UpdatePatch{SingleLimit: &limit, Expiry: &expiry, Payload: payload}, Fee: fee})
				case "draw":
					p, err = builder.Draw(context.Background(), DrawPreparation{TokenAddress: token, Credits: members, Recipient: "owner", Payload: payload, Fee: fee})
				case "revoke":
					p, err = builder.Revoke(context.Background(), RevokePreparation{TokenAddress: token, Shard: 1, Fee: fee})
				}
				done <- result{p, err}
			}()
			<-entered
			amount.SetInt64(99)
			fee.Amount.SetInt64(88)
			fee.Recipient = common.HexToAddress("0x9999")
			payload[0] = 'X'
			limit.SetUint64(1)
			expiry.SetUint64(1)
			members[0].Credit = common.HexToHash("0x9")
			close(release)
			got := <-done
			if got.err != nil {
				t.Fatal(got.err)
			}
			intent, err := got.prepared.ExportIntent()
			if err != nil {
				t.Fatal(err)
			}
			recipient, value := intentFee(intent)
			if recipient != common.HexToAddress("0x1234") || value.Int64() != 2 {
				t.Fatal("fee mutated during frame read")
			}
			switch kind {
			case "create":
				if primitives.ElementToBigInt(intent.CreatedState.AvailableCredit).Int64() != 10 || string(intent.CreatedState.Payload) != "before" {
					t.Fatal("create request mutated during frame read")
				}
			case "update":
				if intent.UpdateChange.Policy.SingleLimit != primitives.ElementFromUint64(90) || intent.UpdateChange.Expiry != primitives.ElementFromUint64(2100000000) || string(intent.UpdateChange.Payload) != "before" {
					t.Fatal("update patch mutated during frame read")
				}
			case "draw":
				if intent.Draw.TotalAmount.Int64() != 10 || primitives.ElementToBigInt(intent.OutputNote.Amount).Int64() != 8 || string(intent.OutputNote.Payload) != "before" {
					t.Fatal("draw request mutated during frame read")
				}
			}
		})
	}
}

func TestPreparationSnapshotPreservesPayloadPresence(t *testing.T) {
	for _, payload := range [][]byte{nil, {}} {
		p := DrawPreparation{Payload: payload, Credits: []DrawMember{{Payload: payload}}}.Clone()
		if (p.Payload == nil) != (payload == nil) || (p.Credits[0].Payload == nil) != (payload == nil) {
			t.Fatal("draw snapshot collapsed payload presence")
		}
		u := UpdatePreparation{UpdatePatch: UpdatePatch{Payload: payload}}.Clone()
		if (u.Payload == nil) != (payload == nil) {
			t.Fatal("update snapshot collapsed payload presence")
		}
	}
}
