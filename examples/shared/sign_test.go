package shared

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
)

func TestExampleSignatureRequiresExpectedOwner(t *testing.T) {
	const env = "INVISI_EXAMPLE_SIGNING_TEST_KEY"
	const secret = "0000000000000000000000000000000000000000000000000000000000000001"
	t.Setenv(env, secret)
	key, err := crypto.HexToECDSA(secret)
	if err != nil {
		t.Fatal(err)
	}
	owner := crypto.PubkeyToAddress(key.PublicKey)
	data := apitypes.TypedData{Types: apitypes.Types{"EIP712Domain": {{Name: "name", Type: "string"}}, "Example": {{Name: "amount", Type: "uint256"}}}, PrimaryType: "Example", Domain: apitypes.TypedDataDomain{Name: "Example"}, Message: apitypes.TypedDataMessage{"amount": "10"}}
	request, err := core.NewSigningRequest(data)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := Sign(request, owner, env)
	if err != nil {
		t.Fatal(err)
	}
	digest, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := crypto.SigToPub(digest, signature)
	if err != nil {
		t.Fatal(err)
	}
	if crypto.PubkeyToAddress(*recovered) != owner {
		t.Fatal("signature did not recover the required owner")
	}
	if _, err = Sign(request, common.Address{1}, env); err == nil {
		t.Fatal("wrong owner key was allowed")
	}
}
