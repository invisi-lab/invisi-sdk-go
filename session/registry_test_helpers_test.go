package session

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	protocolrpc "github.com/invisi-lab/invisi-sdk-go/rpc"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func testRegistry(t *testing.T, session *WalletSession, contract bool) *sdk.AddressRegistry {
	t.Helper()
	deployment := session.protocol.Deployment()
	deployment.Registry = &protocolrpc.ContractDeployment{Address: common.HexToAddress("0x3333333333333333333333333333333333333333"), RuntimeCodeHash: transaction.HashCode([]byte{0x60, 0}), FromBlock: 1}
	pool, err := sdk.NewProtocol(deployment)
	if err != nil {
		t.Fatal(err)
	}
	provider := &registryHarnessProvider{Provider: fundingProvider{sessionProvider: session.options.Provider.(sessionProvider), payer: session.wallet.OwnerAddress(), contract: contract}, deployment: *deployment.Registry}
	registry, err := sdk.NewAddressRegistry(pool, provider)
	if err != nil {
		t.Fatal(err)
	}
	registryHarness.Store(registry, provider)
	return registry
}

func signRegistry(t *testing.T, o *sdk.RegistrationOperation, wallet *sdk.Wallet) []byte {
	t.Helper()
	request, err := o.ExportRegistrationAuthorization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	typed, err := request.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	digest, _, err := apitypes.TypedDataAndHash(typed)
	if err != nil {
		t.Fatal(err)
	}
	key, err := crypto.ToECDSA(transaction.PrivateMaterial(wallet).SpendingPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := crypto.Sign(digest, key)
	if err != nil {
		t.Fatal(err)
	}
	return signature
}

type registryBindingProvider struct {
	sdk.Provider
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

type registryNonceProvider struct {
	sdk.Provider
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

type registrationReceiptProvider struct {
	sdk.Provider
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

func registeredOperationFixture(t *testing.T) (*sdk.RegistrationOperation, *registrationReceiptProvider) {
	t.Helper()
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ready, err := registry.PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	ready, err = ready.WithSignature(context.Background(), signRegistry(t, ready, session.wallet))
	if err != nil {
		t.Fatal(err)
	}
	call, err := ready.ExportTransaction(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	tx := types.NewTx(&types.LegacyTx{To: &call.To, Data: call.Data})
	head := types.CopyHeader(registryProvider(registry).Provider.(fundingProvider).head)
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	event := parsed.Events["ReceivingKeysSet"]
	data, err := event.Inputs.NonIndexed().Pack(ready.Preview().ReceivingKeys.ViewingPublicKey, [32]byte(ready.Preview().ReceivingKeys.ClueKey), ready.Preview().Nonce+1)
	if err != nil {
		t.Fatal(err)
	}
	log := &types.Log{Address: call.To, Topics: []common.Hash{event.ID, common.BytesToHash(ready.Preview().OwnerAddress.Bytes())}, Data: data, BlockHash: head.Hash(), BlockNumber: head.Number.Uint64(), TxHash: tx.Hash()}
	receipt := &types.Receipt{TxHash: tx.Hash(), Status: types.ReceiptStatusSuccessful, BlockNumber: new(big.Int).Set(head.Number), BlockHash: head.Hash(), Logs: []*types.Log{log}}
	provider := &registrationReceiptProvider{Provider: registryProvider(registry).Provider, transaction: tx, receipt: receipt, header: head}
	registryProvider(registry).Provider = provider
	tracked, err := ready.TrackTransaction(context.Background(), tx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	return tracked, provider
}

type erc1271ResponseProvider struct {
	sdk.Provider
	owner    common.Address
	response []byte
}

func (p erc1271ResponseProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if msg.To != nil && *msg.To == p.owner {
		return append([]byte(nil), p.response...), nil
	}
	return p.Provider.CallContract(ctx, msg, block)
}

type registeredStateProvider struct {
	sdk.Provider
	nonce uint64
}

func (p registeredStateProvider) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if len(msg.Data) >= 4 {
		method, err := parsed.MethodById(msg.Data[:4])
		if err == nil {
			switch method.RawName {
			case "isRegistered":
				return method.Outputs.Pack(p.nonce > 0)
			case "nonceOf":
				return method.Outputs.Pack(p.nonce)
			}
		}
	}
	return p.Provider.CallContract(ctx, msg, block)
}

type registryHarnessProvider struct {
	sdk.Provider
	deployment protocolrpc.ContractDeployment
}

var registryHarness sync.Map

func registryProvider(registry *sdk.AddressRegistry) *registryHarnessProvider {
	value, _ := registryHarness.Load(registry)
	return value.(*registryHarnessProvider)
}

func (p *registryHarnessProvider) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	reader, ok := p.Provider.(interface {
		TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
	})
	if !ok {
		return nil, false, errors.New("provider must authenticate registration calldata")
	}
	return reader.TransactionByHash(ctx, hash)
}
