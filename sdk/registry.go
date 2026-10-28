package sdk

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type AddressRegistry struct {
	context         protocol.RegistryContext
	deployment      protocolrpc.ContractDeployment
	provider        Provider
	protocolAddress common.Address
}

func NewAddressRegistry(pool *Protocol, provider Provider) (*AddressRegistry, error) {
	if pool == nil || provider == nil || pool.Deployment().Registry == nil {
		return nil, errors.New("registry deployment and Chain RPC provider required")
	}
	d := pool.Deployment()
	chain, _ := d.ChainNumber()
	return &AddressRegistry{context: protocol.RegistryContext{ChainID: chain, RegistryAddress: d.Registry.Address}, deployment: *d.Registry, provider: provider, protocolAddress: d.Protocol.Address}, nil
}
func (r *AddressRegistry) ValidateDeployment(ctx context.Context) error {
	chain, e := r.provider.ChainID(ctx)
	if e != nil {
		return e
	}
	if chain.Cmp(r.context.ChainID) != 0 {
		return errors.New("registry chain mismatch")
	}
	code, e := r.provider.CodeAt(ctx, r.deployment.Address, nil)
	if e != nil {
		return e
	}
	if len(code) == 0 || hashCode(code) != r.deployment.RuntimeCodeHash {
		return errors.New("registry runtime code mismatch")
	}
	caller, e := contracts.NewInvisiProtocolCaller(r.protocolAddress, r.provider)
	if e != nil {
		return e
	}
	selected, e := caller.AddressRegistry(&bind.CallOpts{Context: ctx})
	if e != nil {
		return e
	}
	if selected != r.deployment.Address {
		return errors.New("registry differs from immutable protocol binding")
	}
	return nil
}
func (r *AddressRegistry) Resolve(ctx context.Context, address string) (string, error) {
	recipient, e := core.ResolveAddress(address, func(identity fr.Element) (*core.ReceivingKeys, error) {
		if e := r.ValidateDeployment(ctx); e != nil {
			return nil, e
		}
		caller, e := contracts.NewInvisiAddressRegistryCaller(r.deployment.Address, r.provider)
		if e != nil {
			return nil, e
		}
		return protocol.CreateRegistryReceivingKeyLookup(caller, &bind.CallOpts{Context: ctx})(identity)
	})
	if e != nil {
		return "", e
	}
	return core.EncodeLongAddress(recipient)
}

type RegistrationPreparation struct {
	Wallet        *Wallet
	Update        bool
	Deadline      uint64
	Authorization RegistrationAuthorizationProvider
}
type RegistrationPreview struct {
	RegistryAddress, WalletIdentity common.Address
	Update                          bool
	Nonce, Deadline                 uint64
	Status                          string
	ReceivingKeys                   core.ReceivingKeys
	ReceiptBlock                    *protocolrpc.CanonicalBlock
}
type RegistrationOperation struct {
	lifecycle       *transactionLifecycle
	generation      uint64
	mu              sync.Mutex
	registry        *AddressRegistry
	wallet          *Wallet
	options         RegistrationPreparation
	registration    protocol.AddressRegistration
	context         protocol.RegistryContext
	status          string
	data            []byte
	transactionHash common.Hash
	receiptBlock    *protocolrpc.CanonicalBlock
}

