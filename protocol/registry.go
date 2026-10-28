package protocol

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

type AddressRegistration = contracts.IInvisiAddressRegistryRegistration
type RegistryContext struct {
	ChainID         *big.Int
	RegistryAddress common.Address
	Now             *big.Int
}
type RegistrationOptions struct {
	Update        bool
	ExpectedNonce *uint64
}
type RegistrationCalldata struct {
	To           common.Address
	Data         []byte
	Value        *big.Int
	PublicInputs [][32]byte
}

func normalizeRegistration(c RegistryContext, r AddressRegistration, options RegistrationOptions) (AddressRegistration, error) {
	if err := CheckUint(c.ChainID, 64, "chainId"); err != nil {
		return r, err
	}
	if c.RegistryAddress == (common.Address{}) {
		return r, fmt.Errorf("registry address must be nonzero")
	}
	var identity fr.Element
	if err := identity.SetBytesCanonical(r.WalletIdentity[:]); err != nil || identity.IsZero() {
		return r, fmt.Errorf("invalid complete wallet identity")
	}
	if _, err := core.EncodeLongAddress(core.ReceivingAddress{WalletIdentity: identity, ViewingPublicKey: r.Keys.ViewingPublicKey, ClueKey: r.Keys.ClueKey[:]}); err != nil {
		return r, err
	}
	if r.Nonce == ^uint64(0) {
		return r, fmt.Errorf("registry nonce exhausted")
	}
	if !options.Update && r.Nonce != 0 {
		return r, fmt.Errorf("initial registration requires nonce zero")
	}
	if options.ExpectedNonce != nil && *options.ExpectedNonce != r.Nonce {
		return r, fmt.Errorf("registry nonce mismatch")
	}
	if c.Now != nil {
		if err := CheckUint(c.Now, 64, "now"); err != nil {
			return r, err
		}
		if c.Now.Cmp(new(big.Int).SetUint64(r.Deadline)) > 0 {
			return r, fmt.Errorf("registration expired")
		}
	}
	r.Keys.ViewingPublicKey = append([]byte(nil), r.Keys.ViewingPublicKey...)
	return r, nil
}

// RegistrationAuthorizationRequest is the public seven-word private-wallet transcript.
func RegistrationAuthorizationRequest(c RegistryContext, r AddressRegistration, options RegistrationOptions) (core.WalletAuthorizationRequest, error) {
	r, err := normalizeRegistration(c, r, options)
	if err != nil {
		return core.WalletAuthorizationRequest{}, err
	}
	hash := crypto.Keccak256Hash(r.Keys.ViewingPublicKey, r.Keys.ClueKey[:])
	high := new(big.Int).SetBytes(hash[:16])
	low := new(big.Int).SetBytes(hash[16:])
	var identity fr.Element
	_ = identity.SetBytesCanonical(r.WalletIdentity[:])
	chain := core.BigIntToElement(c.ChainID)
	registry := core.BigIntToElement(c.RegistryAddress.Big())
	operation := core.WalletRegisterAddress
	if options.Update {
		operation = core.WalletUpdateAddress
	}
	return core.WalletAuthorizationRequest{Context: core.ProtocolContext{ChainID: chain, ProtocolAddress: registry}, Operation: operation, Capacity: 0, Words: []fr.Element{chain, registry, identity, core.ElementFromUint64(r.Nonce), core.ElementFromUint64(r.Deadline), core.BigIntToElement(high), core.BigIntToElement(low)}}, nil
}
func RegistrationTypedData(c RegistryContext, r AddressRegistration, options RegistrationOptions) (apitypes.TypedData, error) {
	request, err := RegistrationAuthorizationRequest(c, r, options)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return core.WalletAuthorizationTypedData(request)
}
func RegistrationPublicInputs(c RegistryContext, r AddressRegistration, options RegistrationOptions) ([][32]byte, error) {
	request, err := RegistrationAuthorizationRequest(c, r, options)
	if err != nil {
		return nil, err
	}
	digest, err := core.ComputeWalletAuthorizationDigest(request)
	if err != nil {
		return nil, err
	}
	high := core.BigIntToElement(new(big.Int).SetBytes(digest[:16]))
	low := core.BigIntToElement(new(big.Int).SetBytes(digest[16:]))
	return [][32]byte{r.WalletIdentity, high.Bytes(), low.Bytes()}, nil
}

// BuildRegistrationCalldata prepares a proof-authorized register/update call without broadcasting.
func BuildRegistrationCalldata(c RegistryContext, r AddressRegistration, proof []byte, options RegistrationOptions) (*RegistrationCalldata, error) {
	r, err := normalizeRegistration(c, r, options)
	if err != nil {
		return nil, err
	}
	if len(proof) == 0 {
		return nil, fmt.Errorf("registration proof is empty")
	}
	parsed, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	method := "registerKeys"
	if options.Update {
		method = "updateKeys"
	}
	data, err := parsed.Pack(method, r, proof)
	if err != nil {
		return nil, err
	}
	inputs, err := RegistrationPublicInputs(c, r, options)
	if err != nil {
		return nil, err
	}
	return &RegistrationCalldata{To: c.RegistryAddress, Data: data, Value: new(big.Int), PublicInputs: inputs}, nil
}

// RegistryReadClient matches a generated caller bound to an explicitly selected registry.
type RegistryReadClient interface {
	ReceivingKeysOf(*bind.CallOpts, [32]byte) (struct {
		Keys  contracts.IInvisiAddressRegistryReceivingKeys
		Nonce uint64
	}, error)
}

// CreateRegistryReceivingKeyLookup makes exactly one query and never falls back on RPC failure.
func CreateRegistryReceivingKeyLookup(client RegistryReadClient, options *bind.CallOpts) core.ReceivingKeyLookup {
	var copied bind.CallOpts
	if options != nil {
		copied = *options
		if options.BlockNumber != nil {
			copied.BlockNumber = new(big.Int).Set(options.BlockNumber)
		}
	}
	return func(identity fr.Element) (*core.ReceivingKeys, error) {
		if client == nil {
			return nil, fmt.Errorf("registry read client is required")
		}
		if _, err := core.EncodeShortAddress(identity); err != nil {
			return nil, err
		}
		callOptions := copied
		if copied.BlockNumber != nil {
			callOptions.BlockNumber = new(big.Int).Set(copied.BlockNumber)
		}
		result, err := client.ReceivingKeysOf(&callOptions, identity.Bytes())
		if err != nil {
			return nil, err
		}
		viewing := append([]byte(nil), result.Keys.ViewingPublicKey...)
		clue := append([]byte(nil), result.Keys.ClueKey[:]...)
		if _, err = core.EncodeLongAddress(core.ReceivingAddress{WalletIdentity: identity, ViewingPublicKey: viewing, ClueKey: clue}); err != nil {
			return nil, err
		}
		return &core.ReceivingKeys{ViewingPublicKey: viewing, ClueKey: clue}, nil
	}
}
