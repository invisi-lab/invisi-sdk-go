package core

import (
	"bytes"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"math/big"

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

// BuildWalletAuthorizationWitnessFromSignature validates an externally obtained EOA signature.
// It reconstructs the local circuit witness without accepting or accessing a private key.
func BuildWalletAuthorizationWitnessFromSignature(identity fr.Element, request WalletAuthorizationRequest, signature []byte) (WalletAuthorizationWitness, error) {
	var witness WalletAuthorizationWitness
	if len(signature) != 65 && len(signature) != 64 {
		return witness, fmt.Errorf("wallet signature must contain 64 compact or 65 recoverable bytes")
	}
	signed := append([]byte(nil), signature...)
	if len(signed) == 64 {
		parity := signed[32] >> 7
		signed[32] &= 0x7f
		signed = append(signed, parity)
	}
	if signed[64] >= 27 {
		signed[64] -= 27
	}
	if signed[64] > 1 || !ethcrypto.ValidateSignatureValues(signed[64], new(big.Int).SetBytes(signed[:32]), new(big.Int).SetBytes(signed[32:64]), true) {
		return witness, fmt.Errorf("invalid canonical wallet signature")
	}
	digest, err := ComputeWalletAuthorizationDigest(request)
	if err != nil {
		return witness, err
	}
	public, err := ethcrypto.SigToPub(digest[:], signed)
	if err != nil {
		return witness, fmt.Errorf("invalid wallet signature: %w", err)
	}
	encoded := ethcrypto.FromECDSAPub(public)
	owner, err := ComputeWalletIdentity(encoded)
	if err != nil {
		return witness, err
	}
	if owner != identity {
		return witness, fmt.Errorf("wallet signature identity mismatch")
	}
	copy(witness.PublicKeyX[:], encoded[1:33])
	copy(witness.PublicKeyY[:], encoded[33:])
	copy(witness.Signature[:], signed[:64])
	return witness, nil
}
