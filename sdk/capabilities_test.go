package sdk

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"math/big"
	"strings"
	"testing"
)

type delegatedAuthorization struct {
	source WalletAuthorizationProvider
	mutate bool
}

func (d delegatedAuthorization) WalletIdentity() common.Address { return d.source.WalletIdentity() }
func (d delegatedAuthorization) Authorize(ctx context.Context, request core.WalletAuthorizationRequest) (core.WalletAuthorizationWitness, error) {
	if d.mutate {
		request.Words[0].SetUint64(123)
	}
	return d.source.Authorize(ctx, request)
}
func TestExplicitPerCallAuthorizationAndProver(t *testing.T) {
	invisi, input, backend := explicitFixture(t)
	invisi.options.Prover = nil
	prepared, err := invisi.PrepareTransfer(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	request, err := prepared.ExportWalletAuthorization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	typed, err := request.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	typed.Message["statementHash"] = common.HexToHash("0x1234").Hex()
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		t.Fatal(err)
	}
	key, _ := crypto.HexToECDSA(strings.Repeat("0", 63) + "1")
	altered, _ := crypto.Sign(digest, key)
	if _, err = prepared.WithWalletSignature(context.Background(), altered); err == nil {
		t.Fatal("changed external authorization accepted")
	}
	authorized, err := externallyAuthorize(context.Background(), prepared)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = authorized.Prove(context.Background(), TransactionProveOptions{Prover: backend}); err != nil {
		t.Fatal(err)
	}
}

type fundingProvider struct {
	sessionProvider
	payer     common.Address
	contract  bool
	reject    bool
	transport bool
}

func (p fundingProvider) CodeAt(ctx context.Context, address common.Address, block *big.Int) ([]byte, error) {
	if address == p.payer && !p.contract {
		return nil, nil
	}
	return p.sessionProvider.CodeAt(ctx, address, block)
}
func (p fundingProvider) CallContract(ctx context.Context, message ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if message.To != nil && *message.To == p.payer {
		if p.transport {
			return nil, errors.New("RPC connection interrupted")
		}
		method, _ := abi.JSON(strings.NewReader(`[{"type":"function","name":"isValidSignature","inputs":[{"type":"bytes32"},{"type":"bytes"}],"outputs":[{"type":"bytes4"}]}]`))
		value := [4]byte{0x16, 0x26, 0xba, 0x7e}
		if p.reject {
			value = [4]byte{}
		}
		return method.Methods["isValidSignature"].Outputs.Pack(value)
	}
	return p.sessionProvider.CallContract(ctx, message, block)
}
func TestEOAAndERC1271FundingVerification(t *testing.T) {
	session, token, _ := testSession(t)
	private, _ := crypto.HexToECDSA(strings.Repeat("0", 63) + "1")
	payer := crypto.PubkeyToAddress(private.PublicKey)
	provider := fundingProvider{sessionProvider: session.options.Provider.(sessionProvider), payer: payer}
	data, err := protocol.EncodeFundingTypedData(protocol.TransactionContext{ChainID: big.NewInt(31337), ProtocolAddress: session.scope.ProtocolAddress, TokenAddress: token}, "Shield", common.Hash{1}, protocol.FundingParams{Payer: payer, Amount: big.NewInt(30), SigNonce: new(big.Int), Deadline: big.NewInt(400)})
	if err != nil {
		t.Fatal(err)
	}
	digest, _, _ := apitypes.TypedDataAndHash(data)
	signature, _ := crypto.Sign(digest, private)
	valid, err := (GethFundingVerifier{Provider: provider}).VerifyTypedData(context.Background(), payer, data, signature)
	if err != nil || !valid {
		t.Fatalf("EOA: %v %v", valid, err)
	}
	provider.contract = true
	valid, err = (GethFundingVerifier{Provider: provider}).VerifyTypedData(context.Background(), payer, data, []byte{1, 2, 3})
	if err != nil || !valid {
		t.Fatalf("ERC1271: %v %v", valid, err)
	}
	provider.reject = true
	valid, err = (GethFundingVerifier{Provider: provider}).VerifyTypedData(context.Background(), payer, data, []byte{1})
	if err != nil || valid {
		t.Fatal("contract rejection accepted")
	}
	provider.transport = true
	if _, err = (GethFundingVerifier{Provider: provider}).VerifyTypedData(context.Background(), payer, data, []byte{1}); err == nil {
		t.Fatal("transport error hidden")
	}
}
func TestPerCallContractFundingSigner(t *testing.T) {
	invisi, input, _ := explicitFixture(t)
	provider := fundingProvider{sessionProvider: invisi.options.Provider.(sessionProvider), payer: input.Wallet.identity, contract: true}
	invisi.options.Provider = provider
	prepared, err := invisi.PrepareShield(context.Background(), ExplicitShieldPreparation{TransactionIdentity: input.TransactionIdentity, ShieldPreparation: ShieldPreparation{TokenAddress: input.TokenAddress, Outputs: input.Outputs, Funding: FundingSource{Payer: input.Wallet.identity, Amount: big.NewInt(30), Authorization: "signature"}}})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := prepared.WithFundingSignature(context.Background(), []byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ready.ExportTransaction(context.Background()); err != nil {
		t.Fatal(err)
	}
}
