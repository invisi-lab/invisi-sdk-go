package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/invisi-lab/invisi-sdk-go/session"
	creditsession "github.com/invisi-lab/invisi-sdk-go/session/credit"
)

func (s *WalletStore) RestoreExternalizedOperations(ctx context.Context, scope session.WalletScope) error {
	before := uint64(^uint64(0) >> 1)
	for {
		rows, e := s.db.QueryContext(ctx, "SELECT locator,data,sequence FROM invisi_wallet_records WHERE scope_key=? AND archived=1 AND kind IN ('operations','credit-operations') AND sequence<? ORDER BY sequence DESC LIMIT 128", scope.Key(), before)
		if e != nil {
			return e
		}
		records := []session.WalletRecord{}
		count := 0
		for rows.Next() {
			var locator string
			var raw []byte
			var seq uint64
			if e = rows.Scan(&locator, &raw, &seq); e != nil {
				rows.Close()
				return e
			}
			before = seq
			count++
			plain, e := s.codec.OpenRecord(scope, locator, raw)
			if e != nil {
				rows.Close()
				return e
			}
			var r session.WalletRecord
			e = json.Unmarshal(plain, &r)
			clear(plain)
			if e != nil {
				rows.Close()
				return e
			}
			if r.Sequence != seq || !r.Archived {
				rows.Close()
				return errors.New("invalid archived journal routing")
			}
			r.Archived = false
			switch r.Kind {
			case "operations":
				var o session.OperationRecord
				if e = json.Unmarshal(r.Value, &o); e != nil {
					rows.Close()
					return e
				}
				if !o.Externalized {
					continue
				}
				o.Status = "reorged"
				o.ResolvedAt = nil
				o.WorkerID = ""
				r.Value, e = json.Marshal(o)
			case "credit-operations":
				var o creditsession.Journal
				if e = json.Unmarshal(r.Value, &o); e != nil {
					rows.Close()
					return e
				}
				if !o.Externalized {
					continue
				}
				o.Status = "reorged"
				o.WorkerID = ""
				r.Value, e = json.Marshal(o)
			default:
				rows.Close()
				return errors.New("invalid journal kind")
			}
			if e != nil {
				rows.Close()
				return e
			}
			records = append(records, r)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		if len(records) > 0 {
			if e = s.Update(ctx, scope, func(v *session.WalletState) error {
				v.Invalidated = true
				for _, r := range records {
					if e := session.ApplyWalletRecord(v, r); e != nil {
						return e
					}
				}
				return nil
			}); e != nil {
				return e
			}
		}
		if count < 128 {
			return nil
		}
	}
}
