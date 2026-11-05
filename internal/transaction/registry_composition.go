package transaction

import (
	"context"
	"errors"
)

// PrepareRegistration discovers the selected provider chain's pinned registry deployment.
// The resulting operation exports public signing requests and never owns a signer.
func (i *Invisi) PrepareRegistration(ctx context.Context, options RegistrationPreparation) (operation *RegistrationOperation, failure error) {
	defer func() { failure = WrapPreparation("prepareRegistration", failure) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if options.Wallet == nil {
		return nil, errors.New("public wallet required")
	}
	if i.options.Provider == nil {
		return nil, &MissingCapabilityError{Capability: "Chain RPC provider"}
	}
	if i.options.ProtocolRPC == nil {
		return nil, &MissingCapabilityError{Capability: "Protocol RPC deployment discovery"}
	}
	chain, err := i.options.Provider.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	if chain == nil || chain.Sign() <= 0 || !chain.IsUint64() {
		return nil, errors.New("invalid provider chain ID")
	}
	deployment, err := i.options.ProtocolRPC.GetDeployment(ctx, chain.Uint64())
	if err != nil {
		return nil, err
	}
	if deployment.Registry == nil {
		return nil, &MissingCapabilityError{Capability: "receiving-key registry deployment"}
	}
	pool, err := NewProtocol(deployment)
	if err != nil {
		return nil, err
	}
	registry, err := NewAddressRegistry(pool, i.options.Provider)
	if err != nil {
		return nil, err
	}
	return registry.PrepareRegistration(ctx, options)
}
