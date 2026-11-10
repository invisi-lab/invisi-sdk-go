package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

// PaymentNoteVerificationOptions can override public connections captured by ProtocolFromRPC.
type PaymentNoteVerificationOptions struct {
	ProtocolRPC   *protocolrpc.ProtocolRpcClient
	Provider      protocolrpc.ChainProvider
	Confirmations uint64
}
type VerifiedPaymentNote struct {
	membership protocolrpc.Membership
	checkedAt  protocolrpc.CanonicalBlock
	valid      bool
}
type SpendablePaymentNote struct {
	verified              VerifiedPaymentNote
	input                 NoteInput
	spendabilityCheckedAt protocolrpc.CanonicalBlock
}

func (v *VerifiedPaymentNote) Membership() protocolrpc.Membership {
	m := v.membership
	m.Path = append([]common.Hash(nil), m.Path...)
	return m
}
func (v *VerifiedPaymentNote) CheckedAt() protocolrpc.CanonicalBlock  { return v.checkedAt }
func (v *SpendablePaymentNote) Membership() protocolrpc.Membership    { return v.verified.Membership() }
func (v *SpendablePaymentNote) CheckedAt() protocolrpc.CanonicalBlock { return v.verified.CheckedAt() }
func (v *SpendablePaymentNote) SpendabilityCheckedAt() protocolrpc.CanonicalBlock {
	return v.spendabilityCheckedAt
}
func (v *SpendablePaymentNote) ToInput() (NoteInput, error) {
	if v == nil || !v.verified.valid {
		return NoteInput{}, errors.New("spendability verification required")
	}
	n := v.input
	n.Note.Payload = append([]byte(nil), n.Note.Payload...)
	return n, nil
}
func paymentVerificationOptions(pool *Protocol, options []PaymentNoteVerificationOptions) (PaymentNoteVerificationOptions, error) {
	var o PaymentNoteVerificationOptions
	if pool == nil || len(options) > 1 {
		return o, errors.New("Protocol and at most one verification option required")
	}
	if len(options) == 1 {
		o = options[0]
	}
	rpc, provider, _ := pool.VerificationConnections()
	if o.ProtocolRPC == nil {
		o.ProtocolRPC = rpc
	}
	if o.Provider == nil {
		o.Provider = provider
	}
	if o.ProtocolRPC == nil || o.Provider == nil {
		return o, errors.New("verification requires ProtocolFromRPC or explicit public connections")
	}
	if o.Confirmations == 0 {
		o.Confirmations = 1
	}
	return o, nil
}
func paymentBlock(ctx context.Context, provider protocolrpc.ChainProvider, number *big.Int) (protocolrpc.CanonicalBlock, error) {
	h, e := provider.HeaderByNumber(ctx, number)
	if e != nil {
		return protocolrpc.CanonicalBlock{}, e
	}
	if h == nil || h.Number == nil || !h.Number.IsUint64() || (number != nil && h.Number.Cmp(number) != 0) {
		return protocolrpc.CanonicalBlock{}, errors.New("invalid verification block")
	}
	return protocolrpc.CanonicalBlock{Number: h.Number.Uint64(), Hash: h.Hash(), ParentHash: h.ParentHash, Timestamp: h.Time}, nil
}
func paymentBarrier(ctx context.Context, provider protocolrpc.ChainProvider, block protocolrpc.CanonicalBlock) error {
	current, e := paymentBlock(ctx, provider, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return e
	}
	if current != block {
		return errors.New("payment note verification block reorganized")
	}
	return nil
}
func paymentBool(ctx context.Context, pool *Protocol, provider protocolrpc.ChainProvider, block protocolrpc.CanonicalBlock, method string, args ...any) (bool, error) {
	data, e := protocol.PackProtocolCalldata(method, args...)
	if e != nil {
		return false, e
	}
	address := pool.Context().ProtocolAddress
	raw, e := provider.CallContract(ctx, ethereum.CallMsg{To: &address, Data: data}, new(big.Int).SetUint64(block.Number))
	if e != nil {
		return false, e
	}
	if len(raw) != 32 || new(big.Int).SetBytes(raw).Cmp(big.NewInt(1)) > 0 {
		return false, errors.New("invalid chain boolean")
	}
	return raw[31] == 1, nil
}
func (n *PaymentNote) Verify(ctx context.Context, pool *Protocol, options ...PaymentNoteVerificationOptions) (*VerifiedPaymentNote, error) {
	result, _, _, e := n.verify(ctx, pool, nil, options)
	return result, e
}
func (n *PaymentNote) VerifySpendable(ctx context.Context, pool *Protocol, wallet *Wallet, options ...PaymentNoteVerificationOptions) (*SpendablePaymentNote, error) {
	if wallet == nil {
		return nil, errors.New("wallet spending authorization capability required")
	}
	result, input, current, e := n.verify(ctx, pool, wallet, options)
	if e != nil {
		return nil, e
	}
	return &SpendablePaymentNote{*result, input, current}, nil
}
func (n *PaymentNote) verify(ctx context.Context, pool *Protocol, wallet *Wallet, options []PaymentNoteVerificationOptions) (*VerifiedPaymentNote, NoteInput, protocolrpc.CanonicalBlock, error) {
	var empty NoteInput
	var zero protocolrpc.CanonicalBlock
	fail := func(e error) (*VerifiedPaymentNote, NoteInput, protocolrpc.CanonicalBlock, error) {
		return nil, empty, zero, e
	}
	if n == nil || n.ipn == "" {
		return fail(errors.New("uninitialized payment note"))
	}
	o, e := paymentVerificationOptions(pool, options)
	if e != nil {
		return fail(e)
	}
	if n.data.ChainID.Uint64() != pool.Context().ChainID || n.data.ProtocolAddress != pool.Context().ProtocolAddress {
		return fail(errors.New("payment note protocol context mismatch"))
	}
	if e = pool.ValidateDeployment(ctx, o.Provider); e != nil {
		return fail(e)
	}
	if wallet != nil && wallet.OwnerAddress() != n.data.OwnerAddress {
		return fail(errors.New("payment note owner mismatch"))
	}
	latest, e := paymentBlock(ctx, o.Provider, nil)
	if e != nil {
		return fail(e)
	}
	if latest.Number < o.Confirmations-1 {
		return fail(errors.New("insufficient confirmed blocks"))
	}
	number := latest.Number - (o.Confirmations - 1)
	if number < pool.Deployment().Protocol.FromBlock {
		return fail(errors.New("insufficient confirmed blocks"))
	}
	rpc := o.ProtocolRPC.WithProvider(o.Provider)
	status, e := rpc.GetStatus(ctx, pool.Context())
	if e != nil {
		return fail(e)
	}
	if status.IndexedHead == nil {
		return fail(errors.New("Protocol RPC has no indexed blocks"))
	}
	if status.IndexedHead.Number < number {
		number = status.IndexedHead.Number
	}
	at, e := paymentBlock(ctx, o.Provider, new(big.Int).SetUint64(number))
	if e != nil {
		return fail(e)
	}
	lookup, e := rpc.GetNote(ctx, pool.Context(), n.data.Commitment, at)
	if e != nil {
		return fail(e)
	}
	if lookup == nil {
		return fail(errors.New("payment note is not indexed at the confirmed block"))
	}
	output, m := lookup.Output, lookup.Membership
	if output.TokenAddress != n.data.TokenAddress || output.Commitment != n.data.Commitment || m.Leaf != n.data.Commitment || m.TreeID != output.TreeID || m.LeafIndex != output.LeafIndex || len(m.Path) != primitives.TreeDepth {
		return fail(errors.New("payment note membership metadata mismatch"))
	}
	// RPC has independently verified the Merkle path. Recheck chain state for this operation.
	known, e := paymentBool(ctx, pool, o.Provider, at, "rootHistory", new(big.Int).SetUint64(m.TreeID), m.Root)
	if e != nil {
		return fail(e)
	}
	if !known {
		return fail(errors.New("payment note root is not accepted on chain"))
	}
	if e = paymentBarrier(ctx, o.Provider, output.Block); e != nil {
		return fail(e)
	}
	var spendAt protocolrpc.CanonicalBlock
	if wallet != nil {
		spendAt, e = paymentBlock(ctx, o.Provider, nil)
		if e != nil {
			return fail(e)
		}
		note := paymentNoteOpening(n.data)
		nf, e := primitives.ComputeNoteNullifier(note.ChainID, note.ProtocolAddress, note.Nonce, primitives.BigIntToElement(n.data.Commitment.Big()))
		if e != nil {
			return fail(e)
		}
		spent, e := paymentBool(ctx, pool, o.Provider, spendAt, "isNullifierSpent", FieldHash(nf))
		if e != nil {
			return fail(e)
		}
		if spent {
			return fail(errors.New("payment note is already spent"))
		}
		known, e = paymentBool(ctx, pool, o.Provider, spendAt, "rootHistory", new(big.Int).SetUint64(m.TreeID), m.Root)
		if e != nil {
			return fail(e)
		}
		if !known {
			return fail(errors.New("payment note root is no longer spendable"))
		}
		if e = paymentBarrier(ctx, o.Provider, spendAt); e != nil {
			return fail(e)
		}
	}
	if e = paymentBarrier(ctx, o.Provider, at); e != nil {
		return fail(e)
	}
	m.Path = append([]common.Hash(nil), m.Path...)
	return &VerifiedPaymentNote{m, at, true}, NoteInput{Note: paymentNoteOpening(n.data), Commitment: n.data.Commitment, TreeID: m.TreeID, LeafIndex: m.LeafIndex}, spendAt, nil
}