func (r *AddressRegistry) PrepareRegistration(ctx context.Context, options RegistrationPreparation) (*RegistrationOperation, error) {
	if options.Wallet == nil {
		return nil, errors.New("wallet required")
	}
	if e := r.ValidateDeployment(ctx); e != nil {
		return nil, e
	}
	if options.Authorization == nil {
		if local, err := options.Wallet.ExportRegistrationAuthorization(); err == nil {
			options.Authorization = local
		}
	}
	if options.Authorization != nil && options.Authorization.WalletIdentity() != options.Wallet.identity {
		return nil, errors.New("registry authorization owner mismatch")
	}
	if len(options.Wallet.keys.ViewingPublicKey) != 33 || len(options.Wallet.keys.ClueKey) != 32 {
		return nil, errors.New("complete receiving keys required")
	}
	head, e := r.provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return nil, e
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return nil, errors.New("latest block unavailable")
	}
	deadline := options.Deadline
	if deadline == 0 {
		deadline = head.Time + 300
	}
	if deadline < head.Time {
		return nil, errors.New("registration deadline has expired")
	}
	caller, e := contracts.NewInvisiAddressRegistryCaller(r.deployment.Address, r.provider)
	if e != nil {
		return nil, e
	}
	identityField := core.BigIntToElement(options.Wallet.identity.Big())
	identity := identityField.Bytes()
	opts := &bind.CallOpts{Context: ctx, BlockNumber: head.Number}
	registered, e := caller.IsRegistered(opts, identity)
	if e != nil {
		return nil, e
	}
	if registered != options.Update {
		return nil, errors.New("registry registration/update state mismatch")
	}
	nonce, e := caller.NonceOf(opts, identity)
	if e != nil {
		return nil, e
	}
	var clue [32]byte
	copy(clue[:], options.Wallet.keys.ClueKey)
	registration := protocol.AddressRegistration{WalletIdentity: identity, Keys: contracts.IAddressRegistryReceivingKeys{ViewingPublicKey: append([]byte(nil), options.Wallet.keys.ViewingPublicKey...), ClueKey: clue}, Nonce: nonce, Deadline: deadline}
	context := r.context
	context.Now = new(big.Int).SetUint64(head.Time)
	if _, e = protocol.RegistrationTypedData(context, registration, protocol.RegistrationOptions{Update: options.Update}); e != nil {
		return nil, e
	}
	return &RegistrationOperation{lifecycle: &transactionLifecycle{}, registry: r, wallet: options.Wallet, options: options, registration: registration, context: context, status: "prepared"}, nil
}
func (o *RegistrationOperation) Preview() RegistrationPreview {
	o.mu.Lock()
	defer o.mu.Unlock()
	preview := RegistrationPreview{RegistryAddress: o.registry.deployment.Address, WalletIdentity: o.wallet.identity, Update: o.options.Update, Nonce: o.registration.Nonce, Deadline: o.registration.Deadline, Status: o.status, ReceivingKeys: core.ReceivingKeys{ViewingPublicKey: append([]byte(nil), o.registration.Keys.ViewingPublicKey...), ClueKey: append([]byte(nil), o.registration.Keys.ClueKey[:]...)}}
	if o.receiptBlock != nil {
		block := *o.receiptBlock
		preview.ReceiptBlock = &block
	}
	return preview
}
func (o *RegistrationOperation) String() string {
	return "RegistrationOperation([authorization redacted])"
}
func (o *RegistrationOperation) GoString() string { return o.String() }

