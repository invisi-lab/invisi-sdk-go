package core

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// vectorsFile returns the absolute path to testdata/vectors.json
// relative to this source file.
func vectorsFile() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", "vectors.json")
}

// ---------- JSON schema types ----------

type poseidon2Vector struct {
	Inputs   []string `json:"inputs"`
	Expected string   `json:"expected"`
}

type deriveNullifierKeyVector struct {
	MasterKey string `json:"masterKey"`
	InvoiceID string `json:"invoiceId"`
	Expected  string `json:"expected"`
}

type deriveCreditNullifierKeyVector struct {
	ReceiverMasterKey string `json:"receiverMasterKey"`
	CreditID          string `json:"creditId"`
	Expected          string `json:"expected"`
}

type deriveChangeNullifierKeyVector struct {
	PayerMasterKey    string `json:"payerMasterKey"`
	PaymentCommitment string `json:"paymentCommitment"`
	Expected          string `json:"expected"`
}

type computeIdentityCommitmentVector struct {
	NullifierKey     string `json:"nullifierKey"`
	ReceivingAddress string `json:"receivingAddress"`
	Expected         string `json:"expected"`
}

type computePaymentCommitmentVector struct {
	ChainID            string `json:"chainId"`
	IdentityCommitment string `json:"identityCommitment"`
	TokenAddress       string `json:"tokenAddress"`
	Amount             string `json:"amount"`
	Nonce              string `json:"nonce"`
	Expected           string `json:"expected"`
}

type computePaymentNullifierVector struct {
	ChainID      string `json:"chainId"`
	NullifierKey string `json:"nullifierKey"`
	TokenAddress string `json:"tokenAddress"`
	Amount       string `json:"amount"`
	Nonce        string `json:"nonce"`
	Expected     string `json:"expected"`
}

type computeNullifierKeyHashVector struct {
	NullifierKey string `json:"nullifierKey"`
	Expected     string `json:"expected"`
}

type computeCreditCommitmentVector struct {
	ChainID            string `json:"chainId"`
	IdentityCommitment string `json:"identityCommitment"`
	PayerCommitment    string `json:"payerCommitment"`
	TokenAddress       string `json:"tokenAddress"`
	Amount             string `json:"amount"`
	Nonce              string `json:"nonce"`
	Expiry             string `json:"expiry"`
	NullifierKeyHash   string `json:"nullifierKeyHash"`
	Expected           string `json:"expected"`
}

type computeCreditNullifierVector struct {
	ChainID      string `json:"chainId"`
	NullifierKey string `json:"nullifierKey"`
	TokenAddress string `json:"tokenAddress"`
	Amount       string `json:"amount"`
	Nonce        string `json:"nonce"`
	Expiry       string `json:"expiry"`
	Expected     string `json:"expected"`
}

type deriveRemainderNonceVector struct {
	Nonce    string `json:"nonce"`
	NewNonce string `json:"newNonce"`
	Expected string `json:"expected"`
}

type deriveRemainderNullifierKeyVector struct {
	NullifierKey string `json:"nullifierKey"`
	NewNonce     string `json:"newNonce"`
	Expected     string `json:"expected"`
}

type computeRefundLinkVector struct {
	OriginalCommitment string `json:"originalCommitment"`
	RefundNonce        string `json:"refundNonce"`
	Expected           string `json:"expected"`
}

type vectorsJSON struct {
	Poseidon2                   []poseidon2Vector                   `json:"poseidon2"`
	DeriveNullifierKey          []deriveNullifierKeyVector          `json:"deriveNullifierKey"`
	DeriveCreditNullifierKey    []deriveCreditNullifierKeyVector    `json:"deriveCreditNullifierKey"`
	DeriveChangeNullifierKey    []deriveChangeNullifierKeyVector    `json:"deriveChangeNullifierKey"`
	ComputeIdentityCommitment   []computeIdentityCommitmentVector   `json:"computeIdentityCommitment"`
	ComputePaymentCommitment    []computePaymentCommitmentVector    `json:"computePaymentCommitment"`
	ComputePaymentNullifier     []computePaymentNullifierVector     `json:"computePaymentNullifier"`
	ComputeNullifierKeyHash     []computeNullifierKeyHashVector     `json:"computeNullifierKeyHash"`
	ComputeCreditCommitment     []computeCreditCommitmentVector     `json:"computeCreditCommitment"`
	ComputeCreditNullifier      []computeCreditNullifierVector      `json:"computeCreditNullifier"`
	DeriveRemainderNonce        []deriveRemainderNonceVector        `json:"deriveRemainderNonce"`
	DeriveRemainderNullifierKey []deriveRemainderNullifierKeyVector `json:"deriveRemainderNullifierKey"`
	ComputeRefundLink           []computeRefundLinkVector           `json:"computeRefundLink"`
}

