package core

import (
	"crypto/ecdsa"
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// InvisiWallet is a private key group. Never serialize it into payment requests,
// registry calldata, links, logs, or other public transport.
type InvisiWallet struct {
	SpendingPrivateKey, SpendingPublicKey []byte
	ViewingPrivateKey, ViewingPublicKey   []byte
	DetectionSeed, DetectionKey, ClueKey  []byte
	WalletIdentity                        fr.Element
}
type WalletOperation string

const (
	WalletTransact        WalletOperation = "transact"
	WalletCreateCredit    WalletOperation = "create-credit"
	WalletCreditDraw      WalletOperation = "credit-draw"
	WalletCancelCredit    WalletOperation = "cancel-credit"
	WalletTopUpCredit     WalletOperation = "top-up-credit"
	WalletRegisterAddress WalletOperation = "register-address"
	WalletUpdateAddress   WalletOperation = "update-address"
)

// WalletAuthorizationRequest binds a private spending-key signature to the
// current public statement. Its words precede the final two digest limbs.
// This authorization is separate from EVM funding/permit signatures.
type WalletAuthorizationRequest struct {
	Context   ProtocolContext
	Operation WalletOperation
	Capacity  int
	Words     []fr.Element
}

func walletScalar(key []byte, salt string, littleEndian bool) ([]byte, error) {
	for counter := uint64(0); counter <= uint64(^uint32(0)); counter++ {
		var info [4]byte
		binary.BigEndian.PutUint32(info[:], uint32(counter))
		candidate, e := hkdf.Key(sha256.New, key, []byte(salt), string(info[:]), 32)
		if e != nil {
			return nil, e
		}
		var n *big.Int
		if littleEndian {
			candidate[31] &= 7
			n = decafReadLE(candidate)
		} else {
			n = new(big.Int).SetBytes(candidate)
		}
		modulus := ethcrypto.S256().Params().N
		if littleEndian {
			modulus = decafScalarOrder
		}
		if n.Sign() > 0 && n.Cmp(modulus) < 0 {
			return candidate, nil
		}
	}
	return nil, fmt.Errorf("wallet scalar derivation exhausted")
}

// DeriveWallet restores the versioned Invisi key group from a 32-byte EVM account key.
// Mnemonics and account derivation remain the application's responsibility.
func DeriveWallet(privateKey []byte) (*InvisiWallet, error) {
	spend := append([]byte(nil), privateKey...)
	spendKey, err := ethcrypto.ToECDSA(spend)
	if err != nil {
		return nil, fmt.Errorf("invalid EVM account private key: %w", err)
	}
	root, err := hkdf.Key(sha256.New, spend, []byte("Invisi.WalletRoot.v1"), "", 32)
	if err != nil {
		return nil, err
	}
	defer clear(root)
	view, err := walletScalar(root, "Invisi.WalletViewingKey.v1", false)
	if err != nil {
		return nil, err
	}
	seed, err := hkdf.Key(sha256.New, root, []byte("Invisi.WalletDetectionSeed.v1"), "", 32)
	if err != nil {
		return nil, err
	}
	detection, err := walletScalar(seed, "Invisi.WalletDetectionKey.v1", true)
	if err != nil {
		return nil, err
	}
	clue, err := DeriveClueKey(detection)
	if err != nil {
		return nil, err
	}
	public := ethcrypto.FromECDSAPub(&spendKey.PublicKey)
	identity, err := ComputeWalletIdentity(public)
	if err != nil {
		return nil, err
	}
	viewKey, err := ethcrypto.ToECDSA(view)
	if err != nil {
		return nil, err
	}
	return &InvisiWallet{SpendingPrivateKey: spend, SpendingPublicKey: public, ViewingPrivateKey: view, ViewingPublicKey: ethcrypto.CompressPubkey(&viewKey.PublicKey), DetectionSeed: seed, DetectionKey: detection, ClueKey: clue, WalletIdentity: identity}, nil
}
func walletPublicKey(input []byte) (*ecdsa.PublicKey, error) {
	switch len(input) {
	case 33:
		if input[0] != 2 && input[0] != 3 {
			return nil, fmt.Errorf("invalid compressed spending public key")
		}
		return ethcrypto.DecompressPubkey(input)
	case 65:
		if input[0] != 4 {
			return nil, fmt.Errorf("invalid spending public key prefix")
		}
		return ethcrypto.UnmarshalPubkey(input)
	default:
		return nil, fmt.Errorf("spending public key must have 33 or 65 bytes")
	}
}

// ComputeWalletIdentity is the EVM owner of the spending public key.
func ComputeWalletIdentity(publicKey []byte) (fr.Element, error) {
	point, err := walletPublicKey(publicKey)
	if err != nil {
		return fr.Element{}, err
	}
	return BigIntToElement(ethcrypto.PubkeyToAddress(*point).Big()), nil
}
func walletOperationName(operation WalletOperation, capacity int) (string, error) {
	name := ""
	allowed := false
	switch operation {
	case WalletTransact:
		name = "Transact"
		allowed = capacity == 14 || capacity == 56 || capacity == 112 || capacity == 168
	case WalletCreateCredit:
		name = "CreateCredit"
		allowed = capacity == 10
	case WalletCreditDraw:
		name = "CreditDraw"
		allowed = capacity == 1 || capacity == 12 || capacity == 50 || capacity == 120
	case WalletCancelCredit:
		name = "CancelCredit"
		allowed = capacity == 1
	case WalletTopUpCredit:
		name = "TopUpCredit"
		allowed = capacity == 10
	case WalletRegisterAddress:
		name = "RegisterAddress"
		allowed = capacity == 0
	case WalletUpdateAddress:
		name = "UpdateAddress"
		allowed = capacity == 0
	default:
		return "", fmt.Errorf("unsupported wallet operation")
	}
	if !allowed {
		return "", fmt.Errorf("invalid wallet operation capacity")
	}
	return name, nil
}

// WalletAuthorizationTypedData constructs the exact EIP712 public transcript.
func WalletAuthorizationTypedData(request WalletAuthorizationRequest) (apitypes.TypedData, error) {
	name, e := walletOperationName(request.Operation, request.Capacity)
	if e != nil {
		return apitypes.TypedData{}, e
	}
	if len(request.Words) < 2 {
		return apitypes.TypedData{}, fmt.Errorf("public words required")
	}
	chain, address := ElementToBigInt(request.Context.ChainID), ElementToBigInt(request.Context.ProtocolAddress)
	if chain.BitLen() > 64 {
		return apitypes.TypedData{}, fmt.Errorf("chain id exceeds uint64")
	}
	if address.Sign() == 0 || address.BitLen() > 160 {
		return apitypes.TypedData{}, fmt.Errorf("invalid protocol address")
	}
	if request.Words[0] != request.Context.ChainID {
		return apitypes.TypedData{}, fmt.Errorf("statement chain mismatch")
	}
	if request.Words[1] != request.Context.ProtocolAddress {
		return apitypes.TypedData{}, fmt.Errorf("statement instance mismatch")
	}
	statement := make([]byte, 96+32*len(request.Words))
	big.NewInt(int64(request.Capacity)).FillBytes(statement[:32])
	big.NewInt(64).FillBytes(statement[32:64])
	new(big.Int).SetUint64(uint64(len(request.Words))).FillBytes(statement[64:96])
	for i, value := range request.Words {
		b := value.Bytes()
		copy(statement[96+32*i:], b[:])
	}
	chainID := ethmath.HexOrDecimal256(*chain)
	return apitypes.TypedData{Types: apitypes.Types{
		"EIP712Domain":        {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}},
		"WalletAuthorization": {{Name: "operation", Type: "bytes32"}, {Name: "statementHash", Type: "bytes32"}},
	}, PrimaryType: "WalletAuthorization", Domain: apitypes.TypedDataDomain{Name: "InvisibleProtocol", Version: "1", ChainId: &chainID, VerifyingContract: common.BigToAddress(address).Hex()}, Message: apitypes.TypedDataMessage{
		"operation": ethcrypto.Keccak256Hash([]byte("Invisi." + name + ".v1")).Hex(), "statementHash": ethcrypto.Keccak256Hash(statement).Hex(),
	}}, nil
}
func ComputeWalletAuthorizationDigest(request WalletAuthorizationRequest) (common.Hash, error) {
	data, e := WalletAuthorizationTypedData(request)
	if e != nil {
		return common.Hash{}, e
	}
	hash, _, e := apitypes.TypedDataAndHash(data)
	if e != nil {
		return common.Hash{}, e
	}
	return common.BytesToHash(hash), nil
}

// SignWalletAuthorization returns raw low-s r||s (64 bytes), never EVM r||s||v.
func SignWalletAuthorization(privateKey []byte, request WalletAuthorizationRequest) ([]byte, error) {
	digest, e := ComputeWalletAuthorizationDigest(request)
	if e != nil {
		return nil, e
	}
	key, e := ethcrypto.ToECDSA(privateKey)
	if e != nil {
		return nil, e
	}
	signed, e := ethcrypto.Sign(digest[:], key)
	if e != nil {
		return nil, e
	}
	return append([]byte(nil), signed[:64]...), nil
}
func VerifyWalletAuthorization(identity fr.Element, publicKey, signature []byte, request WalletAuthorizationRequest) (bool, error) {
	if len(signature) != 64 {
		return false, fmt.Errorf("wallet signature must contain exactly 64 bytes")
	}
	derived, e := ComputeWalletIdentity(publicKey)
	if e != nil {
		return false, e
	}
	if derived != identity {
		return false, nil
	}
	digest, e := ComputeWalletAuthorizationDigest(request)
	if e != nil {
		return false, e
	}
	return ethcrypto.VerifySignature(publicKey, digest[:], signature), nil
}
