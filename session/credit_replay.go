package session

import "github.com/ethereum/go-ethereum/common"

// scheduleCreditReplay merges dirty ranges instead of discarding an interrupted
// replay checkpoint whenever discovery sees another Credit event. Its caller
// commits the obligation atomically with the newly authenticated private state.
func scheduleCreditReplay(state *PrivateScanState, from, target uint64) {
	if state.CreditReplayTarget == 0 {
		state.CreditReplayFrom = from
		state.CreditReplayCursor = 0
	} else {
		// An old snapshot has no lower bound: retain its deployment-wide replay.
		if state.CreditReplayFrom != 0 && from < state.CreditReplayFrom {
			state.CreditReplayFrom = from
		}
		if from <= state.CreditReplayCursor {
			state.CreditReplayCursor = 0
		}
	}
	if target > state.CreditReplayTarget {
		state.CreditReplayTarget = target
	}
	state.CreditReplayRevision++
}

func creditReplayRelevant(state WalletState, page PublicPage) bool {
	if len(state.PostpaidCredit.Grants) == 0 {
		return false
	}
	parents := make(map[common.Hash]struct{}, len(state.PostpaidCredit.Grants))
	for _, grant := range state.PostpaidCredit.Grants {
		parents[grant.Nullifier] = struct{}{}
	}
	for _, tx := range page.Transactions {
		for _, nf := range tx.Batch.CreditNullifiers {
			if _, ok := parents[nf]; ok {
				return true
			}
		}
	}
	return false
}
