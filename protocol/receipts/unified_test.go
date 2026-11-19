package receipts

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
)

func TestCommittedRealReceipts(t *testing.T) {
	raw, e := os.ReadFile("testdata/unified-receipts.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Cases []struct {
			Name    string
			Context struct {
				ProtocolAddress common.Address
				ChainID         string
			}
			Receipt  struct{ Logs []types.Log }
			Expected struct{ OrdinaryNotes, CreditStates, Nullifiers int }
		}
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			chain, _ := new(big.Int).SetString(c.Context.ChainID, 10)
			d, _ := NewDecoder(protocol.TransactionContext{ChainID: chain, ProtocolAddress: c.Context.ProtocolAddress})
			logs := []types.Log{}
			for _, l := range c.Receipt.Logs {
				if l.Address == c.Context.ProtocolAddress {
					logs = append(logs, l)
				}
			}
			b, e := d.Decode(Block{}, logs)
			if e != nil {
				t.Fatal(e)
			}
			n := len(b.Nullifiers) + len(b.CreditNullifiers)
			if len(b.Outputs) != c.Expected.OrdinaryNotes || len(b.CreditStates) != c.Expected.CreditStates || n != c.Expected.Nullifiers {
				t.Fatalf("notes %d credits %d nullifiers %d", len(b.Outputs), len(b.CreditStates), n)
			}
		})
	}
}

func TestUnifiedReceiptRejectsBrokenCorrelation(t *testing.T) {
	raw, e := os.ReadFile("testdata/unified-receipts.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Cases []struct {
			Context struct {
				ProtocolAddress common.Address
				ChainID         string
			}
			Receipt struct{ Logs []types.Log }
		}
	}
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	c := f.Cases[0]
	chain, _ := new(big.Int).SetString(c.Context.ChainID, 10)
	d, _ := NewDecoder(protocol.TransactionContext{ChainID: chain, ProtocolAddress: c.Context.ProtocolAddress})
	logs := []types.Log{}
	for _, l := range c.Receipt.Logs {
		if l.Address == c.Context.ProtocolAddress {
			logs = append(logs, l)
		}
	}
	for _, name := range []string{"duplicate", "missing-operation", "changed-Core", "changed-output-id", "wrong-chain"} {
		t.Run(name, func(t *testing.T) {
			copyLogs := append([]types.Log(nil), logs...)
			decoder := d
			switch name {
			case "duplicate":
				copyLogs = append(copyLogs, copyLogs[0])
			case "missing-operation":
				copyLogs = copyLogs[1:]
			case "changed-Core":
				copyLogs[0].Address = common.HexToAddress("1")
			case "changed-output-id":
				for i, l := range copyLogs {
					event, _ := d.abi.EventByID(l.Topics[0])
					if event.Name == "NoteOutputs" {
						copyLogs[i].Topics = append([]common.Hash(nil), l.Topics...)
						copyLogs[i].Topics[1] = common.HexToHash("999")
					}
				}
			case "wrong-chain":
				decoder, _ = NewDecoder(protocol.TransactionContext{ChainID: big.NewInt(1), ProtocolAddress: c.Context.ProtocolAddress})
			}
			if _, e := decoder.Decode(Block{}, copyLogs); e == nil {
				t.Fatal("accepted broken receipt correlation")
			}
		})
	}
}
