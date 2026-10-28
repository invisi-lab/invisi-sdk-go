package sdk

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
)

func testRegistry(t *testing.T, session *WalletSession, contract bool) *AddressRegistry {
	t.Helper()
	deployment := session.protocol.Deployment()
	deployment.Registry = &protocolrpc.ContractDeployment{Address: common.HexToAddress("0x3333333333333333333333333333333333333333"), RuntimeCodeHash: hashCode([]byte{0x60, 0}), FromBlock: 1}
	pool, err := NewProtocol(deployment)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewAddressRegistry(pool, fundingProvider{sessionProvider: session.options.Provider.(sessionProvider), payer: session.wallet.identity, contract: contract})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestRegistrationExternalSignerUsesFrozenEIP712Data(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	wallet := session.wallet
	address, _ := wallet.GetAddress()
	watch, _ := WalletFromAddress(address)
	signer, err := wallet.ExportRegistrationAuthorization()
	if err != nil {
		t.Fatal(err)
	}
	called := false
	delegated := RegistrationAuthorizationFunc{Owner: wallet.identity, Sign: func(ctx context.Context, data apitypes.TypedData) ([]byte, error) {
		called = true
		if data.PrimaryType != "RegisterReceivingKeys" || data.Domain.Name != "InvisiAddressRegistry" || data.Domain.VerifyingContract != registry.deployment.Address.Hex() {
			t.Fatal(data)
		}
		data.Message["nonce"] = "1"
		return signer.SignTypedData(ctx, data)
	}}
	prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: watch, Authorization: delegated})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.Authorize(context.Background()); err == nil {
		t.Fatal("signature for changed nonce accepted")
	}
	if !called || prepared.Status() != "prepared" {
		t.Fatal("failed authorization changed stage")
	}
	// Offline authorization is verified against the original, unchanged statement.
	data, err := prepared.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	signature, err := signer.SignTypedData(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	ready, err := prepared.Authorize(context.Background(), signature)
	if err != nil {
		t.Fatal(err)
	}
	signature[0] ^= 1
	call, err := ready.ExportTransaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	values, err := parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
	if err != nil || len(values[1].([]byte)) != 65 || values[1].([]byte)[64] < 27 {
		t.Fatal("expected ordinary recoverable EOA signature", err)
	}
}

