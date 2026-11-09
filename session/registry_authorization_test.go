package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	transaction "github.com/invisi-lab/invisi-sdk-go/internal/transaction"
	"github.com/invisi-lab/invisi-sdk-go/protocol/contracts"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestRegistrationExternalSignerUsesFrozenEIP712Data(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ctx := context.Background()
	address, _ := session.wallet.GetAddress()
	watch, _ := sdk.PublicWalletFromAddress(address)
	prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: watch})
	if err != nil {
		t.Fatal(err)
	}
	request, err := prepared.ExportRegistrationAuthorization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if request.PrimaryType != "RegisterReceivingKeys" || request.Domain["name"] != "InvisiAddressRegistry" {
		t.Fatal(request)
	}
	request.Message["nonce"] = "1"
	typed, err := request.TypedData()
	if err != nil {
		t.Fatal(err)
	}
	digest, _, _ := apitypes.TypedDataAndHash(typed)
	key, _ := crypto.ToECDSA(transaction.PrivateMaterial(session.wallet).SpendingPrivateKey)
	bad, _ := crypto.Sign(digest, key)
	if _, err = prepared.WithSignature(ctx, bad); err == nil {
		t.Fatal("signature for changed nonce accepted")
	}
	if prepared.Status() != "prepared" {
		t.Fatal("failed signature changed stage")
	}
	signature := signRegistry(t, prepared, session.wallet)
	ready, err := prepared.WithSignature(ctx, signature)
	if err != nil {
		t.Fatal(err)
	}
	signature[0] ^= 1
	call, err := ready.ExportTransaction(ctx)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
	values, err := parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
	if err != nil || len(values[1].([]byte)) != 65 || values[1].([]byte)[64] < 27 {
		t.Fatal("expected canonical EOA signature", err)
	}
}

func TestRegistrationAcceptsArbitraryERC1271SignatureAndRejectsInvalidEOA(t *testing.T) {
	session, _, _ := testSession(t)
	ctx := context.Background()
	for _, contract := range []bool{false, true} {
		registry := testRegistry(t, session, contract)
		prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
		if err != nil {
			t.Fatal(err)
		}
		signature := []byte{1, 2, 3, 4, 5}
		ready, err := prepared.WithSignature(ctx, signature)
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
	prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.WithSignature(ctx, nil); err == nil {
		t.Fatal("high-level direct mode must be explicit or unavailable")
	}
	wrong, _ := crypto.GenerateKey()
	data, _ := prepared.TypedData()
	digest, _, _ := apitypes.TypedDataAndHash(data)
	signature, _ := crypto.Sign(digest, wrong)
	if _, err = prepared.WithSignature(ctx, signature); err == nil {
		t.Fatal("wrong signer accepted")
	}
	cancel, done := context.WithCancel(ctx)
	done()
	if _, err = prepared.WithSignature(cancel, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	transaction.ReceivingKeySnapshot(session.wallet.PublicWallet).ViewingPublicKey = make([]byte, 33)
	if _, err = registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet}); err == nil {
		t.Fatal("invalid viewing key accepted at preparation")
	}
}

func TestRegistrationExportRejectsExpiry(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	prepared, err := registry.PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet, Deadline: 101})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := prepared.WithSignature(context.Background(), signRegistry(t, prepared, session.wallet))
	if err != nil {
		t.Fatal(err)
	}
	registryProvider(registry).Provider.(fundingProvider).head.Time = 102
	if _, err = ready.ExportTransaction(context.Background()); err == nil {
		t.Fatal("expired signature exported")
	}
}

func TestRegistrationAuthorizationRejectsChangedNonceAndExpiry(t *testing.T) {
	for _, change := range []string{"nonce", "expiry"} {
		t.Run(change, func(t *testing.T) {
			session, _, _ := testSession(t)
			registry := testRegistry(t, session, false)
			ctx := context.Background()
			prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet, Deadline: 101})
			if err != nil {
				t.Fatal(err)
			}
			signature := signRegistry(t, prepared, session.wallet)
			if change == "nonce" {
				registryProvider(registry).Provider = registryNonceProvider{Provider: registryProvider(registry).Provider, nonce: 1}
			} else {
				registryProvider(registry).Provider.(fundingProvider).head.Time = 102
			}
			if _, err = prepared.WithSignature(ctx, signature); err == nil {
				t.Fatal("stale registration accepted")
			}
			if _, err = prepared.ExportRegistrationAuthorization(ctx); err == nil {
				t.Fatal("stale request exported")
			}
		})
	}
}

func TestRegistrationAcceptsEOASignatureWithDelegatedOwnerCode(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, true)
	provider := registryProvider(registry).Provider.(fundingProvider)
	provider.transport = true // Any ERC-1271 call would fail.
	registryProvider(registry).Provider = provider
	prepared, err := registry.PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = prepared.WithSignature(context.Background(), signRegistry(t, prepared, session.wallet)); err != nil {
		t.Fatal("delegated EOA recovery must precede ERC-1271", err)
	}
}

