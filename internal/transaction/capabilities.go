package transaction

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// WalletAuthorizationProvider can delegate private statement signing without
// disclosing an account private key to the SDK.
type WalletAuthorizationProvider interface {
	OwnerAddress() common.Address
	Authorize(context.Context, core.WalletAuthorizationRequest) (core.WalletAuthorizationWitness, error)
}

func (a *WalletAuthorization) OwnerAddress() common.Address {
	if a == nil || a.material == nil {
		return common.Address{}
	}
	return common.BigToAddress(core.ElementToBigInt(a.material.OwnerAddress))
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
	return VerifyTypedData(ctx, v.Provider, address, data, signature, "funding")
}

func VerifyTypedData(ctx context.Context, provider Provider, address common.Address, data apitypes.TypedData, signature []byte, purpose string) (bool, error) {
	if provider == nil {
		return false, errors.New(purpose + " verification provider required")
	}
	chain, err := provider.ChainID(ctx)
	if err != nil {
		return false, err
	}
	if data.Domain.ChainId == nil || chain.Cmp((*big.Int)(data.Domain.ChainId)) != 0 {
		return false, errors.New(purpose + " verification chain mismatch")
	}
	digest, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return false, err
	}
	// SignatureChecker first tries ECDSA even when the owner has code (for
	// example, an EIP-7702 delegated EOA), then tries ERC-1271.
	if recovered, valid := RecoverTypedDataSigner(digest, signature); valid && recovered == address {
		return true, nil
	}
	code, err := provider.CodeAt(ctx, address, nil)
	if err != nil {
		return false, err
	}
	if len(code) == 0 {
		return false, nil
	}
	contractABI, _ := abi.JSON(strings.NewReader(`[{"type":"function","name":"isValidSignature","stateMutability":"view","inputs":[{"type":"bytes32"},{"type":"bytes"}],"outputs":[{"type":"bytes4"}]}]`))
	call, err := contractABI.Pack("isValidSignature", common.BytesToHash(digest), signature)
	if err != nil {
		return false, err
	}
	result, err := provider.CallContract(ctx, ethereum.CallMsg{To: &address, Data: call}, nil)
	if err != nil {
		// A contract rejection is an invalid signature; transport errors remain errors.
		if strings.Contains(err.Error(), "execution reverted") {
			return false, nil
		}
		return false, err
	}
	// OpenZeppelin compares the complete ABI word, including zero padding.
	// Additional return data is permitted; short or dirty words are rejected.
	var magic [32]byte
	copy(magic[:], []byte{0x16, 0x26, 0xba, 0x7e})
	return len(result) >= 32 && bytes.Equal(result[:32], magic[:]), nil
}

// RecoverTypedDataSigner accepts the recoverable geth and EVM signature forms.
func RecoverTypedDataSigner(digest, signature []byte) (common.Address, bool) {
	if len(signature) == 64 {
		expanded := make([]byte, 65)
		copy(expanded, signature)
		expanded[64] = expanded[32] >> 7
		expanded[32] &= 0x7f
		signature = expanded
	}
	if len(signature) != 65 {
		return common.Address{}, false
	}
	sig := append([]byte(nil), signature...)
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	if sig[64] > 1 || !crypto.ValidateSignatureValues(sig[64], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:64]), true) {
		return common.Address{}, false
	}
	public, err := crypto.SigToPub(digest, sig)
	if err != nil {
		return common.Address{}, false
	}
	return crypto.PubkeyToAddress(*public), true
}

func VerifyAuthorizationWitness(identity common.Address, request core.WalletAuthorizationRequest, witness core.WalletAuthorizationWitness) error {
	public := append([]byte{4}, witness.PublicKeyX[:]...)
	public = append(public, witness.PublicKeyY[:]...)
	valid, err := core.VerifyWalletAuthorization(core.BigIntToElement(identity.Big()), public, witness.Signature[:], request)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("wallet authorization differs from Frozen statement")
	}
	return nil
}