func TestRegistrationAcceptsArbitraryERC1271SignatureAndRejectsInvalidEOA(t *testing.T) {
	session, _, _ := testSession(t)
	ctx := context.Background()
	for _, contract := range []bool{false, true} {
		registry := testRegistry(t, session, contract)
		prepared, err := registry.PrepareRegistration(ctx, RegistrationPreparation{Wallet: session.wallet})
		if err != nil {
			t.Fatal(err)
		}
		signature := []byte{1, 2, 3, 4, 5}
		ready, err := prepared.Authorize(ctx, signature)
		if !contract {
			if err == nil {
				t.Fatal("invalid EOA signature accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		call, err := ready.ExportTransaction(ctx)
		if err != nil {
			t.Fatal(err)
		}
		parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
		values, err := parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
		if err != nil || !bytes.Equal(values[1].([]byte), signature) {
			t.Fatal("ERC1271 signature changed", err)
		}
	}
}

func TestRegistrationRequiresSignatureAndValidatesPreparedMaterial(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ctx := context.Background()
	prepared, err := registry.PrepareRegistration(ctx, RegistrationPreparation{Wallet: session.wallet})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.Authorize(ctx, nil); err == nil {
		t.Fatal("high-level direct mode must be explicit or unavailable")
	}
	wrong, _ := crypto.GenerateKey()
	data, _ := prepared.TypedData()
	digest, _, _ := apitypes.TypedDataAndHash(data)
	signature, _ := crypto.Sign(digest, wrong)
	if _, err = prepared.Authorize(ctx, signature); err == nil {
		t.Fatal("wrong signer accepted")
	}
	cancel, done := context.WithCancel(ctx)
	done()
	if _, err = prepared.Authorize(cancel); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	session.wallet.keys.ViewingPublicKey = make([]byte, 33)
	if _, err = registry.PrepareRegistration(ctx, RegistrationPreparation{Wallet: session.wallet}); err == nil {
		t.Fatal("invalid viewing key accepted at preparation")
	}
}

func TestRegistrationExportRejectsExpiry(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet, Deadline: 101})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := prepared.Authorize(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	registry.provider.(fundingProvider).head.Time = 102
	if _, err = ready.ExportTransaction(context.Background()); err == nil {
		t.Fatal("expired signature exported")
	}
}

type registryBindingProvider struct {
	Provider
	selected common.Address
}

func (p registryBindingProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	parsed, _ := protocol.ProtocolABI()
	method, err := parsed.MethodById(msg.Data[:4])
	if err == nil && method.RawName == "addressRegistry" {
		return method.Outputs.Pack(p.selected)
	}
	return p.Provider.CallContract(ctx, msg, block)
}
func TestRegistryValidatesImmutableProtocolBinding(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	registry.provider = registryBindingProvider{Provider: registry.provider, selected: common.HexToAddress("0x9999")}
	if err := registry.ValidateDeployment(context.Background()); err == nil {
		t.Fatal("registry selected outside immutable protocol binding")
	}
	registry.provider = registryBindingProvider{Provider: registry.provider, selected: registry.deployment.Address}
	if err := registry.ValidateDeployment(context.Background()); err != nil {
		t.Fatal(err)
	}
}

type registryNonceProvider struct {
	Provider
	nonce uint64
}

func (p registryNonceProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	method, err := parsed.MethodById(msg.Data[:4])
	if err == nil && method.RawName == "nonceOf" {
		return method.Outputs.Pack(p.nonce)
	}
	return p.Provider.CallContract(ctx, msg, block)
}
func TestRegistrationAuthorizationRejectsChangedNonceAndExpiry(t *testing.T) {
	for _, change := range []string{"nonce", "expiry"} {
		t.Run(change, func(t *testing.T) {
			session, _, _ := testSession(t)
			registry := testRegistry(t, session, false)
			called := false
			authority, _ := session.wallet.ExportRegistrationAuthorization()
			signer := RegistrationAuthorizationFunc{Owner: session.wallet.identity, Sign: func(ctx context.Context, data apitypes.TypedData) ([]byte, error) {
				called = true
				return authority.SignTypedData(ctx, data)
			}}
			prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet, Deadline: 101, Authorization: signer})
			if err != nil {
				t.Fatal(err)
			}
			if change == "nonce" {
				registry.provider = registryNonceProvider{Provider: registry.provider, nonce: 1}
			} else {
				registry.provider.(fundingProvider).head.Time = 102
			}
			if _, err = prepared.Authorize(context.Background()); err == nil {
				t.Fatal("stale registration signed")
			}
			if called {
				t.Fatal("signer prompted before freshness check")
			}
		})
	}
}

func TestProtocolDeploymentValidatesImmutableRegistryBinding(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	deployment := session.protocol.Deployment()
	deployment.Registry = &registry.deployment
	pool, err := NewProtocol(deployment)
	if err != nil {
		t.Fatal(err)
	}
	provider := registryBindingProvider{Provider: session.options.Provider, selected: common.HexToAddress("0x9999")}
	if err = pool.ValidateDeployment(context.Background(), provider); err == nil {
		t.Fatal("shared protocol validation accepted a mismatched registry")
	}
	session.protocol = pool
	session.options.Provider = provider
	if _, err = session.resolve(context.Background(), session.wallet.GetShortAddress()); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatal("default resolver did not enforce immutable registry binding", err)
	}
}

type registrationReceiptProvider struct {
	Provider
	receipt     *types.Receipt
	receiptErr  error
	transaction *types.Transaction
	header      *types.Header
	afterRead   func()
}

