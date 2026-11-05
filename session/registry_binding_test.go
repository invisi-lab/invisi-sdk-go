package session

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	sdk "github.com/invisi-lab/invisi-sdk-go/sdk"
)

func TestRegistryValidatesImmutableProtocolBinding(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	registryProvider(registry).Provider = registryBindingProvider{Provider: registryProvider(registry).Provider, selected: common.HexToAddress("0x9999")}
	if err := registry.ValidateDeployment(context.Background()); err == nil {
		t.Fatal("registry selected outside immutable protocol binding")
	}
	registryProvider(registry).Provider = registryBindingProvider{Provider: registryProvider(registry).Provider, selected: registryProvider(registry).deployment.Address}
	if err := registry.ValidateDeployment(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestProtocolDeploymentValidatesImmutableRegistryBinding(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	deployment := session.protocol.Deployment()
	deployment.Registry = &registryProvider(registry).deployment
	pool, err := sdk.NewProtocol(deployment)
	if err != nil {
		t.Fatal(err)
	}
	provider := registryBindingProvider{Provider: session.options.Provider, selected: common.HexToAddress("0x9999")}
	if err = pool.ValidateDeployment(context.Background(), provider); err == nil {
		t.Fatal("shared protocol validation accepted a mismatched registry")
	}
	session.protocol = pool
	session.options.Provider = provider
	if _, err = session.runtimeSnapshot().ResolveRecipient(context.Background(), session.wallet.OwnerAddress().Hex()); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatal("default resolver did not enforce immutable registry binding", err)
	}
}

func TestInvisiDiscoversRegistryAndExportsPortableRequest(t *testing.T) {
	session, _, _ := testSession(t)
	registry := testRegistry(t, session, false)
	ctx := context.Background()
	deployment := session.protocol.Deployment()
	deployment.Registry = &registryProvider(registry).deployment
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     uint64
			Method string
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		requests++
		if request.Method != "invisi_getDeployment" {
			t.Error("unexpected RPC", request.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"apiVersion": "v1", "protocolVersion": "1", "deployment": deployment}})
	}))
	defer server.Close()
	rpc, err := sdk.NewProtocolRpcClient(server.URL + "/v1")
	if err != nil {
		t.Fatal(err)
	}
	invisi := sdk.NewInvisi(sdk.InvisiOptions{Provider: registryProvider(registry).Provider, ProtocolRPC: rpc})
	operation, err := invisi.PrepareRegistration(ctx, sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	if err != nil {
		t.Fatal(err)
	}
	request, err := operation.ExportRegistrationAuthorization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var restored core.SigningRequest
	if err = json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if requests != 1 || restored.Domain["chainId"] != "31337" || restored.Domain["verifyingContract"] != registryProvider(registry).deployment.Address.Hex() {
		t.Fatal("wrong discovery or portable domain")
	}
	if _, err = operation.WithSignature(ctx, signRegistry(t, operation, session.wallet)); err != nil {
		t.Fatal(err)
	}
}

func TestRegistrationCompositionNormalizesMissingCapabilities(t *testing.T) {
	session, _, _ := testSession(t)
	_, err := sdk.NewInvisi(sdk.InvisiOptions{}).PrepareRegistration(context.Background(), sdk.RegistrationPreparation{Wallet: session.wallet.PublicWallet})
	var failure *sdk.InvisiError
	if !errors.As(err, &failure) || failure.Code != sdk.MissingCapability || failure.Operation != "prepareRegistration" || failure.Capability != "Chain RPC provider" {
		t.Fatal("missing capability lost at composition boundary", err)
	}
}
