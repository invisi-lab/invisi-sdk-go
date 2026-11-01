// Package shared contains editable public settings, local account loading and authenticated discovery.
package shared

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/invisi-lab/invisi-sdk-go/prover/native"
	"github.com/invisi-lab/invisi-sdk-go/sdk"
)

// Edit these public values for your deployment. Run examples from the repository root.
var Settings = struct {
	ChainID                                                               uint64
	NativeCurrencySymbol                                                  string
	NativeCurrencyDecimals                                                uint8
	SuggestedSubmitterGasBalance                                          string
	SubmitterNames                                                        []string
	ChainRPCURL, ProtocolRPCURL, TokenAddress, AccountsFile, DatabaseFile string
	Prover                                                                native.NativeOptions
}{
	ChainID: 97, NativeCurrencySymbol: "tBNB", NativeCurrencyDecimals: 18,
	SuggestedSubmitterGasBalance: "0.1", SubmitterNames: []string{"Alice", "Bob", "Charlie", "Frank"},
	ChainRPCURL: "http://127.0.0.1:18565", ProtocolRPCURL: "https://rpc.invisipay.fi/v1/chains/97",
	TokenAddress: "0xf78DB005c9af4e9Ce4ebE6A5C162f9C42f2E4b65", AccountsFile: "examples/accounts.json", DatabaseFile: "examples/wallet.sqlite",
	Prover: native.NativeOptions{BBBinary: "bb", NoirBinary: "noir-execute", Threads: 4, Timeout: time.Hour},
}

type accountFile struct {
	Accounts []accountEntry `json:"accounts"`
}
type accountEntry struct {
	Name              string `json:"name"`
	PrivateKey        string `json:"privateKey"`
	Address           string `json:"address"`
	ReceivingAddress  string `json:"receivingAddress"`
	ViewingPrivateKey string `json:"viewingPrivateKey"`
}

// Account separates public receiving identity from explicitly loaded local signing/viewing keys.
// Loading a recipient never derives a private Invisi wallet.
type Account struct {
	Name         string
	Key          *ecdsa.PrivateKey
	PublicWallet *sdk.PublicWallet
	ViewingKey   []byte
}

// GetAccounts reads only the explicitly selected local file. It never creates or funds accounts.
func GetAccounts(path string) ([]Account, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load accounts (run go run ./examples/scripts/init-accounts first): %w", err)
	}
	defer clear(data)
	var file accountFile
	if err = json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if len(file.Accounts) < 2 {
		return nil, fmt.Errorf("at least two accounts required")
	}
	accounts := make([]Account, 0, len(file.Accounts))
	seen := map[string]bool{}
	for _, entry := range file.Accounts {
		key, err := crypto.HexToECDSA(strings.TrimPrefix(entry.PrivateKey, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid private key for %s", entry.Name)
		}
		wallet, err := sdk.PublicWalletFromAddress(entry.ReceivingAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid receiving address for %s: %w", entry.Name, err)
		}
		if !common.IsHexAddress(entry.Address) || wallet.OwnerAddress() != common.HexToAddress(entry.Address) || wallet.OwnerAddress() != crypto.PubkeyToAddress(key.PublicKey) {
			return nil, fmt.Errorf("account identity mismatch for %s", entry.Name)
		}
		viewingKey, err := hex.DecodeString(strings.TrimPrefix(entry.ViewingPrivateKey, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid viewing key for %s", entry.Name)
		}
		if _, err = sdk.NewViewingCapability(wallet.OwnerAddress(), viewingKey, wallet.ViewingPublicKey()); err != nil {
			return nil, fmt.Errorf("viewing key mismatch for %s", entry.Name)
		}
		address := wallet.OwnerAddress().Hex()
		if seen[address] {
			return nil, fmt.Errorf("duplicate account %s", entry.Name)
		}
		seen[address] = true
		accounts = append(accounts, Account{entry.Name, key, wallet, viewingKey})
	}
	return accounts, nil
}

// InitializeAccounts creates ten fresh local keys with exclusive creation and owner-only permissions.
// Existing files are never replaced. No chain or funded account file is consulted.
func InitializeAccounts(path string) error {
	file := accountFile{}
	for _, name := range []string{"Alice", "Bob", "Charlie", "Dave", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Judy"} {
		key, err := crypto.GenerateKey()
		if err != nil {
			return err
		}
		privateKey := "0x" + hex.EncodeToString(crypto.FromECDSA(key))
		wallet, err := sdk.WalletFromPrivateKey(privateKey)
		if err != nil {
			return err
		}
		address, err := wallet.GetAddress()
		if err != nil {
			return err
		}
		viewer, err := wallet.GetViewer()
		if err != nil {
			return err
		}
		viewingKey := viewer.ExportKey()
		file.Accounts = append(file.Accounts, accountEntry{Name: name, PrivateKey: privateKey, Address: wallet.OwnerAddress().Hex(), ReceivingAddress: address, ViewingPrivateKey: "0x" + hex.EncodeToString(viewingKey)})
		clear(viewingKey)
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	defer clear(data)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("refusing to replace an existing account file: %w", err)
	}
	if _, err = f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
