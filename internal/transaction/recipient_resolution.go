package transaction

import (
	"context"
	"errors"

	registryrpc "github.com/invisi-lab/invisi-sdk-go/rpc"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
)

func (s *Runtime) ResolveRecipient(ctx context.Context, address string) (primitives.ReceivingAddress, error) {
	return s.ResolveWithLookup(ctx, address, "protocol-rpc")
}

func (s *Runtime) ResolveWithLookup(ctx context.Context, address string, mode string) (primitives.ReceivingAddress, error) {
	decoded, err := primitives.DecodeAddress(address)
	if err != nil {
		return primitives.ReceivingAddress{}, &InvisiError{Code: InvalidRecipient, Address: address, Field: "to", Cause: err}
	}
	if decoded.Kind == "long" {
		return decoded.ReceivingAddress, nil
	}
	lookupFailure := func(err error) (primitives.ReceivingAddress, error) {
		return primitives.ReceivingAddress{}, &InvisiError{Code: RecipientLookupFailed, Address: address, Field: "to", Cause: err}
	}
	if err = s.Protocol.ValidateDeployment(ctx, s.Options.Provider); err != nil {
		return lookupFailure(err)
	}
	deployment := s.Protocol.Deployment()
	if deployment.Registry == nil {
		return primitives.ReceivingAddress{}, &MissingCapabilityError{Capability: "receiving-key registry deployment"}
	}
	var keys *primitives.ReceivingKeys
	if mode == "" {
		mode = "protocol-rpc"
	}
	switch mode {
	case "protocol-rpc":
		registered, err := s.RPC.GetReceivingAddress(ctx, common.BigToAddress(primitives.ElementToBigInt(decoded.OwnerAddress)))
		if err != nil {
			return lookupFailure(err)
		}
		if registered != nil {
			found, err := primitives.DecodeAddress(registered.ReceivingAddress)
			if err != nil || found.Kind != "long" || found.OwnerAddress != decoded.OwnerAddress {
				if err == nil {
					err = errors.New("registered receiving identity mismatch")
				}
				return lookupFailure(err)
			}
			keys = &primitives.ReceivingKeys{ViewingPublicKey: found.ViewingPublicKey, ClueKey: found.ClueKey}
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
		keys, err = registryrpc.CreateRegistryReceivingKeyLookup(caller, &bind.CallOpts{Context: ctx})(decoded.OwnerAddress)
		if err != nil {
			return lookupFailure(err)
		}
	default:
		return lookupFailure(errors.New("address lookup must be protocol-rpc or chain"))
	}
	if keys == nil {
		return primitives.ReceivingAddress{}, &InvisiError{Code: EVMAddressNotRegistered, Address: address, Field: "to", Cause: errors.New("EVM address has no registered receiving keys")}
	}
	result, err := primitives.ResolveAddress(address, func(fr.Element) (*primitives.ReceivingKeys, error) { return keys, nil })
	if err != nil {
		return lookupFailure(err)
	}
	return result, nil
}
