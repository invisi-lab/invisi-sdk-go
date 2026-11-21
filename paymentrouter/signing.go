package paymentrouter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
)

func typedPayment(chain *big.Int, verifying common.Address, name, primary string, salt string, fields []apitypes.Type, msg apitypes.TypedDataMessage) (apitypes.TypedData, error) {
	if e := uintCheck(chain, 256, true, "chain ID"); e != nil {
		return apitypes.TypedData{}, e
	}
	if !nonzero(verifying) {
		return apitypes.TypedData{}, fmt.Errorf("missing verifying contract")
	}
	domain := []apitypes.Type{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}
	if salt != "" {
		domain = append(domain, apitypes.Type{Name: "salt", Type: "bytes32"})
	}
	return apitypes.TypedData{Types: apitypes.Types{"EIP712Domain": domain, primary: fields}, PrimaryType: primary, Domain: apitypes.TypedDataDomain{Name: name, Version: "1", ChainId: (*ethmath.HexOrDecimal256)(copyInt(chain)), VerifyingContract: verifying.Hex(), Salt: salt}, Message: msg}, nil
}
func IntentTypedData(chain *big.Int, router common.Address, i Intent) (apitypes.TypedData, error) {
	for _, n := range []*big.Int{i.AmountInMaximum, i.Nonce, i.Deadline} {
		if e := uintCheck(n, 256, false, "intent integer"); e != nil {
			return apitypes.TypedData{}, e
		}
	}
	fields := []apitypes.Type{{Name: "payer", Type: "address"}, {Name: "tokenIn", Type: "address"}, {Name: "amountInMaximum", Type: "uint256"}, {Name: "refundRecipient", Type: "address"}, {Name: "relayer", Type: "address"}, {Name: "nonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "coreRequestHash", Type: "bytes32"}, {Name: "routeHash", Type: "bytes32"}, {Name: "mode", Type: "uint8"}}
	return typedPayment(chain, router, "PaymentRouter", "Intent", "", fields, apitypes.TypedDataMessage{"payer": i.Payer.Hex(), "tokenIn": i.TokenIn.Hex(), "amountInMaximum": i.AmountInMaximum.String(), "refundRecipient": i.RefundRecipient.Hex(), "relayer": i.Relayer.Hex(), "nonce": i.Nonce.String(), "deadline": i.Deadline.String(), "coreRequestHash": i.CoreRequestHash.Hex(), "routeHash": i.RouteHash.Hex(), "mode": fmt.Sprint(i.Mode)})
}
func NativePaymentTypedData(chain *big.Int, player, implementation, router common.Address, routerCodeHash common.Hash, p AccountPayment) (apitypes.TypedData, error) {
	for _, n := range []*big.Int{p.Value, p.Nonce, p.Deadline} {
		if e := uintCheck(n, 256, false, "account integer"); e != nil {
			return apitypes.TypedData{}, e
		}
	}
	if !nonzero(implementation) || !nonzero(router) {
		return apitypes.TypedData{}, fmt.Errorf("missing account implementation or router")
	}
	fields := []apitypes.Type{{Name: "routerCalldataHash", Type: "bytes32"}, {Name: "value", Type: "uint256"}, {Name: "nonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}, {Name: "relayer", Type: "address"}, {Name: "router", Type: "address"}, {Name: "routerCodeHash", Type: "bytes32"}}
	return typedPayment(chain, player, "PaymentRouterAccount", "NativePayment", common.BytesToHash(implementation.Bytes()).Hex(), fields, apitypes.TypedDataMessage{"routerCalldataHash": crypto.Keccak256Hash(p.RouterCalldata).Hex(), "value": p.Value.String(), "nonce": p.Nonce.String(), "deadline": p.Deadline.String(), "relayer": p.Relayer.Hex(), "router": router.Hex(), "routerCodeHash": routerCodeHash.Hex()})
}
func typedDigest(t apitypes.TypedData, e error) (common.Hash, error) {
	if e != nil {
		return common.Hash{}, e
	}
	h, _, e := apitypes.TypedDataAndHash(t)
	return common.BytesToHash(h), e
}
func HashIntent(chain *big.Int, router common.Address, i Intent) (common.Hash, error) {
	return typedDigest(IntentTypedData(chain, router, i))
}
func HashNativePayment(chain *big.Int, player, implementation, router common.Address, routerCodeHash common.Hash, p AccountPayment) (common.Hash, error) {
	return typedDigest(NativePaymentTypedData(chain, player, implementation, router, routerCodeHash, p))
}
func PermitTypedData(chain *big.Int, token common.Address, name, version string, owner, router common.Address, amount, nonce, deadline *big.Int) (apitypes.TypedData, error) {
	for _, n := range []*big.Int{amount, nonce, deadline} {
		if e := uintCheck(n, 256, false, "permit integer"); e != nil {
			return apitypes.TypedData{}, e
		}
	}
	fields := []apitypes.Type{{Name: "owner", Type: "address"}, {Name: "spender", Type: "address"}, {Name: "value", Type: "uint256"}, {Name: "nonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}}
	t, e := typedPayment(chain, token, name, "Permit", "", fields, apitypes.TypedDataMessage{"owner": owner.Hex(), "spender": router.Hex(), "value": amount.String(), "nonce": nonce.String(), "deadline": deadline.String()})
	t.Domain.Version = version
	return t, e
}
func PermitFromSignature(deadline *big.Int, signature []byte) (Permit, error) {
	if e := uintCheck(deadline, 256, true, "permit deadline"); e != nil {
		return Permit{}, e
	}
	if len(signature) != 65 {
		return Permit{}, fmt.Errorf("expected 65-byte permit signature")
	}
	v := signature[64]
	if v < 2 {
		v += 27
	}
	if v != 27 && v != 28 {
		return Permit{}, fmt.Errorf("invalid signature recovery ID %s", hexutil.EncodeUint64(uint64(v)))
	}
	return Permit{copyInt(deadline), v, [32]byte(signature[:32]), [32]byte(signature[32:64])}, nil
}
