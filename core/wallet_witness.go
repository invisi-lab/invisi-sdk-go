package core

import (
	"bytes"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// WalletAuthorizationWitness is sensitive local circuit data. Fixed byte arrays
// encode as JSON numeric arrays, as required by Noir, rather than Base64 strings.
type WalletAuthorizationWitness struct {
	PublicKeyX [32]byte `json:"public_key_x"`
	PublicKeyY [32]byte `json:"public_key_y"`
	Signature  [64]byte `json:"signature"`
}

// BuildWalletAuthorizationWitness validates the key group and signs locally.
// The signature stays in the local proving witness.
func BuildWalletAuthorizationWitness(wallet *InvisiWallet, request WalletAuthorizationRequest) (WalletAuthorizationWitness, error) {
	var witness WalletAuthorizationWitness
	if wallet == nil {
		return witness, fmt.Errorf("wallet is required")
	}
	private := append([]byte(nil), wallet.SpendingPrivateKey...)
	defer clear(private)
	public, err := walletPublicKey(append([]byte(nil), wallet.SpendingPublicKey...))
	if err != nil {
		return witness, err
	}
	key, err := ethcrypto.ToECDSA(private)
	if err != nil {
		return witness, err
	}
	encoded := ethcrypto.FromECDSAPub(public)
	if !bytes.Equal(encoded, ethcrypto.FromECDSAPub(&key.PublicKey)) {
		return witness, fmt.Errorf("spending key pair mismatch")
	}
	identity, err := ComputeWalletIdentity(encoded)
	if err != nil {
		return witness, err
	}
	if identity != wallet.WalletIdentity {
		return witness, fmt.Errorf("wallet identity mismatch")
	}
	signature, err := SignWalletAuthorization(private, request)
	if err != nil {
		return witness, err
	}
	copy(witness.PublicKeyX[:], encoded[1:33])
	copy(witness.PublicKeyY[:], encoded[33:])
	copy(witness.Signature[:], signature)
	return witness, nil
}
