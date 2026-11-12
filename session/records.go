package session

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
	"sort"
)

type RecordManifestEntry struct {
	Kind     string
	Sequence uint64
	Digest   [32]byte
}
type RecordMetadata struct {
	Version      int
	IndexSecret  []byte
	NextSequence uint64
	Active       map[string]RecordManifestEntry
	Counts       map[string]uint64
}
type WalletRecord struct {
	Key, Kind            string
	Value                json.RawMessage
	Sequence, Generation uint64
	Archived             bool
}
type RecordQuery struct {
	Kind   string
	Limit  int
	Cursor string
}
type RecordPage struct {
	Items      []WalletRecord
	NextCursor string
	Revision   uint64
}
type RecordWalletStore interface {
	WalletStore
	ReadRecords(context.Context, WalletScope, RecordQuery) (RecordPage, error)
	ReadMetadata(context.Context, WalletScope) (WalletState, error)
}

func (c *EncryptedWalletCodec) SealRecord(scope WalletScope, locator string, value []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, e := rand.Read(nonce); e != nil {
		return nil, e
	}
	return c.aead.Seal(nonce, nonce, value, []byte("Invisi.WalletRecord.v1:"+scope.Key()+":"+locator)), nil
}
func (c *EncryptedWalletCodec) OpenRecord(scope WalletScope, locator string, value []byte) ([]byte, error) {
	n := c.aead.NonceSize()
	if len(value) < n+c.aead.Overhead() {
		return nil, errors.New("invalid encrypted wallet record")
	}
	return c.aead.Open(nil, value[:n], value[n:], []byte("Invisi.WalletRecord.v1:"+scope.Key()+":"+locator))
}
func RecordLocator(secret []byte, kind, key string) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(kind + ":" + key))
	return hex.EncodeToString(h.Sum(nil))
}
func WalletRecords(v WalletState) ([]WalletRecord, error) {
	records := []WalletRecord{}
	add := func(kind, key string, val any) error {
		raw, e := json.Marshal(val)
		if e != nil {
			return e
		}
		records = append(records, WalletRecord{Key: key, Kind: kind, Value: raw, Generation: v.Scan.Generation})
		return nil
	}
	for k, n := range v.Notes {
		if e := add("notes", k.Hex(), n); e != nil {
			return nil, e
		}
	}
	for k, c := range v.PostpaidCredit.Grants {
		if e := add("credits", k, c); e != nil {
			return nil, e
		}
	}
	for k, o := range v.Operations {
		if e := add("operations", k, o); e != nil {
			return nil, e
		}
	}
	for k, o := range v.PostpaidCredit.Operations {
		if e := add("credit-operations", k, o); e != nil {
			return nil, e
		}
	}
	for k, u := range v.Scan.Unresolved {
		if e := add("unresolved", k, u); e != nil {
			return nil, e
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Kind != records[j].Kind {
			return records[i].Kind < records[j].Kind
		}
		return records[i].Key < records[j].Key
	})
	return records, nil
}
func ApplyWalletRecord(v *WalletState, r WalletRecord) error {
	switch r.Kind {
	case "notes":
		var n StoredNote
		if e := json.Unmarshal(r.Value, &n); e != nil {
			return e
		}
		v.Notes[common.HexToHash(r.Key)] = n
	case "credits":
		var c creditsession.StoredState
		if e := json.Unmarshal(r.Value, &c); e != nil {
			return e
		}
		v.PostpaidCredit.Grants[r.Key] = c
	case "operations":
		var o OperationRecord
		if e := json.Unmarshal(r.Value, &o); e != nil {
			return e
		}
		v.Operations[r.Key] = o
	case "credit-operations":
		var o creditsession.Journal
		if e := json.Unmarshal(r.Value, &o); e != nil {
			return e
		}
		v.PostpaidCredit.Operations[r.Key] = o
	case "unresolved":
		var u UnresolvedDelivery
		if e := json.Unmarshal(r.Value, &u); e != nil {
			return e
		}
		if v.Scan.Unresolved == nil {
			v.Scan.Unresolved = map[string]UnresolvedDelivery{}
		}
		v.Scan.Unresolved[r.Key] = u
	default:
		return errors.New("unknown wallet record kind")
	}
	return nil
}
func ClearWalletRecords(v *WalletState) {
	v.Notes = map[common.Hash]StoredNote{}
	v.Operations = map[string]OperationRecord{}
	v.PostpaidCredit.Grants = map[string]creditsession.StoredState{}
	v.PostpaidCredit.Operations = map[string]creditsession.Journal{}
	v.Scan.Unresolved = map[string]UnresolvedDelivery{}
}
func IsTerminalRecord(r WalletRecord, pinCredits bool) bool {
	switch r.Kind {
	case "notes":
		var n StoredNote
		if json.Unmarshal(r.Value, &n) != nil {
			return false
		}
		return n.Spent && n.ReservedBy == ""
	case "credits":
		if pinCredits {
			return false
		}
		var c creditsession.StoredState
		if json.Unmarshal(r.Value, &c) != nil {
			return false
		}
		return (c.Consumed || c.Revoked) && c.ReservedBy == ""
	case "operations":
		var o OperationRecord
		if json.Unmarshal(r.Value, &o) != nil {
			return false
		}
		if o.Externalized && (o.Status == "failed" || o.Status == "expired") && o.ResolvedAt == nil {
			return false
		}
		return o.Status == "confirmed" || o.Status == "reverted" || o.Status == "expired" || o.Status == "failed"
	case "credit-operations":
		var o creditsession.Journal
		if json.Unmarshal(r.Value, &o) != nil {
			return false
		}
		return o.Status == "confirmed" || o.Status == "reverted" || o.Status == "failed"
	}
	return false
}
func PrepareRecordMetadata(v *WalletState) error {
	if v.Records.Version == 0 {
		v.Records.Version = 1
		v.Records.IndexSecret = make([]byte, 32)
		if _, e := rand.Read(v.Records.IndexSecret); e != nil {
			return e
		}
		v.Records.Active = map[string]RecordManifestEntry{}
		v.Records.Counts = map[string]uint64{}
	}
	if v.Records.Version != 1 || len(v.Records.IndexSecret) != 32 || v.Records.Active == nil || v.Records.Counts == nil {
		return errors.New("invalid wallet record metadata")
	}
	return nil
}

type JournalRestorer interface {
	RestoreExternalizedOperations(context.Context, WalletScope) error
}
