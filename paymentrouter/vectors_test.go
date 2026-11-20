package paymentrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
	"os"
	"testing"
)

func number(s string) *big.Int { n, _ := new(big.Int).SetString(s, 10); return n }

type vectorCall struct {
	To    common.Address
	Data  string
	Value string
}

func (v vectorCall) call() protocol.TransactionCall {
	return protocol.TransactionCall{To: v.To, Data: common.FromHex(v.Data), Value: number(v.Value)}
}

type vectorQuote struct {
	Channel                                                    Channel
	ChainID                                                    string
	RouterAddress, TokenOut                                    common.Address
	AmountOut, EstimatedInput, MaxInput, Deadline, ReversePath string
	Target                                                     common.Address
	Data                                                       string
	SettlerAddress                                             common.Address
}

func (v vectorQuote) quote() SwapQuote {
	return SwapQuote{Channel: v.Channel, ChainID: number(v.ChainID), RouterAddress: v.RouterAddress, TokenOut: v.TokenOut, AmountOut: number(v.AmountOut), EstimatedInput: number(v.EstimatedInput), MaxInput: number(v.MaxInput), Deadline: number(v.Deadline), ReversePath: common.FromHex(v.ReversePath), Target: v.Target, Data: common.FromHex(v.Data), SettlerAddress: v.SettlerAddress}
}
func TestSharedTypeScriptVectors(t *testing.T) {
	raw, e := os.ReadFile("testdata/payment-router-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		PathCases []struct {
			Name     string
			Tokens   []common.Address
			Fees     []uint32
			Expected string
		}
		ZeroExPayloads []struct {
			Name  string
			Valid bool
			Input struct {
				Target                                 common.Address
				Data                                   string
				RouterAddress, TokenOut                common.Address
				AmountOut, MaxInput                    string
				SettlerAddress, AllowanceHolderAddress common.Address
				Now                                    string
			}
			Expected struct{ Data, NativeDeadline string }
		}
		CorePayment     vectorCall
		RefundRecipient common.Address
		Payments        []struct {
			Name       string
			Quote      vectorQuote
			Expected   vectorCall
			Simulation struct{ GasEstimate, ETHSpent, ETHRefunded string }
		}
	}
	if e = json.Unmarshal(raw, &v); e != nil {
		t.Fatal(e)
	}
	if len(v.ZeroExPayloads) != 14 || len(v.Payments) != 3 {
		t.Fatal("incomplete vectors")
	}
	for _, p := range v.PathCases {
		t.Run(p.Name, func(t *testing.T) {
			got, e := EncodeV3ExactOutputPath(p.Tokens, p.Fees)
			if e != nil || hexutil.Encode(got) != p.Expected {
				t.Fatal("path mismatch", e)
			}
		})
	}
	for _, p := range v.ZeroExPayloads {
		t.Run(p.Name, func(t *testing.T) {
			i := p.Input
			got, e := ValidateZeroExPayload(ZeroExPayloadOptions{Target: i.Target, Data: common.FromHex(i.Data), RouterAddress: i.RouterAddress, TokenOut: i.TokenOut, AmountOut: number(i.AmountOut), MaxInput: number(i.MaxInput), SettlerAddress: i.SettlerAddress, AllowanceHolderAddress: i.AllowanceHolderAddress, Now: number(i.Now)})
			if !p.Valid {
				if e == nil {
					t.Fatal("accepted invalid fixture")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			if hexutil.Encode(got.Data) != p.Expected.Data || got.NativeDeadline.String() != p.Expected.NativeDeadline {
				t.Fatal("canonical payload mismatch")
			}
		})
	}
	for _, p := range v.Payments {
		t.Run(p.Name, func(t *testing.T) {
			ctx := context.Background()
			rpc := newChain()
			c, e := NewClient(deployment(), rpc)
			if e != nil {
				t.Fatal(e)
			}
			call, e := c.BuildPaymentTransaction(ctx, v.CorePayment.call(), p.Quote.quote(), v.RefundRecipient)
			if e != nil {
				t.Fatal(e)
			}
			expected := p.Expected.call()
			if call.To != expected.To || call.Value.Cmp(expected.Value) != 0 || !bytes.Equal(call.Data, expected.Data) {
				t.Fatalf("complete calldata mismatch: %x", call.Data)
			}
			r, e := c.SimulatePayment(ctx, call, v.RefundRecipient)
			if e != nil {
				t.Fatal(e)
			}
			if r.ETHSpent.String() != p.Simulation.ETHSpent || r.ETHRefunded.String() != p.Simulation.ETHRefunded {
				t.Fatalf("simulation mismatch: %+v expected %+v", r, p.Simulation)
			}
		})
	}
}
