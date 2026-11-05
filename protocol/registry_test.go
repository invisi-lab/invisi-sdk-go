package protocol

import (
	"bytes"
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func registrationCase(t *testing.T) (RegistryContext, AddressRegistration) {
	t.Helper()
	wallet, err := core.DeriveWallet(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	var clue [32]byte
	copy(clue[:], wallet.ClueKey)
	return RegistryContext{ChainID: big.NewInt(31337), RegistryAddress: common.HexToAddress("0x1234"), Now: big.NewInt(100)}, AddressRegistration{OwnerAddress: wallet.OwnerAddress.Bytes(), Keys: contracts.IAddressRegistryReceivingKeys{ViewingPublicKey: wallet.ViewingPublicKey, ClueKey: clue}, Deadline: 200}
}

func TestRegistrationUsesFlatRegistryTypedData(t *testing.T) {
	context, registration := registrationCase(t)
	data, err := RegistrationTypedData(context, registration, RegistrationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if data.PrimaryType != "RegisterReceivingKeys" || data.Domain.Name != "InvisiAddressRegistry" || data.Domain.Version != "1" || data.Domain.VerifyingContract != context.RegistryAddress.Hex() {
		t.Fatalf("unexpected registry domain: %+v", data)
	}
	fields := data.Types[data.PrimaryType]
	expected := []string{"walletIdentity:bytes32", "viewingPublicKey:bytes", "clueKey:bytes32", "nonce:uint64", "deadline:uint64"}
	if len(fields) != len(expected) {
		t.Fatal(fields)
	}
	for i, field := range fields {
		if field.Name+":"+field.Type != expected[i] {
			t.Fatal(fields)
		}
	}
	register, err := ComputeRegistrationDigest(context, registration, RegistrationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	update, err := ComputeRegistrationDigest(context, registration, RegistrationOptions{Update: true})
	if err != nil || register == update {
		t.Fatal("register and update digests must differ", err)
	}
	context.ChainID = big.NewInt(97)
	other, err := ComputeRegistrationDigest(context, registration, RegistrationOptions{})
	if err != nil || other == register {
		t.Fatal("chain is not bound", err)
	}
	context.RegistryAddress = common.HexToAddress("0x5678")
	again, err := ComputeRegistrationDigest(context, registration, RegistrationOptions{})
	if err != nil || again == other {
		t.Fatal("registry is not bound", err)
	}
}

func TestRegistrationCalldataPreservesDirectAndContractSignatures(t *testing.T) {
	context, registration := registrationCase(t)
	parsed, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	for _, signature := range [][]byte{nil, {1, 2, 3}, bytes.Repeat([]byte{4}, 65)} {
		call, err := BuildRegistrationCalldata(context, registration, signature, RegistrationOptions{})
		if err != nil {
			t.Fatal(err)
		}
		method, err := parsed.MethodById(call.Data[:4])
		if err != nil || method.RawName != "registerKeys" {
			t.Fatal(err)
		}
		args, err := method.Inputs.Unpack(call.Data[4:])
		if err != nil || !bytes.Equal(args[1].([]byte), signature) || call.To != context.RegistryAddress || call.Value.Sign() != 0 {
			t.Fatal("signature calldata changed", err)
		}
	}
}

func TestRegistrationValidatesMaterialAndFreshnessOnEveryPath(t *testing.T) {
	for _, name := range []string{"zero owner", "wide owner", "invalid viewing", "invalid clue", "expired", "exhausted nonce", "initial nonce", "nonce mismatch"} {
		t.Run(name, func(t *testing.T) {
			context, registration := registrationCase(t)
			options := RegistrationOptions{}
			switch name {
			case "zero owner":
				registration.OwnerAddress = [32]byte{}
			case "wide owner":
				registration.OwnerAddress[0] = 1
			case "invalid viewing":
				registration.Keys.ViewingPublicKey = make([]byte, 33)
			case "invalid clue":
				registration.Keys.ClueKey = [32]byte{}
			case "expired":
				registration.Deadline = 99
			case "exhausted nonce":
				registration.Nonce = ^uint64(0)
				options.Update = true
			case "initial nonce":
				registration.Nonce = 1
			case "nonce mismatch":
				expected := uint64(1)
				options.ExpectedNonce = &expected
			}
			if _, err := RegistrationTypedData(context, registration, options); err == nil {
				t.Fatal("invalid typed data accepted")
			}
			if _, err := ComputeRegistrationDigest(context, registration, options); err == nil {
				t.Fatal("invalid digest accepted")
			}
			if _, err := BuildRegistrationCalldata(context, registration, nil, options); err == nil {
				t.Fatal("invalid direct call accepted")
			}
		})
	}
}

func TestRegistrationMatchesIndependentProtocolVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/address-registration-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Rows []struct {
			Context struct {
				ChainID         string
				RegistryAddress common.Address
			}
			Registration map[string]any
			Update       bool
			TypedData    struct {
				Domain      struct{ Name, Version, ChainID, VerifyingContract string }
				Types       apitypes.Types
				PrimaryType string
				Message     apitypes.TypedDataMessage
			}
			Digest    common.Hash
			Signature string
			Signer    common.Address
			CallData  string
		}
	}
	if err = json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Rows) != 4 {
		t.Fatal("expected four cross-language registry vectors")
	}
	for i, row := range vectors.Rows {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			chain, ok := new(big.Int).SetString(row.Context.ChainID, 10)
			if !ok {
				t.Fatal("invalid chain id")
			}
			context := RegistryContext{ChainID: chain, RegistryAddress: row.Context.RegistryAddress}
			var registration AddressRegistration
			decodeReference(t, row.Registration, &registration)
			options := RegistrationOptions{Update: row.Update}
			data, err := RegistrationTypedData(context, registration, options)
			if err != nil {
				t.Fatal(err)
			}
			if data.Domain.Name != row.TypedData.Domain.Name || data.Domain.Version != row.TypedData.Domain.Version || (*big.Int)(data.Domain.ChainId).String() != row.TypedData.Domain.ChainID || data.Domain.VerifyingContract != row.TypedData.Domain.VerifyingContract || data.PrimaryType != row.TypedData.PrimaryType || !reflect.DeepEqual(data.Types[data.PrimaryType], row.TypedData.Types[data.PrimaryType]) || !reflect.DeepEqual(data.Message, row.TypedData.Message) {
				t.Fatal("typed data differs from protocol vector")
			}
			digest, err := ComputeRegistrationDigest(context, registration, options)
			if err != nil || digest != row.Digest {
				t.Fatal("digest differs from protocol vector", err)
			}
			signature := common.FromHex(row.Signature)
			recovery := append([]byte(nil), signature...)
			recovery[64] -= 27
			public, err := crypto.SigToPub(digest[:], recovery)
			if err != nil || crypto.PubkeyToAddress(*public) != row.Signer {
				t.Fatal("protocol signature mismatch", err)
			}
			call, err := BuildRegistrationCalldata(context, registration, signature, options)
			if err != nil || "0x"+common.Bytes2Hex(call.Data) != row.CallData {
				t.Fatal("calldata differs from protocol vector", err)
			}
		})
	}
}
