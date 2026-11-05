package transaction

import (
	"errors"
	"math/big"

	"github.com/invisi-lab/invisi-sdk-go/core"
)

func (o *Execution) CreditPrivateInputs(private map[string]any) (map[string]any, error) {
	if o.Plan.Kind == "cancel-credit" {
		note := o.Plan.CreditInputs[0].State.Note
		m := o.Plan.CreditMemberships[0]
		private["spender_payout"] = note.SpenderPayout.String()
		private["spender_wallet_identity"] = note.SpenderWalletIdentity.String()
		private["owner_wallet_identity"] = note.OwnerWalletIdentity.String()
		private["input_nonce"] = note.Nonce.String()
		private["expiry"] = note.Expiry.String()
		private["credit_shared_key"] = note.CreditSharedKey.String()
		AddCreditStateWitness(private, note.CreditSpendingState)
		private["leaf_index"] = new(big.Int).SetUint64(uint64(m.LeafIndex)).String()
		path := []string{}
		for _, h := range m.Path {
			path = append(path, h.Big().String())
		}
		private["merkle_path"] = path
		return private, nil
	}
	capacity := o.Plan.Capacity
	amounts, nonces, indices, sources := []string{}, []string{}, []string{}, []string{}
	paths := [][]string{}
	if o.Plan.Kind == "credit-draw" {
		identities, payouts, keys, draws := []string{}, []string{}, []string{}, []string{}
		for i, input := range o.Plan.CreditInputs {
			note := input.State.Note
			amounts = append(amounts, note.Amount.String())
			nonces = append(nonces, note.Nonce.String())
			indices = append(indices, new(big.Int).SetUint64(uint64(input.LeafIndex)).String())
			path := []string{}
			for _, h := range o.Plan.CreditMemberships[i].Path {
				path = append(path, h.Big().String())
			}
			paths = append(paths, path)
			idx := -1
			for j, source := range o.Plan.DrawCredit.Sources {
				if source.TreeId.Uint64() == input.TreeID {
					idx = j
				}
			}
			if idx < 0 {
				return nil, errors.New("missing Credit source")
			}
			sources = append(sources, big.NewInt(int64(idx)).String())
			identities = append(identities, note.OwnerWalletIdentity.String())
			payouts = append(payouts, note.OwnerPayout.String())
			keys = append(keys, note.CreditSharedKey.String())
			draws = append(draws, o.Plan.DrawAmounts[i].String())
		}
		private["spender_wallet_identity"] = o.Plan.CreditInputs[0].State.Note.SpenderWalletIdentity.String()
		private["owner_payouts"] = PadStrings(payouts, capacity)
		private["owner_wallet_identities"] = PadStrings(identities, capacity)
		private["credit_shared_keys"] = PadStrings(keys, capacity)
		private["draw_amounts"] = PadStrings(draws, capacity)
		for column, name := range []string{"single_limits", "daily_limits", "monthly_limits", "last_days", "daily_spents", "last_months", "monthly_spents"} {
			values := []string{}
			for _, input := range o.Plan.CreditInputs {
				values = append(values, input.State.Note.CreditSpendingState.Fields()[column].String())
			}
			private[name] = PadStrings(values, capacity)
		}
	} else {
		for i, input := range o.Plan.Inputs {
			amounts = append(amounts, input.Note.Amount.String())
			nonces = append(nonces, input.Note.Nonce.String())
			indices = append(indices, new(big.Int).SetUint64(uint64(input.LeafIndex)).String())
			path := []string{}
			for _, h := range o.Plan.Memberships[i].Path {
				path = append(path, h.Big().String())
			}
			paths = append(paths, path)
			idx := -1
			for j, source := range o.Plan.CreateCredit.Sources {
				if source.TreeId.Uint64() == input.TreeID {
					idx = j
				}
			}
			if idx < 0 {
				return nil, errors.New("missing funds source")
			}
			sources = append(sources, big.NewInt(int64(idx)).String())
		}
		note := o.Plan.CreditOutput.Note
		private["spender_payout"] = note.SpenderPayout.String()
		private["spender_wallet_identity"] = note.SpenderWalletIdentity.String()
		private["owner_wallet_identity"] = note.OwnerWalletIdentity.String()
		private["credit_amount"] = note.Amount.String()
		private["credit_nonce"] = note.Nonce.String()
		private["credit_shared_key"] = note.CreditSharedKey.String()
		if o.Plan.Kind == "create-credit" {
			private["single_limit"] = note.SingleLimit.String()
			private["daily_limit"] = note.DailyLimit.String()
			private["monthly_limit"] = note.MonthlyLimit.String()
		} else {
			AddCreditStateWitness(private, note.CreditSpendingState)
		}
		private["change_amount"] = "0"
		private["change_nonce"] = "0"
		if len(o.Plan.Outputs) > 0 {
			private["change_amount"] = o.Plan.Outputs[0].Note.Amount.String()
			private["change_nonce"] = o.Plan.Outputs[0].Note.Nonce.String()
		}
		if o.Plan.Kind == "top-up-credit" {
			previous := o.Plan.CreditInputs[0]
			private["previous_single_limit"] = previous.State.Note.SingleLimit.String()
			private["previous_daily_limit"] = previous.State.Note.DailyLimit.String()
			private["previous_monthly_limit"] = previous.State.Note.MonthlyLimit.String()
			private["previous_credit_amount"] = previous.State.Note.Amount.String()
			private["previous_credit_nonce"] = previous.State.Note.Nonce.String()
			private["previous_credit_leaf_index"] = new(big.Int).SetUint64(uint64(previous.LeafIndex)).String()
			path := []string{}
			for _, h := range o.Plan.CreditMemberships[0].Path {
				path = append(path, h.Big().String())
			}
			private["previous_credit_merkle_path"] = path
		}
	}
	for len(paths) < capacity {
		paths = append(paths, ZeroPath())
	}
	private["input_amounts"] = PadStrings(amounts, capacity)
	private["input_nonces"] = PadStrings(nonces, capacity)
	private["leaf_indices"] = PadStrings(indices, capacity)
	private["merkle_paths"] = paths
	private["input_source_indices"] = PadStrings(sources, capacity)
	return private, nil
}

func AddCreditStateWitness(private map[string]any, state core.CreditSpendingState) {
	for i, name := range []string{"single_limit", "daily_limit", "monthly_limit", "last_day", "daily_spent", "last_month", "monthly_spent"} {
		private[name] = state.Fields()[i].String()
	}
}
