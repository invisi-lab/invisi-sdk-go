package payer

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

// creditFixture builds a self-consistent "accepted authorize" fixture: payerCommitment
// is derived from the payer master key, and the server-side values
// (nullifier_key_hash, credit_commitment) are computed exactly as the server would,
// so CreateCredit's two-step verification passes.
type creditFixture struct {
	params          CreateCreditParams
	creditNullKey   fr.Element
	payerCommitment common.Hash
	serverNKHash    common.Hash
}

func newCreditFixture(t *testing.T) creditFixture {
	t.Helper()

	var payerMasterKey, creditID, payerAddr, creditNullKey, nonce fr.Element
	payerMasterKey.SetUint64(111)
	creditID.SetUint64(222)
	payerAddr.SetUint64(333)
	creditNullKey.SetUint64(444) // server-derived; opaque to the payer
	nonce.SetUint64(555)

	pcHex, err := DeriveCreditAuthorizeFields(payerMasterKey, creditID, payerAddr)
	if err != nil {
		t.Fatal(err)
	}
	payerCommitment := common.HexToHash(pcHex)

	// Tag byte placed low-order (not "0xab" + zeros) so the raw 32-byte value stays
	// below the BN254 scalar field modulus once run through core.HexToElement below:
	// CreateCredit's two-step verification folds identityCommitment into
	// ComputeCreditCommitment as a field element, and a high-order tag byte overflows it.
	identityCommitment := common.HexToHash("0x" + strings.Repeat("00", 31) + "ab")
	tokenAddress := common.HexToAddress("0xcc" + strings.Repeat("00", 19))
	amount := big.NewInt(1_000_000)
	expiry := big.NewInt(1700000000)
	treeID := big.NewInt(1)
	chainID := big.NewInt(84532)

	nkHashElem, err := core.ComputeNullifierKeyHash(creditNullKey)
	if err != nil {
		t.Fatal(err)
	}
	serverNKHash := common.HexToHash(core.ElementToHex(nkHashElem))

	icElem, err := core.HexToElement(identityCommitment.Hex())
	if err != nil {
		t.Fatal(err)
	}
	pcElem, err := core.HexToElement(payerCommitment.Hex())
	if err != nil {
		t.Fatal(err)
	}
	tokenElem, err := core.HexToElement(addressToTokenElement(tokenAddress).Hex())
	if err != nil {
		t.Fatal(err)
	}
	creditCommElem, err := core.ComputeCreditCommitment(
		core.BigIntToElement(chainID), icElem, pcElem, tokenElem,
		core.BigIntToElement(amount), nonce, core.BigIntToElement(expiry), nkHashElem,
	)
	if err != nil {
		t.Fatal(err)
	}
	serverCreditComm := common.HexToHash(core.ElementToHex(creditCommElem))

	return creditFixture{
		params: CreateCreditParams{
			CreditID:               creditID,
			IdentityCommitment:     identityCommitment,
			PayerCommitment:        payerCommitment,
			CreditNullifierKey:     creditNullKey,
			TokenAddress:           tokenAddress,
			Amount:                 amount,
			Nonce:                  nonce,
			Expiry:                 expiry,
			TreeID:                 treeID,
			ChainID:                chainID,
			ServerNullifierKeyHash: serverNKHash,
			ServerCreditCommitment: serverCreditComm,
		},
		creditNullKey:   creditNullKey,
		payerCommitment: payerCommitment,
		serverNKHash:    serverNKHash,
	}
}

func TestCreateCredit_Basic(t *testing.T) {
	f := newCreditFixture(t)
	result, err := CreateCredit(f.params)
	if err != nil {
		t.Fatal(err)
	}

	if result.Authorization.PayerCommitment != f.payerCommitment {
		t.Error("payerCommitment should be echoed unchanged")
	}
	// depositCredit binds the LOCALLY recomputed hash (reclaim safety anchor).
	if result.Authorization.NullifierKeyHash != f.serverNKHash {
		t.Errorf("nullifierKeyHash = %s, want %s",
			result.Authorization.NullifierKeyHash.Hex(), f.serverNKHash.Hex())
	}
	if !result.Authorization.CreditNullifierKey.Equal(&f.creditNullKey) {
		t.Error("creditNullifierKey should be stored for reclaim")
	}
	if result.ApproveAmount.Cmp(big.NewInt(1_000_000)) != 0 {
		t.Errorf("approveAmount = %s, want 1000000", result.ApproveAmount)
	}
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
}