// TypedData returns a fresh copy of the exact registration for external signing.
func (o *RegistrationOperation) TypedData() (apitypes.TypedData, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return protocol.RegistrationTypedData(o.context, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
}

func (o *RegistrationOperation) authorizeInternal(ctx context.Context, signatures ...[]byte) (*RegistrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.status != "prepared" {
		return nil, errors.New("registration must be prepared")
	}
	if len(signatures) > 1 {
		return nil, errors.New("expected at most one registration signature")
	}
	current, err := o.currentContext(ctx)
	if err != nil {
		return nil, err
	}
	data, err := protocol.RegistrationTypedData(current, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
	if err != nil {
		return nil, err
	}
	var signature []byte
	if len(signatures) == 1 {
		signature = append([]byte(nil), signatures[0]...)
	} else {
		if o.options.Authorization == nil {
			return nil, errors.New("registration signer or signature required")
		}
		// Signers receive a fresh transcript; they cannot mutate the frozen request.
		copy, err := protocol.RegistrationTypedData(current, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
		if err != nil {
			return nil, err
		}
		signature, err = o.options.Authorization.SignTypedData(ctx, copy)
		if err != nil {
			return nil, err
		}
		signature = append([]byte(nil), signature...)
	}
	if len(signature) == 0 {
		return nil, errors.New("registration signature required; direct owner calls use the low-level calldata builder")
	}
	// Normalize geth recovery IDs only after recovering this exact owner.
	// Other contract signatures are opaque and remain byte-for-byte unchanged.
	if len(signature) == 65 && signature[64] <= 1 {
		digest, _, err := apitypes.TypedDataAndHash(data)
		if err != nil {
			return nil, err
		}
		if owner, valid := recoverTypedDataSigner(digest, signature); valid && owner == o.wallet.identity {
			signature[64] += 27
		}
	}
	valid, err := verifyTypedData(ctx, o.registry.provider, o.wallet.identity, data, signature, "registration")
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid registration signature")
	}
	// Signing can take time (for example, a hardware-wallet confirmation).
	// Do not mark an operation ready after its nonce or deadline changes.
	current, err = o.currentContext(ctx)
	if err != nil {
		return nil, err
	}
	call, err := protocol.BuildRegistrationCalldata(current, o.registration, signature, protocol.RegistrationOptions{Update: o.options.Update})
	if err != nil {
		return nil, err
	}
	o.data = append([]byte(nil), call.Data...)
	o.status = "ready"
	return o, nil
}
func (o *RegistrationOperation) ExportTransaction(ctx context.Context) (TransactionCall, error) {
	if err := ctx.Err(); err != nil {
		return TransactionCall{}, err
	}
	o.lifecycle.mu.Lock()
	defer o.lifecycle.mu.Unlock()
	if o.lifecycle.busy || o.lifecycle.generation != o.generation {
		return TransactionCall{}, errors.New("registration handle is busy or stale")
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.data) == 0 || o.status == "prepared" {
		return TransactionCall{}, errors.New("authorized registration required")
	}
	if _, err := o.currentContext(ctx); err != nil {
		return TransactionCall{}, err
	}
	return TransactionCall{To: o.registry.deployment.Address, Data: append([]byte(nil), o.data...), Value: new(big.Int)}, nil
}

// currentContext rejects a stale signature before signing or exporting it.
func (o *RegistrationOperation) currentContext(ctx context.Context) (protocol.RegistryContext, error) {
	current := o.context
	if err := o.registry.ValidateDeployment(ctx); err != nil {
		return current, err
	}
	head, err := o.registry.provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return current, err
	}
	if head == nil || head.Number == nil || !head.Number.IsUint64() {
		return current, errors.New("latest block unavailable")
	}
	if head.Time > o.registration.Deadline {
		return current, errors.New("registration expired")
	}
	caller, err := contracts.NewInvisiAddressRegistryCaller(o.registry.deployment.Address, o.registry.provider)
	if err != nil {
		return current, err
	}
	nonce, err := caller.NonceOf(&bind.CallOpts{Context: ctx, BlockNumber: head.Number}, o.registration.WalletIdentity)
	if err != nil {
		return current, err
	}
	if nonce != o.registration.Nonce {
		return current, errors.New("registry nonce changed; prepare again")
	}
	canonical, err := o.registry.provider.HeaderByNumber(ctx, head.Number)
	if err != nil {
		return current, err
	}
	if canonical == nil || canonical.Number == nil || canonical.Number.Cmp(head.Number) != 0 || canonical.Hash() != head.Hash() {
		return current, errors.New("registration read block is not canonical")
	}
	current.Now = new(big.Int).SetUint64(head.Time)
	return current, nil
}

func (o *RegistrationOperation) trackInternal(ctx context.Context, hash common.Hash) (*RegistrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if hash == (common.Hash{}) || (o.status != "ready" && o.status != "uncertain") {
		return nil, errors.New("ready registration and transaction hash required")
	}
	o.transactionHash, o.status = hash, "uncertain"
	o.receiptBlock = nil
	return o, nil
}
func (o *RegistrationOperation) reconcileInternal(ctx context.Context) (operationConfirmation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	result := operationConfirmation{o.status, o.transactionHash}
	if o.transactionHash == (common.Hash{}) {
		return result, errors.New("registration is not submitted")
	}
	uncertain := func() (operationConfirmation, error) {
		o.status = "uncertain"
		o.receiptBlock = nil
		return operationConfirmation{o.status, o.transactionHash}, nil
	}
	receipt, err := o.registry.provider.TransactionReceipt(ctx, o.transactionHash)
	if errors.Is(err, ethereum.NotFound) {
		return uncertain()
	}
	if err != nil {
		return result, err
	}
	if receipt == nil {
		return uncertain()
	}
	if receipt.TxHash != o.transactionHash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		return result, errors.New("registration receipt transaction or block mismatch")
	}
	var receiptHeader *types.Header
	canonical := func() (bool, error) {
		header, err := o.registry.provider.HeaderByNumber(ctx, receipt.BlockNumber)
		if err != nil {
			return false, err
		}
		valid := header != nil && header.Number != nil && header.Number.Cmp(receipt.BlockNumber) == 0 && header.Hash() == receipt.BlockHash
		if valid {
			receiptHeader = types.CopyHeader(header)
		}
		return valid, nil
	}
	valid, err := canonical()
	if err != nil {
		return result, err
	}
	if !valid {
		return uncertain()
	}
	reader, ok := o.registry.provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return result, errors.New("provider must authenticate registration calldata")
	}
	transaction, pending, err := reader.TransactionByHash(ctx, o.transactionHash)
	if err != nil {
		return result, err
	}
	if pending || transaction == nil || transaction.Hash() != o.transactionHash || transaction.To() == nil || *transaction.To() != o.registry.deployment.Address || !bytes.Equal(transaction.Data(), o.data) || transaction.Value().Sign() != 0 {
		return result, errors.New("registration transaction mismatch")
	}
	status := "failed"
	if receipt.Status == types.ReceiptStatusSuccessful {
		if err = o.registry.ValidateDeployment(ctx); err != nil {
			return result, err
		}
		if err = o.verifyReceivingKeysEvent(receipt); err != nil {
			return result, err
		}
		status = "confirmed"
	}
	// RPC calls above can cross a reorganization. Recheck the receipt anchor
	// after authenticating calldata and the transaction's own event.
	valid, err = canonical()
	if err != nil {
		return result, err
	}
	if !valid {
		return uncertain()
	}
	o.status = status
	o.receiptBlock = &protocolrpc.CanonicalBlock{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash, ParentHash: receiptHeader.ParentHash, Timestamp: receiptHeader.Time}
	return operationConfirmation{o.status, o.transactionHash}, nil
}

