package sdk

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	"github.com/invisi-lab/invisi-sdk-go/prover"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

type AddressRegistry struct {
	context    protocol.RegistryContext
	deployment protocolrpc.ContractDeployment
	provider   Provider
}

func NewAddressRegistry(pool *Protocol, provider Provider) (*AddressRegistry, error) {
	if pool == nil || provider == nil || pool.Deployment().Registry == nil {
		return nil, errors.New("registry deployment and Chain RPC provider required")
	}
	d := pool.Deployment()
	chain, _ := d.ChainNumber()
	return &AddressRegistry{protocol.RegistryContext{ChainID: chain, RegistryAddress: d.Registry.Address}, *d.Registry, provider}, nil
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
	Prover        Prover
	Update        bool
	Deadline      uint64
	Authorization WalletAuthorizationProvider
}
type RegistrationPreview struct {
	RegistryAddress, WalletIdentity common.Address
	Update                          bool
	Nonce, Deadline                 uint64
	Status                          string
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
	authority       *core.WalletAuthorizationWitness
	bundle          *prover.ProofBundle
	data            []byte
	transactionHash common.Hash
}

func (r *AddressRegistry) PrepareRegistration(ctx context.Context, options RegistrationPreparation) (*RegistrationOperation, error) {
	if options.Wallet == nil {
		return nil, errors.New("wallet required")
	}
	if e := r.ValidateDeployment(ctx); e != nil {
		return nil, e
	}
	if options.Authorization == nil {
		if local, err := options.Wallet.ExportAuthorization(); err == nil {
			options.Authorization = local
		}
	}
	if options.Authorization == nil {
		return nil, errors.New("wallet authorization permission required")
	}
	if options.Authorization.WalletIdentity() != options.Wallet.identity {
		return nil, errors.New("registry authorization owner mismatch")
	}
	if len(options.Wallet.keys.ViewingPublicKey) != 33 || len(options.Wallet.keys.ClueKey) != 32 {
		return nil, errors.New("complete receiving keys required")
	}
	head, e := r.provider.HeaderByNumber(ctx, nil)
	if e != nil {
		return nil, e
	}
	if head == nil {
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
	registration := protocol.AddressRegistration{WalletIdentity: identity, Keys: contracts.IInvisiAddressRegistryReceivingKeys{ViewingPublicKey: append([]byte(nil), options.Wallet.keys.ViewingPublicKey...), ClueKey: clue}, Nonce: nonce, Deadline: deadline}
	context := r.context
	context.Now = new(big.Int).SetUint64(head.Time)
	return &RegistrationOperation{lifecycle: &transactionLifecycle{}, registry: r, wallet: options.Wallet, options: options, registration: registration, context: context, status: "prepared"}, nil
}
func (o *RegistrationOperation) Preview() RegistrationPreview {
	o.mu.Lock()
	defer o.mu.Unlock()
	return RegistrationPreview{o.registry.deployment.Address, o.wallet.identity, o.options.Update, o.registration.Nonce, o.registration.Deadline, o.status}
}
func (o *RegistrationOperation) String() string {
	return "RegistrationOperation([private witness redacted])"
}
func (o *RegistrationOperation) GoString() string { return o.String() }
func (o *RegistrationOperation) authorizeInternal(ctx context.Context) (*RegistrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.status != "prepared" {
		return nil, errors.New("registration must be prepared")
	}
	request, e := protocol.RegistrationAuthorizationRequest(o.context, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
	if e != nil {
		return nil, e
	}
	requestCopy := request
	requestCopy.Words = append([]fr.Element(nil), request.Words...)
	w, e := o.options.Authorization.Authorize(ctx, requestCopy)
	if e != nil {
		return nil, e
	}
	if e = verifyAuthorizationWitness(o.wallet.identity, request, w); e != nil {
		return nil, e
	}
	o.authority = &w
	o.status = "authorized"
	return o, nil
}
func (o *RegistrationOperation) proveInternal(ctx context.Context, options ...ProverOptions) (*RegistrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.status != "authorized" || o.authority == nil {
		return nil, errors.New("authorize registration before proving")
	}
	if o.options.Prover == nil {
		return nil, errors.New("local prover required")
	}
	if len(options) > 1 {
		return nil, errors.New("expected at most one prover option")
	}
	opt := ProverOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	words, e := protocol.RegistrationPublicInputs(o.context, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
	if e != nil {
		return nil, e
	}
	public := []common.Hash{}
	for _, w := range words {
		public = append(public, w)
	}
	request, e := prover.Request("wallet_registration", "evm", map[string]any{"public_key_x": o.authority.PublicKeyX, "public_key_y": o.authority.PublicKeyY, "signature": o.authority.Signature}, public)
	if e != nil {
		return nil, e
	}
	expected := prover.CloneStatement(request.ProofStatement)
	bundle, e := o.options.Prover.Prove(ctx, request, opt)
	if e != nil {
		return nil, e
	}
	bundle = prover.CloneBundle(bundle)
	if e = prover.AssertProofBundle(bundle, expected); e != nil {
		return nil, e
	}
	valid, e := o.options.Prover.Verify(ctx, prover.CloneBundle(bundle), prover.CloneStatement(expected), opt)
	if e != nil {
		return nil, e
	}
	if !valid {
		return nil, errors.New("registration proof verification failed")
	}
	call, e := protocol.BuildRegistrationCalldata(o.context, o.registration, bundle.Proof, protocol.RegistrationOptions{Update: o.options.Update})
	if e != nil {
		return nil, e
	}
	o.data = append([]byte(nil), call.Data...)
	o.bundle = &bundle
	o.authority = nil
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
	if o.bundle == nil || o.status == "prepared" || o.status == "authorized" {
		return TransactionCall{}, errors.New("proved registration required")
	}
	if err := o.registry.ValidateDeployment(ctx); err != nil {
		return TransactionCall{}, err
	}
	head, err := o.registry.provider.HeaderByNumber(ctx, nil)
	if err != nil {
		return TransactionCall{}, err
	}
	if head == nil || head.Time > o.registration.Deadline {
		return TransactionCall{}, errors.New("registration expired")
	}
	caller, err := contracts.NewInvisiAddressRegistryCaller(o.registry.deployment.Address, o.registry.provider)
	if err != nil {
		return TransactionCall{}, err
	}
	nonce, err := caller.NonceOf(&bind.CallOpts{Context: ctx, BlockNumber: head.Number}, o.registration.WalletIdentity)
	if err != nil {
		return TransactionCall{}, err
	}
	if nonce != o.registration.Nonce {
		return TransactionCall{}, errors.New("registry nonce changed; prepare again")
	}
	return TransactionCall{To: o.registry.deployment.Address, Data: append([]byte(nil), o.data...), Value: new(big.Int)}, nil
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
	return o, nil
}
func (o *RegistrationOperation) reconcileInternal(ctx context.Context) (operationConfirmation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	result := operationConfirmation{o.status, o.transactionHash}
	if o.transactionHash == (common.Hash{}) {
		return result, errors.New("registration is not submitted")
	}
	receipt, e := o.registry.provider.TransactionReceipt(ctx, o.transactionHash)
	if errors.Is(e, ethereum.NotFound) {
		return result, nil
	}
	if e != nil {
		return result, e
	}
	if receipt == nil {
		return result, nil
	}
	if receipt.TxHash != o.transactionHash || receipt.BlockNumber == nil || !receipt.BlockNumber.IsUint64() {
		return result, errors.New("registration receipt transaction or block mismatch")
	}
	header, e := o.registry.provider.HeaderByNumber(ctx, receipt.BlockNumber)
	if e != nil {
		return result, e
	}
	if header == nil || header.Number == nil || header.Number.Cmp(receipt.BlockNumber) != 0 || header.Hash() != receipt.BlockHash {
		return result, errors.New("registration receipt is not canonical")
	}
	reader, ok := o.registry.provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return result, errors.New("provider must authenticate registration calldata")
	}
	transaction, pending, e := reader.TransactionByHash(ctx, o.transactionHash)
	if e != nil {
		return result, e
	}
	if pending || transaction == nil || transaction.Hash() != o.transactionHash || transaction.To() == nil || *transaction.To() != o.registry.deployment.Address || common.Bytes2Hex(transaction.Data()) != common.Bytes2Hex(o.data) {
		return result, errors.New("registration transaction mismatch")
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		o.status = "failed"
	} else {
		if e = o.registry.ValidateDeployment(ctx); e != nil {
			return result, e
		}
		caller, e := contracts.NewInvisiAddressRegistryCaller(o.registry.deployment.Address, o.registry.provider)
		if e != nil {
			return result, e
		}
		keys, e := caller.ReceivingKeysOf(&bind.CallOpts{Context: ctx, BlockNumber: receipt.BlockNumber}, o.registration.WalletIdentity)
		if e != nil {
			return result, e
		}
		if keys.Nonce != o.registration.Nonce+1 || common.Bytes2Hex(keys.Keys.ViewingPublicKey) != common.Bytes2Hex(o.registration.Keys.ViewingPublicKey) || keys.Keys.ClueKey != o.registration.Keys.ClueKey {
			return result, errors.New("registration result differs from authorized keys")
		}
		o.status = "confirmed"
	}
	return operationConfirmation{o.status, o.transactionHash}, nil
}
