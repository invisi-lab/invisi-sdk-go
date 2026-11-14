package primitives

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func TestExternalWalletSignatureAndJSONRequest(t *testing.T) {
	private := bytes.Repeat([]byte{1}, 32)
	wallet, err := DeriveWallet(private)
	if err != nil {
		t.Fatal(err)
	}
	request := WalletAuthorizationRequest{Context: ProtocolContext{ElementFromUint64(97), ElementFromUint64(2)}, Operation: WalletTransact, Capacity: 1, Words: []fr.Element{ElementFromUint64(97), ElementFromUint64(2), ElementFromUint64(9)}}
	typed, err := WalletAuthorizationTypedData(request)
	if err != nil {
		t.Fatal(err)
	}
	portable, err := NewSigningRequest(typed)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(portable)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"chainId":"97"`)) || bytes.Contains(encoded, []byte(`"EIP712Domain"`)) {
		t.Fatal(string(encoded))
	}
	var restored SigningRequest
	if err = json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	roundTrip, err := restored.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	a, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := apitypes.TypedDataAndHash(roundTrip)
	if err != nil || !bytes.Equal(a, b) {
		t.Fatal("request digest changed", err)
	}
	digest, err := ComputeWalletAuthorizationDigest(request)
	if err != nil {
		t.Fatal(err)
	}
	key, _ := ethcrypto.ToECDSA(private)
	signature, _ := ethcrypto.Sign(digest[:], key)
	external, err := BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, signature)
	if err != nil {
		t.Fatal(err)
	}
	local, err := BuildWalletAuthorizationWitness(wallet, request)
	if err != nil || external != local {
		t.Fatal("external witness differs", err)
	}
	signature[64] += 27
	if _, err = BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, signature); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{nil, bytes.Repeat([]byte{0}, 64), bytes.Repeat([]byte{0}, 65)} {
		if _, err = BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, bad); err == nil {
			t.Fatal("invalid signature accepted")
		}
	}
	if _, err = BuildWalletAuthorizationWitnessFromSignature(ElementFromUint64(3), request, signature); err == nil {
		t.Fatal("wrong owner accepted")
	}
	request.Words[2] = ElementFromUint64(10)
	if _, err = BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, signature); err == nil {
		t.Fatal("wrong statement accepted")
	}
}

func TestSigningRequestPreservesLargeIntegersAndCopies(t *testing.T) {
	large := new(big.Int).Lsh(big.NewInt(1), 200)
	data := apitypes.TypedData{Types: apitypes.Types{
		"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "chainId", Type: "uint256"}},
		"Grant":        {{Name: "amount", Type: "uint256"}, {Name: "values", Type: "uint256[]"}},
	}, Domain: apitypes.TypedDataDomain{Name: "Test", ChainId: (*ethmath.HexOrDecimal256)(big.NewInt(97))}, PrimaryType: "Grant", Message: apitypes.TypedDataMessage{"amount": large, "values": []*big.Int{large}}}
	request, err := NewSigningRequest(data)
	if err != nil {
		t.Fatal(err)
	}
	large.SetInt64(1)
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	expected := new(big.Int).Lsh(big.NewInt(1), 200).String()
	if !bytes.Contains(encoded, []byte(`"amount":"`+expected+`"`)) {
		t.Fatal(string(encoded))
	}
	var restored SigningRequest
	if err = json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	typed, err := restored.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	if typed.Message["amount"] != expected {
		t.Fatal("integer changed")
	}
}

func TestExternalWalletSignatureRejectsHighS(t *testing.T) {
	private := bytes.Repeat([]byte{1}, 32)
	wallet, _ := DeriveWallet(private)
	request := WalletAuthorizationRequest{Context: ProtocolContext{ElementFromUint64(97), ElementFromUint64(2)}, Operation: WalletTransact, Capacity: 1, Words: []fr.Element{ElementFromUint64(97), ElementFromUint64(2)}}
	digest, _ := ComputeWalletAuthorizationDigest(request)
	key, _ := ethcrypto.ToECDSA(private)
	signature, _ := ethcrypto.Sign(digest[:], key)
	high := new(big.Int).Sub(ethcrypto.S256().Params().N, new(big.Int).SetBytes(signature[32:64]))
	high.FillBytes(signature[32:64])
	signature[64] ^= 1
	if _, err := BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, signature); err == nil {
		t.Fatal("high-s accepted")
	}
}

func TestExternalWalletSignatureAcceptsCompactEncoding(t *testing.T) {
	private := bytes.Repeat([]byte{1}, 32)
	wallet, _ := DeriveWallet(private)
	request := WalletAuthorizationRequest{Context: ProtocolContext{ElementFromUint64(97), ElementFromUint64(2)}, Operation: WalletTransact, Capacity: 1, Words: []fr.Element{ElementFromUint64(97), ElementFromUint64(2)}}
	digest, _ := ComputeWalletAuthorizationDigest(request)
	key, _ := ethcrypto.ToECDSA(private)
	signature, _ := ethcrypto.Sign(digest[:], key)
	compact := bytes.Clone(signature[:64])
	compact[32] |= signature[64] << 7
	witness, err := BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, compact)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := BuildWalletAuthorizationWitnessFromSignature(wallet.OwnerAddress, request, signature)
	if err != nil || witness != expected {
		t.Fatal("compact witness differs", err)
	}
}

func TestWalletStatementBulkBoundaryMatchesABI(t *testing.T) {
	uintType, err := abi.NewType("uint256", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	wordsType, err := abi.NewType("bytes32[]", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	arguments := abi.Arguments{{Type: uintType}, {Type: wordsType}}
	// Guard the signed transcript on both sides of the contract bulk-copy threshold; 128 is not a validity limit.
	for _, length := range []int{2, 127, 128, 129, 1024} {
		t.Run(big.NewInt(int64(length)).String(), func(t *testing.T) {
			words := make([]fr.Element, length)
			encodedWords := make([][32]byte, length)
			for i := range words {
				words[i] = ElementFromUint64(uint64(i + 1))
			}
			words[0], words[1] = ElementFromUint64(97), ElementFromUint64(2)
			for i := range words {
				encodedWords[i] = words[i].Bytes()
			}
			for _, operation := range []WalletOperation{WalletTransact, WalletCreateCredit, WalletCreditDraw, WalletUpdateCredit} {
				request := WalletAuthorizationRequest{Context: ProtocolContext{words[0], words[1]}, Operation: operation, Capacity: 1, Words: words}
				typed, err := WalletAuthorizationTypedData(request)
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := arguments.Pack(big.NewInt(1), encodedWords)
				if err != nil {
					t.Fatal(err)
				}
				if typed.Message["statementHash"] != ethcrypto.Keccak256Hash(encoded).Hex() {
					t.Fatalf("%s statement differs from ABI", operation)
				}
			}
		})
	}
}
