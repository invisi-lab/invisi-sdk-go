package shared

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
)

// Sign uses ordinary geth EIP-712 signing and requires the environment key to match the owner/payer.
func Sign(request primitives.SigningRequest, expected common.Address, environment string) ([]byte, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(os.Getenv(environment), "0x"))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", environment, err)
	}
	if crypto.PubkeyToAddress(key.PublicKey) != expected {
		return nil, fmt.Errorf("%s must match required owner/payer %s", environment, expected.Hex())
	}
	data, err := request.TypedData()
	if err != nil {
		return nil, err
	}
	digest, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(digest, key)
}
