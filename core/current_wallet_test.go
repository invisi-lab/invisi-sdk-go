package core

import (
	"encoding/hex"
	"testing"
)

func TestEVMAccountKeyRestoresCurrentWallet(t *testing.T) {
	key := make([]byte, 32)
	key[31] = 1
	wallet, err := DeriveWallet(key)
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(wallet.ViewingPrivateKey) != "80d9cd50805bf2ff5f95f9a091add4515bebb62293c721362fd077891c23069f" {
		t.Fatal("viewing KDF differs from TS SDK")
	}
}
