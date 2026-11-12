package transaction

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"math/big"
	"testing"
)

func TestProvenanceAutomaticallyComposesNativeCapabilityAndPreservesPrivateTrust(t *testing.T) {
	backend, err := native.NewNativeProver(native.NativeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	identity, _, err := backend.GetProvenanceCapability()
	if err != nil {
		t.Fatal(err)
	}
	private := &provenance.Client{Endpoint: "https://private.invalid", Trust: provenance.TrustContext{Scope: provenance.Scope{ChainID: "1", ProtocolAddress: "2"}, ManifestSHA256: identity.SHA256}}
	invisi := NewInvisi(InvisiOptions{Prover: backend, Provenance: &ProvenanceCapability{Client: private}})
	execution := Execution{Runtime: &Runtime{Options: RuntimeOptions{InvisiOptions: invisi.GetInfrastructure()}}, Plan: Plan{Context: protocol.TransactionContext{ChainID: big.NewInt(1), ProtocolAddress: common.BigToAddress(big.NewInt(2))}}}
	capability, err := execution.sourceCapability(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if capability.Client != private || capability.Prover == nil || capability.Manifest.SHA256 != identity.SHA256 {
		t.Fatal("private trust or automatic native capability lost")
	}
	if invisi.options.Provenance.Prover != nil {
		t.Fatal("caller capability was mutated")
	}
	other := NewInvisi(InvisiOptions{})
	if invisi.options.Provenance.official == other.options.Provenance.official {
		t.Fatal("client instances share trust cache")
	}
	execution.Plan.Context.ProtocolAddress = common.BigToAddress(big.NewInt(3))
	if _, err := execution.sourceCapability(context.Background()); err == nil {
		t.Fatal("different deployment accepted")
	}
}
