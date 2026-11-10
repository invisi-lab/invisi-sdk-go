package provenance_test

import (
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/protocol/provenance"
)

func TestManifestBytesAuthenticatedAndIndependent(t *testing.T) {
	const pin = "6107789c5c0d24c4130bff05da6d389336feb5ccb247c4030eae5198ab05d762"
	first := provenance.ManifestBytes()
	identity, err := provenance.AuthenticateManifest(first, pin)
	if err != nil || identity.SHA256 != pin {
		t.Fatalf("reviewed manifest authentication: %v", err)
	}
	first[0] ^= 0xff
	if _, err := provenance.AuthenticateManifest(first, pin); err == nil {
		t.Fatal("mutated manifest accepted")
	}
	if _, err := provenance.AuthenticateManifest(provenance.ManifestBytes(), pin); err != nil {
		t.Fatalf("caller mutation changed embedded manifest: %v", err)
	}
}
