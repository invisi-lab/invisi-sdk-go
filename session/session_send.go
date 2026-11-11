package session

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"

	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

// TransactionOptions leaves the business destination, calldata, value and signer fixed.
type TransactionOptions struct {
	GasLimit                              uint64
	GasPrice, GasFeeCap, GasTipCap, Nonce *big.Int
}

type sessionSendState struct {
	mu      sync.Mutex
	sending map[string]bool
}

// SendTransaction delegates fees, nonce selection, signing and broadcasting to geth.
// It returns the original transaction without waiting. An Operation is exported and
// its hash tracked; plain TransactionCall values have no operation association.
func (s *WalletSession) SendTransaction(ctx context.Context, input any, settings ...TransactionOptions) (*types.Transaction, error) {
	if len(settings) > 1 {
		return nil, errors.New("expected at most one transaction options value")
	}
	if s.options.Signer == nil {
		return nil, errors.New("EVM transaction signer required")
	}
	signer := *s.options.Signer
	if signer.NoSend {
		return nil, errors.New("sending requires a broadcasting geth signer")
	}
	signer.Context = ctx
	copyNumber := func(value *big.Int) *big.Int {
		if value == nil {
			return nil
		}
		return new(big.Int).Set(value)
	}
	signer.GasPrice = copyNumber(signer.GasPrice)
	signer.GasFeeCap = copyNumber(signer.GasFeeCap)
	signer.GasTipCap = copyNumber(signer.GasTipCap)
	signer.Nonce = copyNumber(signer.Nonce)
	if len(settings) == 1 {
		options := settings[0]
		signer.GasLimit = options.GasLimit
		if options.GasPrice != nil {
			signer.GasPrice = copyNumber(options.GasPrice)
		}
		if options.GasFeeCap != nil {
			signer.GasFeeCap = copyNumber(options.GasFeeCap)
		}
		if options.GasTipCap != nil {
			signer.GasTipCap = copyNumber(options.GasTipCap)
		}
		if options.Nonce != nil {
			signer.Nonce = copyNumber(options.Nonce)
		}
	}
	if signer.GasLimit == 0 {
		return nil, errors.New("transaction gas limit must be supplied by the caller")
	}
	var operation *Operation
	var call sdk.TransactionCall
	switch value := input.(type) {
	case sdk.TransactionCall:
		call = value
	case *Operation:
		operation = value
		if operation == nil || operation.session.scope.Key() != s.scope.Key() {
			return nil, errors.New("operation session mismatch")
		}
		if s.sends == nil {
			return nil, errors.New("session send state missing")
		}
		s.sends.mu.Lock()
		if s.sends.sending[operation.ID()] {
			s.sends.mu.Unlock()
			return nil, errors.New("operation send already active")
		}
		s.sends.sending[operation.ID()] = true
		s.sends.mu.Unlock()
		defer func() { s.sends.mu.Lock(); delete(s.sends.sending, operation.ID()); s.sends.mu.Unlock() }()
		snapshot, err := s.options.Store.ReadSnapshot(ctx, s.scope)
		if err != nil {
			return nil, err
		}
		if record, ok := snapshot.Operations[operation.ID()]; !ok || record.WorkerID != operation.workerID {
			return nil, errors.New("operation does not belong to this session")
		}
		exported, err := operation.ExportTransaction(ctx)
		if err != nil {
			return nil, err
		}
		call = exported
	default:
		return nil, errors.New("expected public transaction call or managed operation")
	}
	call.Data = append([]byte(nil), call.Data...)
	if call.To != s.scope.ProtocolAddress {
		return nil, errors.New("transaction destination differs from session protocol")
	}
	signer.Value = new(big.Int)
	if call.Value != nil {
		signer.Value.Set(call.Value)
	}
	if signer.Value.Sign() != 0 {
		return nil, errors.New("Protocol transactions do not accept native value")
	}
	if err := s.protocol.ValidateDeployment(ctx, s.options.Provider); err != nil {
		return nil, err
	}
	backend, ok := s.options.Provider.(bind.ContractBackend)
	if !ok {
		return nil, errors.New("provider cannot broadcast transactions")
	}
	contract := bind.NewBoundContract(call.To, abi.ABI{}, backend, backend, backend)
	transaction, err := contract.RawTransact(&signer, call.Data)
	if err != nil {
		return nil, err
	}
	if operation != nil {
		if _, err = operation.TrackTransaction(context.WithoutCancel(ctx), transaction.Hash()); err != nil {
			return transaction, fmt.Errorf("transaction %s sent but tracking failed: %w", transaction.Hash(), err)
		}
	}
	return transaction, nil
}