func TestRegistrationRemoteRoundTripCannotMakeExpiredRequestReady(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ctx := context.Background()
	prepared, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet, Deadline: 101})
	if err != nil {
		t.Fatal(err)
	}
	signature := signRegistry(t, prepared, session.wallet)
	registryProvider(registry).Provider.(fundingProvider).head.Time = 102
	if _, err = prepared.WithSignature(ctx, signature); err == nil {
		t.Fatal("expired request became ready")
	}
	if prepared.Status() != "prepared" {
		t.Fatal("failed authorization changed immutable handle")
	}
}

func TestRegistrationNormalizesOnlyRecoverableOwnerSignature(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	prepared, err := registry.PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	typed, _ := prepared.TypedData()
	digest, _, _ := apitypes.TypedDataAndHash(typed)
	key, _ := crypto.ToECDSA(transaction.PrivateMaterial(session.wallet).SpendingPrivateKey)
	signature, _ := crypto.Sign(digest, key)
	ready, err := prepared.WithSignature(context.Background(), signature)
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
	prepared, err = registry.PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	opaque := make([]byte, 65)
	opaque[0] = 7
	ready, err = prepared.WithSignature(context.Background(), opaque)
	if err != nil {
		t.Fatal(err)
	}
	call, _ = ready.ExportTransaction(context.Background())
	args, _ = parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
	if !bytes.Equal(args[1].([]byte), opaque) {
		t.Fatal("opaque contract signature was normalized")
	}
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
			registryProvider(registry).Provider = erc1271ResponseProvider{Provider: registryProvider(registry).Provider, owner: session.wallet.OwnerAddress(), response: response}
			prepared, err := registry.PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
			if err != nil {
				t.Fatal(err)
			}
			_, err = prepared.WithSignature(context.Background(), []byte{1, 2, 3})
			valid := shape == "canonical" || shape == "trailing bytes"
			if (err == nil) != valid {
				t.Fatalf("ERC1271 return %s validity differs from registry: %v", shape, err)
			}
		})
	}
}

func TestRegistrationDirectCallRequiresOwnerSender(t *testing.T) {
	for _, ownerSender := range []bool{false, true} {
		t.Run(fmt.Sprint(ownerSender), func(t *testing.T) {
			session, _, _ := testSession(t)
			registry := testRegistry(t, session, false)
			ctx := context.Background()
			operation, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
			if err != nil {
				t.Fatal(err)
			}
			ready, err := operation.AuthorizeDirect(ctx)
			if err != nil {
				t.Fatal(err)
			}
			call, err := ready.ExportTransaction(ctx)
			if err != nil {
				t.Fatal(err)
			}
			parsed, _ := contracts.InvisiAddressRegistryMetaData.GetAbi()
			values, err := parsed.Methods["registerKeys"].Inputs.Unpack(call.Data[4:])
			if err != nil || len(values[1].([]byte)) != 0 {
				t.Fatal("direct call signature must be empty", err)
			}
			private := transaction.PrivateMaterial(session.wallet).SpendingPrivateKey
			if !ownerSender {
				private = bytes.Repeat([]byte{2}, 32)
			}
			key, _ := crypto.ToECDSA(private)
			tx, _ := types.SignTx(types.NewTx(&types.LegacyTx{To: &call.To, Data: call.Data}), types.LatestSignerForChainID(new(big.Int).SetUint64(session.scope.ChainID)), key)
			head := types.CopyHeader(registryProvider(registry).Provider.(fundingProvider).head)
			receipt := &types.Receipt{TxHash: tx.Hash(), Status: types.ReceiptStatusFailed, BlockNumber: new(big.Int).Set(head.Number), BlockHash: head.Hash()}
			registryProvider(registry).Provider = &registrationReceiptProvider{Provider: registryProvider(registry).Provider, transaction: tx, receipt: receipt, header: head}
			tracked, err := ready.TrackTransaction(ctx, tx.Hash())
			if err != nil {
				t.Fatal(err)
			}
			reconciled, err := tracked.Reconcile(ctx)
			if ownerSender {
				if err != nil || reconciled.Status() != "failed" {
					t.Fatal("owner reverted call must reconcile as failed", err)
				}
			} else if err == nil {
				t.Fatal("different sender accepted")
			}
		})
	}
}

func TestRegistrationDefaultAndUpdateState(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ctx := context.Background()
	if _, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet, Update: true}); err == nil {
		t.Fatal("unregistered update accepted")
	}
	registryProvider(registry).Provider = registeredStateProvider{Provider: registryProvider(registry).Provider, nonce: 1}
	if _, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet}); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatal("default registration must reject registered owner", err)
	}
	update, err := registry.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet, Update: true})
	if err != nil {
		t.Fatal(err)
	}
	request, err := update.ExportRegistrationAuthorization(ctx)
	if err != nil || request.PrimaryType != "UpdateReceivingKeys" || request.Message["nonce"] != "1" {
		t.Fatal("update transcript mismatch", err)
	}
	if _, err = update.WithSignature(ctx, signRegistry(t, update, session.wallet)); err != nil {
		t.Fatal(err)
	}
}
