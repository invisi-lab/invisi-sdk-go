package gas

import (
	"testing"

	"github.com/invisi-lab/invisi-sdk-go/protocol/release"
)

func TestCreditReferencesRequireExactManifestIdentity(t *testing.T) {
	current := ReleaseIdentity{CreditManifestVersion: release.CreditManifestVersion, CreditFormatVersion: release.CreditFormatVersion, CreditManifestSHA256: release.CreditManifestSHA256}
	if !currentCreditRelease(&current) {
		t.Fatal("current manifest identity rejected")
	}
	for _, sha := range []string{"", "00", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"} {
		bad := current
		bad.CreditManifestSHA256 = sha
		if currentCreditRelease(&bad) {
			t.Fatal("reference accepted missing or mismatched manifest", sha)
		}
	}
	if currentCreditRelease(nil) {
		t.Fatal("missing release identity accepted")
	}
}
