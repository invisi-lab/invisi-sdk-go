package transaction

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func (o *Execution) Sources() []protocol.InputSource {
	p := o.Plan
	if p.Kind == "transact" {
		return p.Params.Sources
	}
	return nil
}

func (o *Execution) ValidateExport(ctx context.Context) error {
	if err := o.Runtime.Protocol.ValidateDeployment(ctx, o.Runtime.Options.Provider); err != nil {
		return err
	}
	head, err := o.Runtime.Options.Provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}
	if head == nil || head.Number == nil {
		return errors.New("latest block unavailable")
	}
	if o.Plan.Broadcaster != nil {
		deadline, ok := new(big.Int).SetString(o.Plan.Broadcaster.Signed.Quote.Deadline, 10)
		if !ok || deadline.Cmp(new(big.Int).SetUint64(head.Time)) < 0 || deadline.Cmp(big.NewInt(time.Now().Unix())) < 0 {
			return errors.New("broadcaster quote expired")
		}
	}
	if OperationExpired(o.Plan, head.Time) {
		return errors.New("operation has expired")
	}
	caller, err := contracts.NewInvisiProtocolCaller(o.Plan.Context.ProtocolAddress, o.Runtime.Options.Provider)
	if err != nil {
		return err
	}
	options := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
	if o.Plan.FundingMode == "signature" {
		if o.Plan.Funding.Deadline == nil || new(big.Int).SetUint64(head.Time).Cmp(o.Plan.Funding.Deadline) > 0 {
			return errors.New("funding authorization expired")
		}
		nonce, err := caller.Nonces(options, o.Plan.Funding.Payer)
		if err != nil {
			return err
		}
		if nonce.Cmp(o.Plan.Funding.SigNonce) != 0 {
			return errors.New("funding nonce changed; prepare again")
		}
	}
	for _, source := range o.Sources() {
		known, err := caller.RootHistory(options, source.TreeId, source.MerkleRoot)
		if err != nil {
			return err
		}
		if !known {
			return errors.New("source root is no longer accepted; prepare again")
		}
	}
	nullifiers := []common.Hash{}
	for _, n := range o.Plan.Inputs {
		nullifiers = append(nullifiers, n.Nullifier)
	}

	for _, nf := range nullifiers {
		spent, err := caller.IsNullifierSpent(options, nf)
		if err != nil {
			return err
		}
		if spent {
			return errors.New("input already spent")
		}
	}
	return ctx.Err()
}

func OperationExpired(p Plan, now uint64) bool {
	if now > p.Deadline {
		return true
	}

	return false
}
