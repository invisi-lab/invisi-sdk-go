package rpc

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

type RegistryReadClient interface {
	ReceivingKeysOf(*bind.CallOpts, [32]byte) (struct {
		Keys  contracts.IAddressRegistryReceivingKeys
		Nonce uint64
	}, error)
}

// CreateRegistryReceivingKeyLookup makes exactly one query and never falls back on RPC failure.
func CreateRegistryReceivingKeyLookup(client RegistryReadClient, options *bind.CallOpts) primitives.ReceivingKeyLookup {
	var copied bind.CallOpts
	if options != nil {
		copied = *options
		if options.BlockNumber != nil {
			copied.BlockNumber = new(big.Int).Set(options.BlockNumber)
		}
	}
	return func(identity fr.Element) (*primitives.ReceivingKeys, error) {
		if client == nil {
			return nil, fmt.Errorf("registry read client is required")
		}
		if _, err := primitives.EncodeShortAddress(identity); err != nil {
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
		if _, err = primitives.EncodeLongAddress(primitives.ReceivingAddress{OwnerAddress: identity, ViewingPublicKey: viewing, ClueKey: clue}); err != nil {
			return nil, err
		}
		return &primitives.ReceivingKeys{ViewingPublicKey: viewing, ClueKey: clue}, nil
	}
}
