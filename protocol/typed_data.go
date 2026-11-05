package protocol

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

var domainFields = []apitypes.Type{{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}
var fundingFields = []apitypes.Type{{Name: "requestHash", Type: "bytes32"}, {Name: "payer", Type: "address"}, {Name: "amount", Type: "uint256"}, {Name: "sigNonce", Type: "uint256"}, {Name: "deadline", Type: "uint256"}}

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

// WalletAuthorizationRequest uses only canonical public words; private wallet keys remain local.
func WalletAuthorizationRequest(c TransactionContext, operation core.WalletOperation, capacity int, baseWords [][32]byte) (core.WalletAuthorizationRequest, error) {
	if err := ValidateContext(c); err != nil {
		return core.WalletAuthorizationRequest{}, err
	}
	request := core.WalletAuthorizationRequest{Context: core.ProtocolContext{ChainID: core.BigIntToElement(c.ChainID), ProtocolAddress: core.BigIntToElement(c.ProtocolAddress.Big())}, Operation: operation, Capacity: capacity}
	for _, value := range baseWords {
		if err := Field(value); err != nil {
			return request, err
		}
		request.Words = append(request.Words, core.BigIntToElement(number(value)))
	}
	if _, err := core.WalletAuthorizationTypedData(request); err != nil {
		return request, err
	}
	return request, nil
}
func EncodeWalletAuthorizationTypedData(c TransactionContext, operation core.WalletOperation, capacity int, baseWords [][32]byte) (apitypes.TypedData, error) {
	request, err := WalletAuthorizationRequest(c, operation, capacity, baseWords)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	return core.WalletAuthorizationTypedData(request)
}
func AppendWalletAuthorization(c TransactionContext, operation core.WalletOperation, capacity int, baseWords [][32]byte) ([][32]byte, error) {
	request, err := WalletAuthorizationRequest(c, operation, capacity, baseWords)
	if err != nil {
		return nil, err
	}
	digest, err := core.ComputeWalletAuthorizationDigest(request)
	if err != nil {
		return nil, err
	}
	high, low := splitHash(digest)
	out := append([][32]byte(nil), baseWords...)
	return append(out, Word(high), Word(low)), nil
}
