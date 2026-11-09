package transaction

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	if chain == nil || chain.Cmp(r.context.ChainID) != 0 {
		return errors.New("registry chain mismatch")
	}
	code, e := r.provider.CodeAt(ctx, r.deployment.Address, nil)
	if e != nil {
		return e
	}
	if len(code) == 0 || HashCode(code) != r.deployment.RuntimeCodeHash {
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
	Wallet   *Wallet
	Update   bool
	Deadline uint64
}

type RegistrationPreview struct {
	RegistryAddress, OwnerAddress common.Address
	Update                        bool
	Nonce, Deadline               uint64
	Status                        string
	ReceivingKeys                 core.ReceivingKeys
	ReceiptBlock                  *protocolrpc.CanonicalBlock
}

type RegistrationOperation struct {
	lifecycle       *TransactionLifecycle
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
	direct          bool
	receiptBlock    *protocolrpc.CanonicalBlock
}

func (r *AddressRegistry) PrepareRegistration(ctx context.Context, options RegistrationPreparation) (operation *RegistrationOperation, failure error) {
	defer func() { failure = WrapPreparation("prepareRegistration", failure) }()
	if options.Wallet == nil {
		return nil, errors.New("wallet required")
	}
	if e := r.ValidateDeployment(ctx); e != nil {
		return nil, e
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
		if head.Time > ^uint64(0)-300 {
			return nil, errors.New("registration deadline exceeds uint64")
		}
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
		if options.Update {
			return nil, errors.New("wallet is not registered")
		}
		return nil, errors.New("wallet is already registered; prepare an update")
	}
	nonce, e := caller.NonceOf(opts, identity)
	if e != nil {
		return nil, e
	}
	var clue [32]byte
	copy(clue[:], options.Wallet.keys.ClueKey)
	registration := protocol.AddressRegistration{OwnerAddress: identity, Keys: contracts.IAddressRegistryReceivingKeys{ViewingPublicKey: append([]byte(nil), options.Wallet.keys.ViewingPublicKey...), ClueKey: clue}, Nonce: nonce, Deadline: deadline}
	context := r.context
	context.Now = new(big.Int).SetUint64(head.Time)
	if _, e = protocol.RegistrationTypedData(context, registration, protocol.RegistrationOptions{Update: options.Update}); e != nil {
		return nil, e
	}
	operation = &RegistrationOperation{lifecycle: &TransactionLifecycle{}, registry: r, wallet: options.Wallet, options: options, registration: registration, context: context, status: "prepared"}
	if _, err := operation.currentContext(ctx); err != nil {
		return nil, err
	}
	return operation, nil
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

func (o *RegistrationOperation) withSignatureInternal(ctx context.Context, suppliedSignature []byte) (*RegistrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.status != "prepared" {
		return nil, errors.New("registration must be prepared")
	}
	current, err := o.currentContext(ctx)
	if err != nil {
		return nil, err
	}
	data, err := protocol.RegistrationTypedData(current, o.registration, protocol.RegistrationOptions{Update: o.options.Update})
	if err != nil {
		return nil, err
	}
	signature := append([]byte(nil), suppliedSignature...)
	if len(signature) == 0 {
		return nil, errors.New("registration signature required; use AuthorizeDirect for an owner-submitted call")
	}
	// Normalize geth recovery IDs only after recovering this exact owner.
	// Other contract signatures are opaque and remain byte-for-byte unchanged.
	if len(signature) == 65 && signature[64] <= 1 {
		digest, _, err := apitypes.TypedDataAndHash(data)
		if err != nil {
			return nil, err
		}
		if owner, valid := RecoverTypedDataSigner(digest, signature); valid && owner == o.wallet.identity {
			signature[64] += 27
		}
	}
	valid, err := VerifyTypedData(ctx, o.registry.provider, o.wallet.identity, data, signature, "registration")
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
