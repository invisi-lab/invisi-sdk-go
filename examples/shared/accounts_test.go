package shared

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeAccountsRefusesOverwrite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "accounts.json")
	if err := InitializeAccounts(p); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(p)
	if err := InitializeAccounts(p); err == nil {
		t.Fatal("account file overwritten")
	}
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Fatal("existing account changed")
	}
	var stored accountFile
	if err := json.Unmarshal(before, &stored); err != nil {
		t.Fatal(err)
	}
	for i, name := range []string{"Alice", "Bob", "Charlie", "Dave", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Judy"} {
		if i >= len(stored.Accounts) || stored.Accounts[i].Name != name {
			t.Fatal("missing expected account", name)
		}
		entry := stored.Accounts[i]
		if entry.Address == "" || entry.ReceivingAddress == "" || entry.ViewingPrivateKey == "" {
			t.Fatal("missing saved public/viewing material", name)
		}
	}
	info, err := os.Stat(p)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("account file permissions", info, err)
	}
	accounts, err := GetAccounts(p)
	if err != nil || len(accounts) != 10 {
		t.Fatal(len(accounts), err)
	}
}

func TestGetAccountsRejectsMismatchedSavedIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	if err := InitializeAccounts(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file accountFile
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	file.Accounts[0].ReceivingAddress = file.Accounts[1].ReceivingAddress
	data, err = json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = GetAccounts(path); err == nil {
		t.Fatal("mismatched recipient identity accepted")
	}
}
