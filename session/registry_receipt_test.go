package session

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestRegistrationReconciliationUsesItsOwnCanonicalEvent(t *testing.T) {
	// A later update in the same block may replace end-of-block keys. Its state
	// cannot invalidate the exact ReceivingKeysSet event in this receipt.
	tracked, _ := registeredOperationFixture(t)
	confirmed, err := tracked.Reconcile(context.Background())
	if err != nil || confirmed.Status() != "confirmed" {
		t.Fatal("canonical receiving-key event was not confirmed", err)
	}
}

func TestRegistrationReconciliationRejectsMisboundEvents(t *testing.T) {
	for _, change := range []string{"owner", "keys", "nonce", "block", "transaction", "address", "removed", "missing", "duplicate"} {
		t.Run(change, func(t *testing.T) {
			tracked, provider := registeredOperationFixture(t)
			log := provider.receipt.Logs[0]
			switch change {
			case "owner":
				log.Topics[1] = common.HexToHash("0x1")
			case "keys":
				log.Data[0] ^= 1
			case "nonce":
				log.Data[95] ^= 1
			case "block":
				log.BlockHash = common.HexToHash("0x1")
			case "transaction":
				log.TxHash = common.HexToHash("0x1")
			case "address":
				log.Address = common.HexToAddress("0x9999")
			case "removed":
				log.Removed = true
			case "missing":
				provider.receipt.Logs = nil
			case "duplicate":
				provider.receipt.Logs = append(provider.receipt.Logs, log)
			}
			if _, err := tracked.Reconcile(context.Background()); err == nil {
				t.Fatal("misbound event confirmed")
			}
		})
	}
}

func TestRegistrationConfirmationReturnsToUncertainAfterReorg(t *testing.T) {
	for _, change := range []string{"missing", "notfound", "replaced block", "during read"} {
		t.Run(change, func(t *testing.T) {
			tracked, provider := registeredOperationFixture(t)
			if change == "during read" {
				provider.afterRead = func() { provider.header = types.CopyHeader(provider.header); provider.header.Extra = []byte{1} }
			} else {
				confirmed, err := tracked.Reconcile(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				tracked = confirmed
				switch change {
				case "missing":
					provider.receipt = nil
				case "notfound":
					provider.receipt = nil
					provider.receiptErr = ethereum.NotFound
				case "replaced block":
					provider.header = types.CopyHeader(provider.header)
					provider.header.Extra = []byte{1}
				}
			}
			current, err := tracked.Reconcile(context.Background())
			if err != nil || current.Status() != "uncertain" {
				t.Fatal("orphaned receipt remained confirmed", err)
			}
		})
	}
}