func (p *registrationReceiptProvider) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return p.receipt, p.receiptErr
}
func (p *registrationReceiptProvider) TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error) {
	if p.afterRead != nil {
		callback := p.afterRead
		p.afterRead = nil
		callback()
	}
	return p.transaction, false, nil
}
func (p *registrationReceiptProvider) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if number != nil && p.header != nil && number.Cmp(p.header.Number) == 0 {
		return p.header, nil
	}
	return p.Provider.HeaderByNumber(ctx, number)
}
func registeredOperationFixture(t *testing.T) (*RegistrationOperation, *registrationReceiptProvider) {
	t.Helper()
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ready, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet})
	if err != nil {
		t.Fatal(err)
	}
	ready, err = ready.Authorize(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	call, err := ready.ExportTransaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	tx := types.NewTx(&types.LegacyTx{To: &call.To, Data: call.Data})
	head := types.CopyHeader(registry.provider.(fundingProvider).head)
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	event := parsed.Events["ReceivingKeysSet"]
	data, err := event.Inputs.NonIndexed().Pack(ready.registration.Keys.ViewingPublicKey, ready.registration.Keys.ClueKey, ready.registration.Nonce+1)
	if err != nil {
		t.Fatal(err)
	}
	log := &types.Log{Address: call.To, Topics: []common.Hash{event.ID, ready.registration.WalletIdentity}, Data: data, BlockHash: head.Hash(), BlockNumber: head.Number.Uint64(), TxHash: tx.Hash()}
	receipt := &types.Receipt{TxHash: tx.Hash(), Status: types.ReceiptStatusSuccessful, BlockNumber: new(big.Int).Set(head.Number), BlockHash: head.Hash(), Logs: []*types.Log{log}}
	provider := &registrationReceiptProvider{Provider: registry.provider, transaction: tx, receipt: receipt, header: head}
	registry.provider = provider
	tracked, err := ready.TrackTransaction(context.Background(), tx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	return tracked, provider
}
func TestRegistrationReconciliationUsesItsOwnCanonicalEvent(t *testing.T) {
	// A later update in the same block may replace end-of-block keys. Its state
	// cannot invalidate the exact ReceivingKeysSet event in this receipt.
	tracked, _ := registeredOperationFixture(t)
	confirmed, err := tracked.Reconcile(context.Background())
	if err != nil || confirmed.Status() != "confirmed" {
		t.Fatal("canonical receiving-key event was not confirmed", err)
	}
}
func TestRegistrationReconciliationRejectsMisboundEvents(t *testing.T) {
	for _, change := range []string{"owner", "keys", "nonce", "block", "transaction", "address", "removed", "missing", "duplicate"} {
		t.Run(change, func(t *testing.T) {
			tracked, provider := registeredOperationFixture(t)
			log := provider.receipt.Logs[0]
			switch change {
			case "owner":
				log.Topics[1] = common.HexToHash("0x1")
			case "keys":
				log.Data[0] ^= 1
			case "nonce":
				log.Data[95] ^= 1
			case "block":
				log.BlockHash = common.HexToHash("0x1")
			case "transaction":
				log.TxHash = common.HexToHash("0x1")
			case "address":
				log.Address = common.HexToAddress("0x9999")
			case "removed":
				log.Removed = true
			case "missing":
				provider.receipt.Logs = nil
			case "duplicate":
				provider.receipt.Logs = append(provider.receipt.Logs, log)
			}
			if _, err := tracked.Reconcile(context.Background()); err == nil {
				t.Fatal("misbound event confirmed")
			}
		})
	}
}
func TestRegistrationConfirmationReturnsToUncertainAfterReorg(t *testing.T) {
	for _, change := range []string{"missing", "notfound", "replaced block", "during read"} {
		t.Run(change, func(t *testing.T) {
			tracked, provider := registeredOperationFixture(t)
			if change == "during read" {
				provider.afterRead = func() { provider.header = types.CopyHeader(provider.header); provider.header.Extra = []byte{1} }
			} else {
				confirmed, err := tracked.Reconcile(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				tracked = confirmed
				switch change {
				case "missing":
					provider.receipt = nil
				case "notfound":
					provider.receipt = nil
					provider.receiptErr = ethereum.NotFound
				case "replaced block":
					provider.header = types.CopyHeader(provider.header)
					provider.header.Extra = []byte{1}
				}
			}
			current, err := tracked.Reconcile(context.Background())
			if err != nil || current.Status() != "uncertain" {
				t.Fatal("orphaned receipt remained confirmed", err)
			}
		})
	}
}

func TestRegistrationAcceptsEOASignatureWithDelegatedOwnerCode(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, true)
	provider := registry.provider.(fundingProvider)
	provider.transport = true // Any ERC-1271 call would fail.
	registry.provider = provider
	prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.Authorize(context.Background()); err != nil {
		t.Fatal("delegated EOA recovery must precede ERC-1271", err)
	}
}

func TestRegistrationSignerCannotMakeExpiredRequestReady(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	source, _ := session.wallet.ExportRegistrationAuthorization()
	signer := RegistrationAuthorizationFunc{Owner: session.wallet.identity, Sign: func(ctx context.Context, data apitypes.TypedData) ([]byte, error) {
		registry.provider.(fundingProvider).head.Time = 102
		return source.SignTypedData(ctx, data)
	}}
	prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet, Deadline: 101, Authorization: signer})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.Authorize(context.Background()); err == nil {
		t.Fatal("request expired during signing was marked ready")
	}
	if prepared.Status() != "prepared" {
		t.Fatal("failed authorization changed immutable handle")
	}
}