// loadVectors reads and parses the vectors.json file.
// Returns nil if the file does not exist (caller should t.Skip).
func loadVectors(t *testing.T) *vectorsJSON {
	t.Helper()
	path := vectorsFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("failed to read vectors file: %v", err)
	}
	var v vectorsJSON
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("failed to parse vectors JSON: %v", err)
	}
	return &v
}

// decimalToElement converts a decimal string to a field element.
func decimalToElement(t *testing.T, s string) fr.Element {
	t.Helper()
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("invalid decimal string: %q", s)
	}
	return BigIntToElement(b)
}

// hexOrDecimalToElement converts a hex (0x...) or decimal string to a field element.
func hexOrDecimalToElement(t *testing.T, s string) fr.Element {
	t.Helper()
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		e, err := HexToElement(s)
		if err != nil {
			t.Fatalf("invalid hex element: %q: %v", s, err)
		}
		return e
	}
	return decimalToElement(t, s)
}

// normalizeHex lowercases and ensures 0x prefix + 64-char padding.
func normalizeHex(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "0x")
	for len(s) < 64 {
		s = "0" + s
	}
	return "0x" + s
}

// assertHexEqual compares two hex strings after normalization.
func assertHexEqual(t *testing.T, label, got, want string) {
	t.Helper()
	if normalizeHex(got) != normalizeHex(want) {
		t.Errorf("%s: mismatch\n  got:  %s\n  want: %s", label, normalizeHex(got), normalizeHex(want))
	}
}

// ---------- Cross-language test ----------

