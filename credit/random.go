package credit

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

// GenerateSharedKey samples an independent, uniformly random nonzero field secret.
func GenerateSharedKey() (fr.Element, error) {
	for {
		key, err := core.GenerateNonce()
		if err != nil {
			return fr.Element{}, err
		}
		if !key.IsZero() {
			return key, nil
		}
	}
}
