package receiver

import "fmt"

// SelectBatchCircuit returns the smallest batch settle circuit that fits
// the given payment count. Available circuits: 1, 3, 10, 100.
//
// Count must be in range [1, 100].
func SelectBatchCircuit(count int) (BatchCircuitInfo, error) {
	if count < 1 || count > 100 {
		return BatchCircuitInfo{}, fmt.Errorf("receiver: invalid batch size: %d (must be 1..100)", count)
	}
	if count == 1 {
		return BatchCircuitInfo{Name: "batch_settle_1", MaxSize: 1}, nil
	}
	if count <= 3 {
		return BatchCircuitInfo{Name: "batch_settle_3", MaxSize: 3}, nil
	}
	if count <= 10 {
		return BatchCircuitInfo{Name: "batch_settle_10", MaxSize: 10}, nil
	}
	return BatchCircuitInfo{Name: "batch_settle_100", MaxSize: 100}, nil
}

// SelectCreditDrawCircuit returns the smallest credit draw circuit that fits
// the given credit count. Available circuits: 1, 3, 10, 100.
//
// Count must be in range [1, 100].
func SelectCreditDrawCircuit(count int) (BatchCircuitInfo, error) {
	if count < 1 || count > 100 {
		return BatchCircuitInfo{}, fmt.Errorf("receiver: invalid credit draw batch size: %d (must be 1..100)", count)
	}
	if count == 1 {
		return BatchCircuitInfo{Name: "credit_draw_1", MaxSize: 1}, nil
	}
	if count <= 3 {
		return BatchCircuitInfo{Name: "credit_draw_3", MaxSize: 3}, nil
	}
	if count <= 10 {
		return BatchCircuitInfo{Name: "credit_draw_10", MaxSize: 10}, nil
	}
	return BatchCircuitInfo{Name: "credit_draw_100", MaxSize: 100}, nil
}