func TestCrossLanguagePoseidon2(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.Poseidon2 {
		inputs := make([]fr.Element, len(tc.Inputs))
		for j, s := range tc.Inputs {
			inputs[j] = decimalToElement(t, s)
		}
		got, err := Poseidon2Hash(inputs...)
		if err != nil {
			t.Fatalf("vector[%d]: Poseidon2Hash error: %v", i, err)
		}
		assertHexEqual(t, "Poseidon2Hash", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageDeriveNullifierKey(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.DeriveNullifierKey {
		masterKey := decimalToElement(t, tc.MasterKey)
		invoiceId := decimalToElement(t, tc.InvoiceID)
		got, err := DeriveNullifierKey(masterKey, invoiceId)
		if err != nil {
			t.Fatalf("vector[%d]: DeriveNullifierKey error: %v", i, err)
		}
		assertHexEqual(t, "DeriveNullifierKey", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputeIdentityCommitment(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputeIdentityCommitment {
		nullifierKey := hexOrDecimalToElement(t, tc.NullifierKey)
		receivingAddress := decimalToElement(t, tc.ReceivingAddress)
		got, err := ComputeIdentityCommitment(nullifierKey, receivingAddress)
		if err != nil {
			t.Fatalf("vector[%d]: ComputeIdentityCommitment error: %v", i, err)
		}
		assertHexEqual(t, "ComputeIdentityCommitment", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputePaymentCommitment(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputePaymentCommitment {
		chainId := decimalToElement(t, tc.ChainID)
		identityCommitment := hexOrDecimalToElement(t, tc.IdentityCommitment)
		tokenAddress := decimalToElement(t, tc.TokenAddress)
		amount := decimalToElement(t, tc.Amount)
		nonce := decimalToElement(t, tc.Nonce)
		got, err := ComputePaymentCommitment(chainId, identityCommitment, tokenAddress, amount, nonce)
		if err != nil {
			t.Fatalf("vector[%d]: ComputePaymentCommitment error: %v", i, err)
		}
		assertHexEqual(t, "ComputePaymentCommitment", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputePaymentNullifier(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputePaymentNullifier {
		chainId := decimalToElement(t, tc.ChainID)
		nullifierKey := hexOrDecimalToElement(t, tc.NullifierKey)
		tokenAddress := decimalToElement(t, tc.TokenAddress)
		amount := decimalToElement(t, tc.Amount)
		nonce := decimalToElement(t, tc.Nonce)
		got, err := ComputePaymentNullifier(chainId, nullifierKey, tokenAddress, amount, nonce)
		if err != nil {
			t.Fatalf("vector[%d]: ComputePaymentNullifier error: %v", i, err)
		}
		assertHexEqual(t, "ComputePaymentNullifier", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputeNullifierKeyHash(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputeNullifierKeyHash {
		nullifierKey := hexOrDecimalToElement(t, tc.NullifierKey)
		got, err := ComputeNullifierKeyHash(nullifierKey)
		if err != nil {
			t.Fatalf("vector[%d]: ComputeNullifierKeyHash error: %v", i, err)
		}
		assertHexEqual(t, "ComputeNullifierKeyHash", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputeCreditCommitment(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputeCreditCommitment {
		chainId := decimalToElement(t, tc.ChainID)
		identityCommitment := hexOrDecimalToElement(t, tc.IdentityCommitment)
		payerCommitment := hexOrDecimalToElement(t, tc.PayerCommitment)
		tokenAddress := decimalToElement(t, tc.TokenAddress)
		amount := decimalToElement(t, tc.Amount)
		nonce := decimalToElement(t, tc.Nonce)
		expiry := decimalToElement(t, tc.Expiry)
		nullifierKeyHash := hexOrDecimalToElement(t, tc.NullifierKeyHash)
		got, err := ComputeCreditCommitment(chainId, identityCommitment, payerCommitment, tokenAddress, amount, nonce, expiry, nullifierKeyHash)
		if err != nil {
			t.Fatalf("vector[%d]: ComputeCreditCommitment error: %v", i, err)
		}
		assertHexEqual(t, "ComputeCreditCommitment", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputeCreditNullifier(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputeCreditNullifier {
		chainId := decimalToElement(t, tc.ChainID)
		nullifierKey := hexOrDecimalToElement(t, tc.NullifierKey)
		tokenAddress := decimalToElement(t, tc.TokenAddress)
		amount := decimalToElement(t, tc.Amount)
		nonce := decimalToElement(t, tc.Nonce)
		expiry := decimalToElement(t, tc.Expiry)
		got, err := ComputeCreditNullifier(chainId, nullifierKey, tokenAddress, amount, nonce, expiry)
		if err != nil {
			t.Fatalf("vector[%d]: ComputeCreditNullifier error: %v", i, err)
		}
		assertHexEqual(t, "ComputeCreditNullifier", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageDeriveRemainderNonce(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.DeriveRemainderNonce {
		nonce := decimalToElement(t, tc.Nonce)
		newNonce := decimalToElement(t, tc.NewNonce)
		got, err := DeriveRemainderNonce(nonce, newNonce)
		if err != nil {
			t.Fatalf("vector[%d]: DeriveRemainderNonce error: %v", i, err)
		}
		assertHexEqual(t, "DeriveRemainderNonce", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageDeriveRemainderNullifierKey(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.DeriveRemainderNullifierKey {
		nullifierKey := decimalToElement(t, tc.NullifierKey)
		newNonce := decimalToElement(t, tc.NewNonce)
		got, err := DeriveRemainderNullifierKey(nullifierKey, newNonce)
		if err != nil {
			t.Fatalf("vector[%d]: DeriveRemainderNullifierKey error: %v", i, err)
		}
		assertHexEqual(t, "DeriveRemainderNullifierKey", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageComputeRefundLink(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.ComputeRefundLink {
		originalCommitment := decimalToElement(t, tc.OriginalCommitment)
		refundNonce := decimalToElement(t, tc.RefundNonce)
		got, err := ComputeRefundLink(originalCommitment, refundNonce)
		if err != nil {
			t.Fatalf("vector[%d]: ComputeRefundLink error: %v", i, err)
		}
		assertHexEqual(t, "ComputeRefundLink", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageDeriveCreditNullifierKey(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}

	for i, tc := range v.DeriveCreditNullifierKey {
		receiverMasterKey := decimalToElement(t, tc.ReceiverMasterKey)
		creditId := decimalToElement(t, tc.CreditID)
		got, err := DeriveCreditNullifierKey(receiverMasterKey, creditId)
		if err != nil {
			t.Fatalf("vector[%d]: DeriveCreditNullifierKey error: %v", i, err)
		}
		assertHexEqual(t, "DeriveCreditNullifierKey", ElementToHex(got), tc.Expected)
	}
}

func TestCrossLanguageDeriveChangeNullifierKey(t *testing.T) {
	v := loadVectors(t)
	if v == nil {
		t.Skip("testdata/vectors.json not found; run generate_vectors.ts to create it")
	}
	if len(v.DeriveChangeNullifierKey) == 0 {
		t.Fatal("missing deriveChangeNullifierKey vectors")
	}

	for i, tc := range v.DeriveChangeNullifierKey {
		payerMasterKey := hexOrDecimalToElement(t, tc.PayerMasterKey)
		paymentCommitment := hexOrDecimalToElement(t, tc.PaymentCommitment)
		got, err := DeriveChangeNullifierKey(payerMasterKey, paymentCommitment)
		if err != nil {
			t.Fatalf("vector[%d]: DeriveChangeNullifierKey error: %v", i, err)
		}
		assertHexEqual(t, "DeriveChangeNullifierKey", ElementToHex(got), tc.Expected)
	}
}