func (o *RegistrationOperation) verifyReceivingKeysEvent(receipt *types.Receipt) error {
	parsed, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		return err
	}
	filterer, err := contracts.NewInvisiAddressRegistryFilterer(o.registry.deployment.Address, nil)
	if err != nil {
		return err
	}
	count := 0
	eventID := parsed.Events["ReceivingKeysSet"].ID
	for _, log := range receipt.Logs {
		if log == nil || log.Address != o.registry.deployment.Address || len(log.Topics) == 0 || log.Topics[0] != eventID {
			continue
		}
		if log.Removed || log.BlockHash != receipt.BlockHash || log.BlockNumber != receipt.BlockNumber.Uint64() || log.TxHash != receipt.TxHash || log.TxIndex != receipt.TransactionIndex || len(log.Topics) != 2 {
			return errors.New("registration event transaction or block mismatch")
		}
		event, err := filterer.ParseReceivingKeysSet(*log)
		if err != nil {
			return err
		}
		if event.WalletIdentity != o.registration.WalletIdentity || !bytes.Equal(event.ViewingPublicKey, o.registration.Keys.ViewingPublicKey) || event.ClueKey != o.registration.Keys.ClueKey || event.NextNonce != o.registration.Nonce+1 {
			return errors.New("registration event differs from authorized keys")
		}
		count++
	}
	if count != 1 {
		return errors.New("exactly one authorized receiving-key event required")
	}
	return nil
}