func TestRegistrationNormalizesOnlyRecoverableOwnerSignature(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet})
	if err != nil {
		t.Fatal(err)
	}
	typed, _ := prepared.TypedData()
	digest, _, _ := apitypes.TypedDataAndHash(typed)
	key, _ := crypto.ToECDSA(session.wallet.material.SpendingPrivateKey)
	signature, _ := crypto.Sign(digest, key)
	ready, err := prepared.Authorize(context.Background(), signature)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	call, _ := ready.ExportTransaction(context.Background())
	args, _ := parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
	if args[1].([]byte)[64] != signature[64]+27 {
		t.Fatal("geth recovery ID was not normalized")
	}
	registry = testRegistry(t, session, true)
	prepared, err = registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet})
	if err != nil {
		t.Fatal(err)
	}
	opaque := make([]byte, 65)
	opaque[0] = 7
	ready, err = prepared.Authorize(context.Background(), opaque)
	if err != nil {
		t.Fatal(err)
	}
	call, _ = ready.ExportTransaction(context.Background())
	args, _ = parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
	if !bytes.Equal(args[1].([]byte), opaque) {
		t.Fatal("opaque contract signature was normalized")
	}
}

type erc1271ResponseProvider struct {
	Provider
	owner    common.Address
	response []byte
}

func (p erc1271ResponseProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if msg.To != nil && *msg.To == p.owner {
		return append([]byte(nil), p.response...), nil
	}
	return p.Provider.CallContract(ctx, msg, block)
}
func TestRegistrationChecksCompleteERC1271MagicWord(t *testing.T) {
	for _, shape := range []string{"short", "dirty padding", "canonical", "trailing bytes"} {
		t.Run(shape, func(t *testing.T) {
			session, _, _ := testSession(t)
			registry := testRegistry(t, session, true)
			response := make([]byte, 32)
			copy(response, []byte{0x16, 0x26, 0xba, 0x7e})
			switch shape {
			case "short":
				response = response[:4]
			case "dirty padding":
				response[31] = 1
			case "trailing bytes":
				response = append(response, 1)
			}
			registry.provider = erc1271ResponseProvider{Provider: registry.provider, owner: session.wallet.identity, response: response}
			prepared, err := registry.PrepareRegistration(context.Background(), RegistrationPreparation{Wallet: session.wallet})
			if err != nil {
				t.Fatal(err)
			}
			_, err = prepared.Authorize(context.Background(), []byte{1, 2, 3})
			valid := shape == "canonical" || shape == "trailing bytes"
			if (err == nil) != valid {
				t.Fatalf("ERC1271 return %s validity differs from registry: %v", shape, err)
			}
		})
	}
}
