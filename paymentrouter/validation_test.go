package paymentrouter

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"math/big"
	"testing"
)

func TestCoreStatementAdversarialShapes(t *testing.T) {
	base := coreCall(t)
	decoded, e := coreMethod.Inputs.Unpack(base.Data[4:])
	if e != nil {
		t.Fatal(e)
	}
	for _, change := range []string{"proof", "payer", "amount", "nonce", "fundingDeadline", "signature", "permitDeadline", "permitV", "permitR", "permitS", "source", "nullifier", "withdrawal", "withdrawalAddress", "weth", "oneOutput", "duplicate", "zeroNote", "field", "deliveryCount", "ciphertext", "fee", "expired", "wideDeadline"} {
		t.Run(change, func(t *testing.T) {
			r := *abi.ConvertType(decoded[1], new(protocol.TransactRequest)).(*protocol.TransactRequest)
			proof := []byte{1}
			r.Params.NoteCommitments = append([][32]byte(nil), r.Params.NoteCommitments...)
			r.Deliveries = append([]protocol.Delivery(nil), r.Deliveries...)
			switch change {
			case "proof":
				proof = nil
			case "payer":
				r.Funding.Payer = addr(10)
			case "amount":
				r.Funding.Amount = new(big.Int).Lsh(bi(1), 128)
			case "nonce":
				r.Funding.SigNonce = bi(1)
			case "fundingDeadline":
				r.Funding.Deadline = bi(1)
			case "signature":
				r.Funding.Signature = []byte{1}
			case "permitDeadline":
				r.Funding.Permit.Deadline = bi(1)
			case "permitV":
				r.Funding.Permit.V = 1
			case "permitR":
				r.Funding.Permit.R[31] = 1
			case "permitS":
				r.Funding.Permit.S[31] = 1
			case "source":
				r.Params.Sources = []protocol.InputSource{{TreeId: bi(1), MerkleRoot: [32]byte{31: 1}}}
			case "nullifier":
				r.Params.Nullifiers = [][32]byte{{31: 1}}
			case "withdrawal":
				r.Params.WithdrawalAmount = bi(1)
			case "withdrawalAddress":
				r.Params.WithdrawalAddress = addr(10)
			case "weth":
				r.Params.Token = addr(4)
			case "oneOutput":
				r.Params.NoteCommitments = r.Params.NoteCommitments[:1]
				r.Deliveries = r.Deliveries[:1]
			case "duplicate":
				r.Params.NoteCommitments[1] = r.Params.NoteCommitments[0]
			case "zeroNote":
				r.Params.NoteCommitments[0] = [32]byte{}
			case "field":
				r.Params.NoteCommitments[0] = [32]byte{0: 255}
			case "deliveryCount":
				r.Deliveries = r.Deliveries[:1]
			case "ciphertext":
				r.Deliveries[0].Ciphertext = []byte{1}
			case "fee":
				r.Params.FeeAmount = bi(1)
			case "expired":
				r.Params.Deadline = bi(999)
			case "wideDeadline":
				r.Params.Deadline = new(big.Int).Lsh(bi(1), 64)
			}
			data, e := coreMethod.Inputs.Pack(proof, r)
			if e != nil {
				t.Fatal(e)
			}
			call := base
			call.Data = append(append([]byte(nil), coreMethod.ID...), data...)
			if _, _, e = InspectCorePayment(call, deployment(), bi(1000)); e == nil {
				t.Fatal("accepted invalid core statement")
			}
		})
	}
}

type codeProvider struct {
	Provider
	empty common.Address
}

func (p codeProvider) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	if a == p.empty {
		return nil, nil
	}
	return p.Provider.CodeAt(ctx, a, n)
}
func TestDeploymentAuthentication(t *testing.T) {
	for _, change := range []string{"chain", "duplicate", "zero", "code", "hash", "getter"} {
		t.Run(change, func(t *testing.T) {
			d := deployment()
			rpc := newChain()
			var p Provider = rpc
			switch change {
			case "chain":
				d.ChainID = bi(0)
			case "duplicate":
				d.CoreAddress = d.RouterAddress
			case "zero":
				d.WETHAddress = common.Address{}
			case "code":
				p = codeProvider{rpc, addr(4)}
			case "hash":
				h := common.Hash{31: 1}
				d.RuntimeCodeHash = &h
			case "getter":
				rpc.core = addr(5)
			}
			c, e := NewClient(d, p)
			if e != nil {
				return
			}
			if _, e = frame(context.Background(), c.provider, c.deployment, false); e == nil {
				t.Fatal("accepted invalid deployment")
			}
		})
	}
}
func TestSimulationRejectsBudgetAndPropagatesRevert(t *testing.T) {
	ctx := context.Background()
	rpc := newChain()
	c, _ := NewClient(deployment(), rpc)
	a, _ := NewUniswapV3Adapter(rpc, addr(9), []V3Route{{[]common.Address{addr(4), addr(2)}, []uint32{3000}}})
	q, e := c.Quote(ctx, a, parameters())
	if e != nil {
		t.Fatal(e)
	}
	call, e := c.BuildPaymentTransaction(ctx, coreCall(t), q, addr(10))
	if e != nil {
		t.Fatal(e)
	}
	call.Value = bi(59)
	if _, e = c.SimulatePayment(ctx, call, addr(10)); e == nil {
		t.Fatal("accepted overspend")
	}
	call.Value = bi(61)
	rpc.fail = true
	if _, e = c.SimulatePayment(ctx, call, addr(10)); e == nil {
		t.Fatal("ignored revert")
	}
}
