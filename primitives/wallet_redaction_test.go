package primitives

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestDerivedWalletDefaultRepresentationsHidePrivateMaterial(t *testing.T) {
	wallet, err := DeriveWallet(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(wallet)
	if err != nil {
		t.Fatal(err)
	}
	representations := []string{fmt.Sprintf("%v", wallet), fmt.Sprintf("%+v", wallet), fmt.Sprintf("%#v", wallet), string(encoded), fmt.Sprintf("%v", *wallet), fmt.Sprintf("%+v", *wallet), fmt.Sprintf("%#v", *wallet)}
	for _, value := range representations {
		for _, name := range []string{"SpendingPrivateKey", "ViewingPrivateKey", "DetectionSeed", "DetectionKey", "spendingPrivateKey", "viewingPrivateKey"} {
			if strings.Contains(value, name) {
				t.Fatalf("private field name leaked: %s", value)
			}
		}
		for _, secret := range [][]byte{wallet.SpendingPrivateKey, wallet.ViewingPrivateKey, wallet.DetectionSeed, wallet.DetectionKey} {
			if strings.Contains(value, hex.EncodeToString(secret)) {
				t.Fatal("private key leaked")
			}
		}
	}
	var public map[string]string
	if err = json.Unmarshal(encoded, &public); err != nil {
		t.Fatal(err)
	}
	if len(public) != 1 || public["ownerAddress"] == "" {
		t.Fatal("expected public identity only")
	}
}
