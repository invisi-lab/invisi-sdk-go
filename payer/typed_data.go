package payer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"math/big"
)

var domainFields = []apitypes.Type{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}
var fundingFields = []apitypes.Type{{Name: "requestHash", Type: "bytes32"}, {Name: "payer", Type: "address"}, {Name: "amount", Type: "uint256"}, {Name: "sigNonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}}
var authorizationFields = []apitypes.Type{{Name: "requestHash", Type: "bytes32"}, {Name: "owner", Type: "address"}, {Name: "sigNonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}}

func typedData(c TransactionContext, primary string, fields []apitypes.Type, message map[string]any) apitypes.TypedData {
	return apitypes.TypedData{Types: apitypes.Types{"EIP712Domain": append([]apitypes.Type(nil), domainFields...), primary: append([]apitypes.Type(nil), fields...)}, PrimaryType: primary,
		Domain: apitypes.TypedDataDomain{Name: "InvisibleProtocol", Version: "1", ChainId: (*ethmath.HexOrDecimal256)(new(big.Int).Set(c.ChainID)), VerifyingContract: c.ProtocolAddress.Hex()}, Message: message}
}
func TypedDataHash(data apitypes.TypedData) (common.Hash, error) {
	h, _, err := apitypes.TypedDataAndHash(data)
	return common.BytesToHash(h), err
}
func EncodeFundingTypedData(c TransactionContext, primary string, hash common.Hash, f FundingParams) (apitypes.TypedData, error) {
	if err := ValidateContext(c); err != nil {
		return apitypes.TypedData{}, err
	}
	f, err := NormalizeFunding(f)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if f.Amount.Sign() == 0 {
		return apitypes.TypedData{}, fmt.Errorf("funding authorization requires positive deposit")
	}
	return typedData(c, primary, fundingFields, map[string]any{"requestHash": hash.Hex(), "payer": f.Payer.Hex(), "amount": f.Amount.String(), "sigNonce": f.SigNonce.String(), "deadline": f.Deadline.String()}), nil
}
func EncodeOwnerTypedData(c TransactionContext, primary string, hash common.Hash, owner common.Address, nonce, deadline *big.Int) (apitypes.TypedData, error) {
	if err := ValidateContext(c); err != nil {
		return apitypes.TypedData{}, err
	}
	if err := CheckUint(nonce, 256, "authorization nonce"); err != nil {
		return apitypes.TypedData{}, err
	}
	if err := CheckDeadline(c, deadline); err != nil {
		return apitypes.TypedData{}, err
	}
	if owner == (common.Address{}) {
		return apitypes.TypedData{}, fmt.Errorf("owner must be nonzero")
	}
	return typedData(c, primary, authorizationFields, map[string]any{"requestHash": hash.Hex(), "owner": owner.Hex(), "sigNonce": nonce.String(), "deadline": deadline.String()}), nil
}
