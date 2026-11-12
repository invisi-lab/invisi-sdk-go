package shared

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
	"time"
)

type finalityFixture struct {
	header  *types.Header
	receipt *types.Receipt
}

func (f finalityFixture) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	return f.header, nil
}
func (f finalityFixture) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	if f.receipt == nil {
		return nil, ethereum.NotFound
	}
	return f.receipt, nil
}
func TestNativeFinalityWaitHasFiniteDeadline(t *testing.T) {
	_, err := WaitNativeFinality(context.Background(), finalityFixture{}, common.HexToHash("0x1"), time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
func TestNativeFinalityWaitRequiresSuccessfulOriginalReceipt(t *testing.T) {
	header := &types.Header{Number: big.NewInt(1), Time: 1}
	hash := common.HexToHash("0x1")
	receipt := &types.Receipt{TxHash: hash, BlockHash: header.Hash(), BlockNumber: big.NewInt(1), Status: 1}
	result, err := WaitNativeFinality(context.Background(), finalityFixture{header, receipt}, hash, time.Second)
	if err != nil || result.Level != "finalized" {
		t.Fatal(result, err)
	}
	receipt.Status = 0
	if _, err = WaitNativeFinality(context.Background(), finalityFixture{header, receipt}, hash, time.Second); err == nil {
		t.Fatal("reverted payment treated as successful")
	}
}
