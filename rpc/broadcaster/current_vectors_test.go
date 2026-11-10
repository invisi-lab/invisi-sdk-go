package broadcaster

import (
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
)

func TestCurrentCanonicalSixMethodVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/sdk-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Method, Data, RequestHash, UnsignedVariant, Signature, QuoteDigest string
		Quote                                                              Quote
		Trust                                                              struct {
			ChainID          string
			Protocol, Signer common.Address
		}
	}
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 6 {
		t.Fatal("six current operations required")
	}
	a, err := contracts.InvisiProtocolMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		t.Run(row.Method, func(t *testing.T) {
			data := common.FromHex(row.Data)
			expected := common.HexToHash(row.RequestHash)
			actual, err := RequestHash(data)
			if err != nil || actual != expected {
				t.Fatal("canonical request hash mismatch", err)
			}
			unsigned, err := RequestHash(common.FromHex(row.UnsignedVariant))
			if err != nil || unsigned != expected {
				t.Fatal("proof or signature stripping differs from independent canonical reference", err)
			}
			chain, ok := new(big.Int).SetString(row.Trust.ChainID, 10)
			if !ok {
				t.Fatal("invalid vector chain")
			}
			trust := Trust{ChainID: chain, Protocol: row.Trust.Protocol, Signer: row.Trust.Signer}
			signed := SignedQuote{Quote: row.Quote, Signature: row.Signature}
			if err = VerifyQuote(signed, trust, data, 0); err != nil {
				t.Fatal(err)
			}
			typed, err := QuoteTypedData(row.Quote, trust)
			if err != nil {
				t.Fatal(err)
			}
			digest, _, err := apitypes.TypedDataAndHash(typed)
			if err != nil || common.BytesToHash(digest) != common.HexToHash(row.QuoteDigest) {
				t.Fatal("canonical quote digest mismatch", err)
			}
			method, err := a.MethodById(data[:4])
			if err != nil {
				t.Fatal(err)
			}
			for _, dimension := range []string{"FeeAmount", "FeeRecipient"} {
				args, err := method.Inputs.Unpack(data[4:])
				if err != nil {
					t.Fatal(err)
				}
				if row.Method == "revokeCredit" {
					if dimension == "FeeAmount" {
						args[6] = new(big.Int).Add(args[6].(*big.Int), big.NewInt(1))
					} else {
						args[5] = common.HexToAddress("0x9999")
					}
				} else {
					index := 1
					if row.Method == "shield" {
						index = 0
					}
					value := reflect.New(reflect.TypeOf(args[index])).Elem()
					value.Set(reflect.ValueOf(args[index]))
					target := value
					if row.Method == "transact" {
						target = field(value, "Params")
					}
					if dimension == "FeeAmount" {
						original := field(target, dimension).Interface().(*big.Int)
						field(target, dimension).Set(reflect.ValueOf(new(big.Int).Add(original, big.NewInt(1))))
					} else {
						field(target, dimension).Set(reflect.ValueOf(common.HexToAddress("0x9999")))
					}
					args[index] = value.Interface()
				}
				encoded, err := method.Inputs.Pack(args...)
				if err != nil {
					t.Fatal(err)
				}
				changed := append(append([]byte(nil), method.ID...), encoded...)
				hash, err := RequestHash(changed)
				if err != nil || hash == expected {
					t.Fatalf("%s was not bound to request hash: %v", dimension, err)
				}
				if err = VerifyQuote(signed, trust, changed, 0); err == nil {
					t.Fatalf("accepted changed %s against original quote", dimension)
				}
			}
		})
	}
}
