package protocol

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

type AddressRegistration struct {
	OwnerAddress [32]byte `json:"ownerAddress"`
	Keys         contracts.IAddressRegistryReceivingKeys
	Nonce        uint64
	Deadline     uint64
}
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
	To    common.Address
	Data  []byte
	Value *big.Int
}

func normalizeRegistration(c RegistryContext, r AddressRegistration, options RegistrationOptions) (AddressRegistration, error) {
	if err := CheckUint(c.ChainID, 64, "chainId"); err != nil {
		return r, err
	}
	if c.RegistryAddress == (common.Address{}) {
		return r, fmt.Errorf("registry address must be nonzero")
	}
	var identity fr.Element
	if err := identity.SetBytesCanonical(r.OwnerAddress[:]); err != nil || identity.IsZero() || new(big.Int).SetBytes(r.OwnerAddress[:]).BitLen() > 160 {
		return r, fmt.Errorf("ownerAddress must be a nonzero zero-padded EVM owner")
	}
	if _, err := primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: identity, ViewingPublicKey: r.Keys.ViewingPublicKey, ClueKey: r.Keys.ClueKey[:]}); err != nil {
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

// RegistrationTypedData constructs the registry's flat EIP-712 authorization.
func RegistrationTypedData(c RegistryContext, r AddressRegistration, options RegistrationOptions) (apitypes.TypedData, error) {
	r, err := normalizeRegistration(c, r, options)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	name := "RegisterReceivingKeys"
	if options.Update {
		name = "UpdateReceivingKeys"
	}
	chain := ethmath.HexOrDecimal256(*new(big.Int).Set(c.ChainID))
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}},
			name:           {{Name: "walletIdentity", Type: "bytes32"}, {Name: "viewingPublicKey", Type: "bytes"}, {Name: "clueKey", Type: "bytes32"}, {Name: "nonce", Type: "uint64"}, {Name: "deadline", Type: "uint64"}},
		},
		PrimaryType: name,
		Domain:      apitypes.TypedDataDomain{Name: "InvisiAddressRegistry", Version: "1", ChainId: &chain, VerifyingContract: c.RegistryAddress.Hex()},
		Message: apitypes.TypedDataMessage{
			"walletIdentity":   common.Hash(r.OwnerAddress).Hex(),
			"viewingPublicKey": "0x" + common.Bytes2Hex(r.Keys.ViewingPublicKey),
			"clueKey":          common.Hash(r.Keys.ClueKey).Hex(),
			"nonce":            new(big.Int).SetUint64(r.Nonce).String(),
			"deadline":         new(big.Int).SetUint64(r.Deadline).String(),
		},
	}, nil
}

func ComputeRegistrationDigest(c RegistryContext, r AddressRegistration, options RegistrationOptions) (common.Hash, error) {
	data, err := RegistrationTypedData(c, r, options)
	if err != nil {
		return common.Hash{}, err
	}
	digest, _, err := apitypes.TypedDataAndHash(data)
	return common.BytesToHash(digest), err
}

// BuildRegistrationCalldata encodes an ordinary EIP-712 or ERC-1271 signature.
// An empty signature is valid only when the transaction sender is the owner.
// Signature authentication is performed by the registry at execution time.
func BuildRegistrationCalldata(c RegistryContext, r AddressRegistration, signature []byte, options RegistrationOptions) (*RegistrationCalldata, error) {
	r, err := normalizeRegistration(c, r, options)
	if err != nil {
		return nil, err
	}
	parsed, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	method := "registerKeys"
	if options.Update {
		method = "updateKeys"
	}
	data, err := parsed.Pack(method, contracts.IAddressRegistryRegistration{WalletIdentity: r.OwnerAddress, Keys: r.Keys, Nonce: r.Nonce, Deadline: r.Deadline}, signature)
	if err != nil {
		return nil, err
	}
	return &RegistrationCalldata{To: c.RegistryAddress, Data: data, Value: new(big.Int)}, nil
}

// RegistryReadClient matches a generated caller bound to an explicitly selected registry.
