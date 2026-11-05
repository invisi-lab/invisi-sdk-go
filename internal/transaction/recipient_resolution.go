package transaction

import (
	"context"
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func (s *Runtime) ResolveRecipient(ctx context.Context, address string) (core.ReceivingAddress, error) {
	return s.ResolveWithLookup(ctx, address, "protocol-rpc")
}

func (s *Runtime) ResolveWithLookup(ctx context.Context, address string, mode string) (core.ReceivingAddress, error) {
	decoded, err := core.DecodeAddress(address)
	if err != nil {
		return core.ReceivingAddress{}, &InvisiError{Code: InvalidRecipient, Address: address, Field: "to", Cause: err}
	}
	if decoded.Kind == "long" {
		return decoded.ReceivingAddress, nil
	}
	lookupFailure := func(err error) (core.ReceivingAddress, error) {
		return core.ReceivingAddress{}, &InvisiError{Code: RecipientLookupFailed, Address: address, Field: "to", Cause: err}
	}
	if err = s.Protocol.ValidateDeployment(ctx, s.Options.Provider); err != nil {
		return lookupFailure(err)
	}
	deployment := s.Protocol.Deployment()
	if deployment.Registry == nil {
		return core.ReceivingAddress{}, &MissingCapabilityError{Capability: "receiving-key registry deployment"}
	}
	var keys *core.ReceivingKeys
	if mode == "" {
		mode = "protocol-rpc"
	}
	switch mode {
	case "protocol-rpc":
		registered, err := s.RPC.GetReceivingAddress(ctx, common.BigToAddress(core.ElementToBigInt(decoded.OwnerAddress)))
		if err != nil {
			return lookupFailure(err)
		}
		if registered != nil {
			found, err := core.DecodeAddress(registered.ReceivingAddress)
			if err != nil || found.Kind != "long" || found.OwnerAddress != decoded.OwnerAddress {
				if err == nil {
					err = errors.New("registered receiving identity mismatch")
				}
				return lookupFailure(err)
			}
			keys = &core.ReceivingKeys{ViewingPublicKey: found.ViewingPublicKey, ClueKey: found.ClueKey}
		}
	case "chain":
		code, err := s.Options.Provider.CodeAt(ctx, deployment.Registry.Address, nil)
		if err != nil {
			return lookupFailure(err)
		}
		if HashCode(code) != deployment.Registry.RuntimeCodeHash {
			return lookupFailure(errors.New("registry runtime mismatch"))
		}
		caller, err := contracts.NewInvisiAddressRegistryCaller(deployment.Registry.Address, s.Options.Provider)
		if err != nil {
			return lookupFailure(err)
		}
		keys, err = protocol.CreateRegistryReceivingKeyLookup(caller, &bind.CallOpts{Context: ctx})(decoded.OwnerAddress)
		if err != nil {
			return lookupFailure(err)
		}
	default:
		return lookupFailure(errors.New("address lookup must be protocol-rpc or chain"))
	}
	if keys == nil {
		return core.ReceivingAddress{}, &InvisiError{Code: EVMAddressNotRegistered, Address: address, Field: "to", Cause: errors.New("EVM address has no registered receiving keys")}
	}
	result, err := core.ResolveAddress(address, func(fr.Element) (*core.ReceivingKeys, error) { return keys, nil })
	if err != nil {
		return lookupFailure(err)
	}
	return result, nil
}
