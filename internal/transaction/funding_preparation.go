package transaction

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func (s *Runtime) Funding(ctx context.Context, source *FundingSource, block protocolrpc.CanonicalBlock, deadline uint64) (protocol.FundingParams, string, error) {
	if source == nil {
		f, e := protocol.NormalizeFunding(protocol.FundingParams{})
		return f, "none", e
	}
	if e := Positive(source.Amount, 128); e != nil {
		return protocol.FundingParams{}, "", e
	}
	mode := source.Authorization
	if mode == "" {
		mode = "signature"
	}
	if mode != "signature" && mode != "caller" {
		return protocol.FundingParams{}, "", errors.New("invalid funding authorization mode")
	}
	if source.Payer == (common.Address{}) {
		return protocol.FundingParams{}, "", errors.New("funding payer must be nonzero")
	}
	if mode == "caller" && ((source.SigNonce != nil && source.SigNonce.Sign() != 0) || source.Deadline != 0) {
		return protocol.FundingParams{}, "", errors.New("caller funding uses zero nonce and deadline")
	}
	f := protocol.FundingParams{Payer: source.Payer, Amount: new(big.Int).Set(source.Amount)}
	if mode == "signature" {
		f.Deadline = new(big.Int).SetUint64(deadline)
		if source.Deadline != 0 {
			f.Deadline.SetUint64(source.Deadline)
		}
		if f.Deadline.Uint64() < block.Timestamp {
			return protocol.FundingParams{}, "", errors.New("funding authorization deadline expired")
		}
		if source.SigNonce != nil {
			f.SigNonce = new(big.Int).Set(source.SigNonce)
		} else {
			caller, err := contracts.NewInvisiProtocolCaller(s.Scope.ProtocolAddress, s.Options.Provider)
			if err != nil {
				return protocol.FundingParams{}, "", err
			}
			f.SigNonce, err = caller.Nonces(&bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(block.Number)}, source.Payer)
			if err != nil {
				return protocol.FundingParams{}, "", err
			}
		}
	}

	if source.Permit != nil {
		f.Permit = *source.Permit
	}
	f, e := protocol.NormalizeFunding(f)
	return f, mode, e
}
