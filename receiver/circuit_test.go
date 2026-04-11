package receiver

import (
	"testing"
)

func TestSelectBatchCircuit(t *testing.T) {
	tests := []struct {
		count   int
		name    string
		maxSize int
	}{
		{1, "batch_settle_1", 1},
		{2, "batch_settle_3", 3},
		{3, "batch_settle_3", 3},
		{4, "batch_settle_10", 10},
		{10, "batch_settle_10", 10},
		{11, "batch_settle_100", 100},
		{50, "batch_settle_100", 100},
		{100, "batch_settle_100", 100},
	}

	for _, tt := range tests {
		info, err := SelectBatchCircuit(tt.count)
		if err != nil {
			t.Errorf("SelectBatchCircuit(%d) error: %v", tt.count, err)
			continue
		}
		if info.Name != tt.name {
			t.Errorf("SelectBatchCircuit(%d).Name = %s, want %s", tt.count, info.Name, tt.name)
		}
		if info.MaxSize != tt.maxSize {
			t.Errorf("SelectBatchCircuit(%d).MaxSize = %d, want %d", tt.count, info.MaxSize, tt.maxSize)
		}
	}
}

func TestSelectBatchCircuit_InvalidCount(t *testing.T) {
	_, err := SelectBatchCircuit(0)
	if err == nil {
		t.Error("SelectBatchCircuit(0) should return error")
	}

	_, err = SelectBatchCircuit(-1)
	if err == nil {
		t.Error("SelectBatchCircuit(-1) should return error")
	}

	_, err = SelectBatchCircuit(101)
	if err == nil {
		t.Error("SelectBatchCircuit(101) should return error")
	}
}

func TestSelectCreditDrawCircuit(t *testing.T) {
	tests := []struct {
		count   int
		name    string
		maxSize int
	}{
		{1, "credit_draw_1", 1},
		{2, "credit_draw_3", 3},
		{3, "credit_draw_3", 3},
		{4, "credit_draw_10", 10},
		{10, "credit_draw_10", 10},
		{11, "credit_draw_100", 100},
		{50, "credit_draw_100", 100},
		{100, "credit_draw_100", 100},
	}

	for _, tt := range tests {
		info, err := SelectCreditDrawCircuit(tt.count)
		if err != nil {
			t.Errorf("SelectCreditDrawCircuit(%d) error: %v", tt.count, err)
			continue
		}
		if info.Name != tt.name {
			t.Errorf("SelectCreditDrawCircuit(%d).Name = %s, want %s", tt.count, info.Name, tt.name)
		}
		if info.MaxSize != tt.maxSize {
			t.Errorf("SelectCreditDrawCircuit(%d).MaxSize = %d, want %d", tt.count, info.MaxSize, tt.maxSize)
		}
	}
}

func TestSelectCreditDrawCircuit_InvalidCount(t *testing.T) {
	_, err := SelectCreditDrawCircuit(0)
	if err == nil {
		t.Error("SelectCreditDrawCircuit(0) should return error")
	}

	_, err = SelectCreditDrawCircuit(-1)
	if err == nil {
		t.Error("SelectCreditDrawCircuit(-1) should return error")
	}

	_, err = SelectCreditDrawCircuit(101)
	if err == nil {
		t.Error("SelectCreditDrawCircuit(101) should return error")
	}
}
