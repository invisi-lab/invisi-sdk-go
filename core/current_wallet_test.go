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

func TestRegistrationIsNotPrivateWalletAuthorization(t *testing.T) {
	for _, operation := range []WalletOperation{"register-address", "update-address"} {
		request := WalletAuthorizationRequest{Context: ProtocolContext{ChainID: ElementFromUint64(1), ProtocolAddress: ElementFromUint64(2)}, Operation: operation}
		if _, err := WalletAuthorizationTypedData(request); err == nil {
			t.Fatal("registry authorization remained in private wallet operations")
		}
	}
}