func TestCreateCredit_ChainIDPreserved(t *testing.T) {
	f := newCreditFixture(t)
	result, err := CreateCredit(f.params)
	if err != nil {
		t.Fatal(err)
	}
	if result.Authorization.ChainID.Int64() != 84532 {
		t.Errorf("chainId = %d, want 84532", result.Authorization.ChainID.Int64())
	}
}

// Verification #1: a credit_nullifier_key that does not hash to the server's
// nullifier_key_hash must abort with ErrCreditKeyHashMismatch — no calldata.
func TestCreateCredit_RejectsKeyHashMismatch(t *testing.T) {
	f := newCreditFixture(t)
	f.params.ServerNullifierKeyHash = common.HexToHash("0xdead" + strings.Repeat("00", 30))
	_, err := CreateCredit(f.params)
	if !errors.Is(err, ErrCreditKeyHashMismatch) {
		t.Fatalf("expected ErrCreditKeyHashMismatch, got %v", err)
	}
}

// Verification #2: a credit_commitment that does not match the local recompute
// (e.g. server folded a wrong nonce) must abort with ErrCreditCommitmentMismatch.
func TestCreateCredit_RejectsCommitmentMismatch(t *testing.T) {
	f := newCreditFixture(t)
	f.params.ServerCreditCommitment = common.HexToHash("0xbeef" + strings.Repeat("00", 30))
	_, err := CreateCredit(f.params)
	if !errors.Is(err, ErrCreditCommitmentMismatch) {
		t.Fatalf("expected ErrCreditCommitmentMismatch, got %v", err)
	}
}

func TestCreateCredit_RequiresServerFields(t *testing.T) {
	f := newCreditFixture(t)
	f.params.ServerNullifierKeyHash = common.Hash{}
	if _, err := CreateCredit(f.params); err == nil {
		t.Error("expected error when ServerNullifierKeyHash is zero")
	}
}

func TestDeriveCreditAuthorizeFields(t *testing.T) {
	var payerMasterKey, creditID, payerAddr fr.Element
	payerMasterKey.SetUint64(111)
	creditID.SetUint64(222)
	payerAddr.SetUint64(333)

	payerCommitment, err := DeriveCreditAuthorizeFields(payerMasterKey, creditID, payerAddr)
	if err != nil {
		t.Fatal(err)
	}

	hexRe := regexp.MustCompile(`^0x[0-9a-f]{64}$`)
	if !hexRe.MatchString(payerCommitment) {
		t.Errorf("payerCommitment %q does not match bytes32 hex shape", payerCommitment)
	}

	// Must equal Poseidon2(D_IDENT, Poseidon2(D_NKEY, payerMasterKey, creditId), payerAddr).
	pnk, err := core.DeriveNullifierKey(payerMasterKey, creditID)
	if err != nil {
		t.Fatal(err)
	}
	want, err := core.ComputeIdentityCommitment(pnk, payerAddr)
	if err != nil {
		t.Fatal(err)
	}
	if payerCommitment != core.ElementToHex(want) {
		t.Errorf("payerCommitment = %s, want %s", payerCommitment, core.ElementToHex(want))
	}
}

func TestBuildReclaimCalldata_Basic(t *testing.T) {
	result, err := BuildReclaimCalldata(ReclaimParams{
		Nullifier:        common.HexToHash("0xdd" + strings.Repeat("00", 31)),
		Amount:           big.NewInt(1_000_000),
		ReceivingAddress: common.HexToAddress("0xdead"),
		MerkleRoot:       common.HexToHash("0xee" + strings.Repeat("00", 31)),
		Proof:            []byte{0xde, 0xad},
		TreeID:           big.NewInt(1),
		Expiry:           big.NewInt(1700000000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) < 4 {
		t.Fatal("calldata too short")
	}
	if result.Nullifier != common.HexToHash("0xdd"+strings.Repeat("00", 31)) {
		t.Error("nullifier should be passed through unchanged")
	}
}

func TestBuildReclaimCalldata_RequiresFields(t *testing.T) {
	_, err := BuildReclaimCalldata(ReclaimParams{
		Nullifier: common.HexToHash("0x01"),
	})
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}
