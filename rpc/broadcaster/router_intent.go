package broadcaster

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"math/big"
)

type relayPermit struct {
	Deadline *big.Int
	V        uint8
	R, S     [32]byte
}

// NormalizeRouterEnvelope removes only proof and wallet authorization bytes.
// The input must be strict canonical ABI; business fields remain bound.
func NormalizeRouterEnvelope(env RouterEnvelope) (RouterEnvelope, error) {
	out := env
	out.AuthorizationList = append(env.AuthorizationList[:0:0], env.AuthorizationList...)
	for n := range out.AuthorizationList {
		out.AuthorizationList[n].V = 0
		out.AuthorizationList[n].R = uint256.Int{}
		out.AuthorizationList[n].S = uint256.Int{}
	}
	inner := []byte(env.Data)
	var outer []any
	native := len(inner) >= 4 && bytes.Equal(inner[:4], relayMethods["executeNativePayment"].ID)
	if native {
		var e error
		outer, e = unpackRelay("executeNativePayment", inner)
		if e != nil {
			return out, e
		}
		inner = outer[0].([]byte)
	} else if len(env.AuthorizationList) != 0 {
		return out, fmt.Errorf("unexpected token delegation authorization")
	}
	name := "payTokenFor"
	if native {
		name = "payNative"
	}
	args, e := unpackRelay(name, inner)
	if e != nil {
		return out, e
	}
	i := abi.ConvertType(args[2], new(relayIntent)).(*relayIntent)
	// Validate current hashes before replacing the proof-dependent hash.
	current, e := routerHashArgs(abi.Arguments{routerKind("bytes32"), relayMethods["transact"].Inputs[1]}, crypto.Keccak256Hash(args[0].([]byte)), args[1])
	if e != nil {
		return out, e
	}
	route, e := routerHashArgs(relayMethods[name].Inputs[3:4], args[3])
	if e != nil {
		return out, e
	}
	if current != i.CoreRequestHash || route != i.RouteHash {
		return out, fmt.Errorf("Core request or route hash mismatch")
	}
	args[0] = []byte{}
	i.CoreRequestHash, e = routerHashArgs(abi.Arguments{routerKind("bytes32"), relayMethods["transact"].Inputs[1]}, crypto.Keccak256Hash(nil), args[1])
	if e != nil {
		return out, e
	}
	args[2] = *i
	if !native {
		p := abi.ConvertType(args[4], new(relayPermit)).(*relayPermit)
		p.V = 0
		p.R = [32]byte{}
		p.S = [32]byte{}
		args[4] = *p
		args[5] = []byte{}
	}
	packed, e := relayMethods[name].Inputs.Pack(args...)
	if e != nil {
		return out, e
	}
	out.Data = append(bytes.Clone(relayMethods[name].ID), packed...)
	if native {
		outer[0] = []byte(out.Data)
		outer[5] = []byte{}
		packed, e = relayMethods["executeNativePayment"].Inputs.Pack(outer...)
		if e != nil {
			return out, e
		}
		out.Data = append(bytes.Clone(relayMethods["executeNativePayment"].ID), packed...)
	}
	return out, nil
}

// RouterIntentHash binds canonical business intent across proving and signing.
func RouterIntentHash(chainID *big.Int, core common.Address, env RouterEnvelope) (common.Hash, error) {
	normalized, e := NormalizeRouterEnvelope(env)
	if e != nil {
		return common.Hash{}, e
	}
	h, e := EnvelopeHash(chainID, core, normalized)
	if e != nil {
		return common.Hash{}, e
	}
	return routerHashArgs(abi.Arguments{routerKind("bytes32"), routerKind("bytes32")}, crypto.Keccak256Hash([]byte("InvisiPaymentRouterIntentV1")), h)
}

// ValidateRouterSubmission prevents draft proof/signatures reaching submission.
// Cryptographic wallet authorization and live deployment checks remain mandatory server-side.
func ValidateRouterSubmission(data []byte, env RouterEnvelope) error {
	if _, e := relayDeadline(data, env); e != nil {
		return e
	}
	if _, e := NormalizeRouterEnvelope(env); e != nil {
		return e
	}
	inner := []byte(env.Data)
	name := "payTokenFor"
	if len(inner) >= 4 && bytes.Equal(inner[:4], relayMethods["executeNativePayment"].ID) {
		a, e := unpackRelay("executeNativePayment", inner)
		if e != nil {
			return e
		}
		if len(a[5].([]byte)) == 0 {
			return fmt.Errorf("account signature required")
		}
		inner = a[0].([]byte)
		name = "payNative"
	}
	a, e := unpackRelay(name, inner)
	if e != nil {
		return e
	}
	if len(a[0].([]byte)) == 0 {
		return fmt.Errorf("Core proof required")
	}
	if name == "payTokenFor" && len(a[5].([]byte)) == 0 {
		return fmt.Errorf("Router signature required")
	}
	if name == "payTokenFor" {
		p := abi.ConvertType(a[4], new(relayPermit)).(*relayPermit)
		if p.Deadline.Sign() > 0 && (p.V < 27 || p.V > 28 || p.R == ([32]byte{}) || p.S == ([32]byte{})) {
			return fmt.Errorf("permit signature required")
		}
	}
	for _, auth := range env.AuthorizationList {
		if _, e := auth.Authority(); e != nil {
			return fmt.Errorf("invalid delegation signature: %w", e)
		}
	}
	return nil
}
