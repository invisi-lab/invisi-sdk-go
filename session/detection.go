package session

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/wallet"
)

func decodeClue(v string) []byte { b, _ := wallet.Bytes(v, -1); return b }
func (s *WalletSession) pageResolver(ctx context.Context, d wallet.DeliveryDecryptor, page PublicPage) (walletResolver, error) {
	base := s.deliveryResolver(ctx, d)
	detector, ok := d.(wallet.BatchDeliveryDetector)
	if !ok {
		return base, nil
	}
	items := []wallet.DeliveryDetectionItem{}
	seen := map[string]bool{}
	add := func(m wallet.Method, ciphertext, clue []byte, token common.Address, commitment common.Hash, role string) {
		p := deliveryRequest(ciphertext, s.coreContext(), token.Big().String(), commitment, role)
		p.Context.Owner = s.wallet.OwnerAddress().Hex()
		p.Context.Clue = wallet.Hex(clue)
		id := walletFingerprint(m, p)
		if seen[id] {
			return
		}
		seen[id] = true
		items = append(items, wallet.DeliveryDetectionItem{ID: id, Owner: s.wallet.OwnerAddress(), ChainID: s.scope.ChainID, ProtocolAddress: s.scope.ProtocolAddress, TokenAddress: token, OutputCommitment: commitment, Role: role, Clue: append([]byte(nil), clue...), Ciphertext: append([]byte(nil), ciphertext...)})
	}
	for _, tx := range page.Transactions {
		for _, o := range tx.Batch.Outputs {
			if len(o.Ciphertext) > 0 {
				add(wallet.DecryptNote, o.Ciphertext, o.Clue, o.Token, o.Commitment, "")
			}
		}
		for _, o := range tx.Batch.CreditStates {
			for _, envelope := range o.Deliveries {
				for _, role := range []string{"owner", "spender"} {
					add(wallet.DecryptCredit, envelope.Ciphertext, envelope.Clue, o.Token, o.Commitment, role)
				}
			}
		}
	}
	if len(items) == 0 {
		return base, nil
	}
	result, e := detector.Detect(ctx, items)
	if e != nil {
		return nil, e
	}
	if result.Owner != s.wallet.OwnerAddress() || len(result.Results) != len(items) {
		return nil, errors.New("native detection account or result count mismatch")
	}
	decisions := map[string]string{}
	for _, r := range result.Results {
		if !seen[r.ID] || decisions[r.ID] != "" {
			return nil, errors.New("invalid native detection result")
		}
		if r.Status != "candidate" && r.Status != "not-recipient" && r.Status != "unresolved" {
			return nil, errors.New("invalid native detection status")
		}
		decisions[r.ID] = r.Status
	}
	return func(m wallet.Method, p wallet.DecryptParams) (any, error) {
		switch decisions[walletFingerprint(m, p)] {
		case "not-recipient":
			return nil, wallet.ErrNotRecipient
		case "unresolved":
			return nil, wallet.ErrAuthenticationFailed
		case "candidate":
			return base(m, p)
		default:
			return nil, errors.New("missing native detection result")
		}
	}, nil
}
