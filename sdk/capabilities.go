package sdk

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"strings"
)

// WalletAuthorizationProvider can delegate private statement signing without
// disclosing an account private key to the SDK.
type WalletAuthorizationProvider interface {
	WalletIdentity() common.Address
	Authorize(context.Context, core.WalletAuthorizationRequest) (core.WalletAuthorizationWitness, error)
}

func (a *WalletAuthorization) WalletIdentity() common.Address {
	if a == nil || a.material == nil {
		return common.Address{}
	}
	return common.BigToAddress(core.ElementToBigInt(a.material.WalletIdentity))
}

// FundingSigner grants public ERC20 funding independently of private note spending.
type FundingSigner interface {
	Address() common.Address
	SignTypedData(context.Context, apitypes.TypedData) ([]byte, error)
}
type FundingSignerFunc struct {
	Payer common.Address
	Sign  func(context.Context, apitypes.TypedData) ([]byte, error)
}

func (s FundingSignerFunc) Address() common.Address { return s.Payer }
func (s FundingSignerFunc) SignTypedData(ctx context.Context, data apitypes.TypedData) ([]byte, error) {
	if s.Sign == nil {
		return nil, errors.New("funding signing callback required")
	}
	return s.Sign(ctx, data)
}

type FundingAuthorizationVerifier interface {
	VerifyTypedData(context.Context, common.Address, apitypes.TypedData, []byte) (bool, error)
}

// GethFundingVerifier handles both EOA signatures and ERC-1271 contract wallets.
type GethFundingVerifier struct{ Provider Provider }

func (v GethFundingVerifier) VerifyTypedData(ctx context.Context, address common.Address, data apitypes.TypedData, signature []byte) (bool, error) {
	if v.Provider == nil {
		return false, errors.New("funding verification provider required")
	}
	chain, err := v.Provider.ChainID(ctx)
	if err != nil {
		return false, err
	}
	if data.Domain.ChainId == nil || chain.Cmp((*big.Int)(data.Domain.ChainId)) != 0 {
		return false, errors.New("funding verification chain mismatch")
	}
	code, err := v.Provider.CodeAt(ctx, address, nil)
	if err != nil {
		return false, err
	}
	digest, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return false, err
	}
	if len(code) == 0 {
		if len(signature) != 65 {
			return false, nil
		}
		sig := append([]byte(nil), signature...)
		if sig[64] >= 27 {
			sig[64] -= 27
		}
		if sig[64] > 1 || !crypto.ValidateSignatureValues(sig[64], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:64]), true) {
			return false, nil
		}
		public, err := crypto.SigToPub(digest, sig)
		return err == nil && crypto.PubkeyToAddress(*public) == address, nil
	}
	contractABI, _ := abi.JSON(strings.NewReader(`[{"type":"function","name":"isValidSignature","stateMutability":"view","inputs":[{"type":"bytes32"},{"type":"bytes"}],"outputs":[{"type":"bytes4"}]}]`))
	call, err := contractABI.Pack("isValidSignature", common.BytesToHash(digest), signature)
	if err != nil {
		return false, err
	}
	result, err := v.Provider.CallContract(ctx, ethereum.CallMsg{To: &address, Data: call}, nil)
	if err != nil {
		// A contract rejection is an invalid signature; transport errors remain errors.
		if strings.Contains(err.Error(), "execution reverted") {
			return false, nil
		}
		return false, err
	}
	values, err := contractABI.Unpack("isValidSignature", result)
	if err != nil || len(values) != 1 {
		return false, nil
	}
	magic, ok := values[0].([4]byte)
	return ok && magic == ([4]byte{0x16, 0x26, 0xba, 0x7e}), nil
}

func verifyAuthorizationWitness(identity common.Address, request core.WalletAuthorizationRequest, witness core.WalletAuthorizationWitness) error {
	public := append([]byte{4}, witness.PublicKeyX[:]...)
	public = append(public, witness.PublicKeyY[:]...)
	valid, err := core.VerifyWalletAuthorization(core.BigIntToElement(identity.Big()), public, witness.Signature[:], request)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("wallet authorization differs from frozen statement")
	}
	return nil
}
