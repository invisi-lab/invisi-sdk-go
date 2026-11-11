package credit

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/rpc/broadcaster"
)

type WithBroadcasterOptions struct {
	MaxFeeAmount *big.Int
	MaxAttempts  int
}
type BroadcasterBinding struct {
	Configuration broadcaster.Configuration
	Quote         broadcaster.SignedQuote
	Intent        []byte
}
type Submission struct {
	RequestID, State string
	TransactionHash  string
}

func unsignedIntent(i Intent) ([]byte, error) {
	if _, _, err := i.statement(); err != nil {
		return nil, err
	}
	switch {
	case i.Create != nil:
		p := *i.Create
		p.FeeAmount, _ = normalizeFee(p.FeeRecipient, p.FeeAmount)
		return pack("createCredit", []byte{}, p, i.Deliveries)
	case i.Update != nil:
		p := *i.Update
		p.FeeAmount, _ = normalizeFee(p.FeeRecipient, p.FeeAmount)
		return pack("updateCredit", []byte{}, p, i.Deliveries)
	case i.Draw != nil:
		p := *i.Draw
		p.FeeAmount, _ = normalizeFee(p.FeeRecipient, p.FeeAmount)
		return pack("creditDraw", []byte{}, p, i.Deliveries)
	case i.Revoke != nil:
		p := *i.Revoke
		fee, _ := normalizeFee(p.FeeRecipient, p.FeeAmount)
		return pack("revokeCredit", p.Token, p.Owner, p.Shard, p.Slot, p.Deadline, p.FeeRecipient, fee, []byte{})
	}
	return nil, fmt.Errorf("Credit operation required")
}

// ExportBroadcasterIntent includes only public request fields and opaque deliveries; proofs and owner signatures are empty.
func (p *Prepared) ExportBroadcasterIntent() ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return unsignedIntent(p.intent)
}
func intentFee(i Intent) (common.Address, *big.Int) {
	switch {
	case i.Create != nil:
		return i.Create.FeeRecipient, i.Create.FeeAmount
	case i.Update != nil:
		return i.Update.FeeRecipient, i.Update.FeeAmount
	case i.Draw != nil:
		return i.Draw.FeeRecipient, i.Draw.FeeAmount
	case i.Revoke != nil:
		return i.Revoke.FeeRecipient, i.Revoke.FeeAmount
	}
	return common.Address{}, nil
}
func reprice(i Intent, recipient common.Address, amount *big.Int) (Intent, error) {
	next, err := freeze(i)
	if err != nil {
		return Intent{}, err
	}
	fee, err := normalizeFee(recipient, amount)
	if err != nil || fee.Sign() == 0 {
		return Intent{}, fmt.Errorf("broadcaster fee must be positive uint128")
	}
	switch {
	case next.Create != nil:
		next.Create.FeeRecipient = recipient
		next.Create.FeeAmount = fee
	case next.Update != nil:
		next.Update.FeeRecipient = recipient
		next.Update.FeeAmount = fee
	case next.Revoke != nil:
		next.Revoke.FeeRecipient = recipient
		next.Revoke.FeeAmount = fee
	case next.Draw != nil:
		if fee.Cmp(next.Draw.TotalAmount) >= 0 {
			return Intent{}, fmt.Errorf("broadcaster fee consumes ordinary Note")
		}
		next.Draw.FeeRecipient = recipient
		next.Draw.FeeAmount = fee
		net := new(big.Int).Sub(next.Draw.TotalAmount, fee)
		commitment, err := primitives.ComputeNoteCommitment(next.Context.ChainID, next.Context.ProtocolAddress, next.OutputIdentity, primitives.BigIntToElement(next.Draw.Token.Big()), primitives.BigIntToElement(net))
		if err != nil {
			return Intent{}, err
		}
		next.Draw.NoteCommitment = common.Hash(commitment.Bytes())
		last := len(next.Deliveries) - 1
		if last < 0 {
			return Intent{}, fmt.Errorf("ordinary Note delivery slot missing")
		}
		if next.OutputNote != nil {
			identity, err := primitives.ComputeNoteIdentity(next.OutputNote.OwnerAddress, next.OutputNote.Nonce)
			if err != nil || identity != next.OutputIdentity {
				return Intent{}, fmt.Errorf("frozen ordinary Note identity mismatch")
			}
			next.OutputNote.Amount = primitives.BigIntToElement(net)
			if len(next.Deliveries[last].Ciphertext) > 0 && next.NoteRecipientKeys == nil {
				return Intent{}, fmt.Errorf("fee quotation requires frozen Note recipient keys")
			}
			delivery, err := noteDelivery(*next.OutputNote, next.NoteRecipientKeys)
			if err != nil {
				return Intent{}, err
			}
			next.Deliveries[last] = delivery
		} else if len(next.Deliveries[last].Ciphertext) > 0 || len(next.Deliveries[last].Clue) > 0 {
			return Intent{}, fmt.Errorf("enveloped draw quotation requires PrepareDrawToNote")
		}
	default:
		return Intent{}, fmt.Errorf("Credit operation required")
	}
	if _, _, err = next.statement(); err != nil {
		return Intent{}, err
	}
	return next, nil
}

// WithBroadcaster freezes the quoted fee before authorization or proving; grant debits remain gross.
func (p *Prepared) WithBroadcaster(ctx context.Context, client *broadcaster.Client, options WithBroadcasterOptions) error {
	if client == nil || options.MaxFeeAmount == nil || options.MaxFeeAmount.Sign() <= 0 || options.MaxFeeAmount.BitLen() > 128 {
		return fmt.Errorf("broadcaster client and positive uint128 fee limit required")
	}
	limit := new(big.Int).Set(options.MaxFeeAmount)
	attempts := options.MaxAttempts
	if attempts == 0 {
		attempts = 5
	}
	if attempts < 1 || attempts > 10 {
		return fmt.Errorf("fee attempts must be 1..10")
	}
	p.mu.Lock()
	if p.authority != nil || p.proof != nil || len(p.revokeSignature) > 0 || p.intent.Broadcaster != nil {
		p.mu.Unlock()
		return fmt.Errorf("select broadcaster before authorization and proving")
	}
	base, err := freeze(p.intent)
	p.mu.Unlock()
	if err != nil {
		return err
	}
	proposal := base
	for attempt := 0; attempt < attempts; attempt++ {
		data, err := unsignedIntent(proposal)
		if err != nil {
			return err
		}
		estimated, err := client.EstimateQuote(ctx, data)
		if err != nil {
			return err
		}
		amount, ok := new(big.Int).SetString(estimated.FeeAmount, 10)
		if !ok || amount.Sign() <= 0 || amount.BitLen() > 128 || amount.Cmp(limit) > 0 {
			return fmt.Errorf("broadcaster fee exceeds the approved limit")
		}
		if estimated.FeeRecipient != client.Configuration().FeeRecipient {
			return fmt.Errorf("broadcaster fee recipient differs from registration")
		}
		oldRecipient, oldAmount := intentFee(proposal)
		oldAmount, err = normalizeFee(oldRecipient, oldAmount)
		if err != nil {
			return err
		}
		if oldRecipient != estimated.FeeRecipient || oldAmount.Cmp(amount) != 0 {
			proposal, err = reprice(base, estimated.FeeRecipient, amount)
			if err != nil {
				return err
			}
			continue
		}
		trust := broadcaster.Trust{ChainID: primitives.ElementToBigInt(proposal.Context.ChainID), Protocol: common.BigToAddress(primitives.ElementToBigInt(proposal.Context.ProtocolAddress)), Signer: client.Configuration().QuoteSigner}
		signed, err := client.GetQuote(ctx, data, trust)
		if err != nil {
			return err
		}
		if err = broadcaster.VerifyQuote(signed, trust, data, uint64(time.Now().Unix())); err != nil {
			return err
		}
		proposal.Broadcaster = &BroadcasterBinding{Configuration: client.Configuration(), Quote: signed, Intent: append([]byte(nil), data...)}
		frozen, err := freeze(proposal)
		if err != nil {
			return err
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.authority != nil || p.proof != nil || len(p.revokeSignature) > 0 || p.intent.Broadcaster != nil {
			return fmt.Errorf("Credit preparation advanced during quotation")
		}
		p.intent = frozen
		return nil
	}
	return fmt.Errorf("broadcaster fee did not converge")
}
func (p *Prepared) ExportBroadcasterQuote() (broadcaster.SignedQuote, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.intent.Broadcaster == nil {
		return broadcaster.SignedQuote{}, fmt.Errorf("broadcaster quote unavailable")
	}
	binding, err := freeze(*p.intent.Broadcaster)
	if err != nil {
		return broadcaster.SignedQuote{}, err
	}
	binding.Quote.Context = &broadcaster.Trust{ChainID: primitives.ElementToBigInt(p.intent.Context.ChainID), Protocol: common.BigToAddress(primitives.ElementToBigInt(p.intent.Context.ProtocolAddress)), Signer: binding.Configuration.QuoteSigner}
	return binding.Quote, nil
}

// ExportBroadcasterSubmission validates a final request before a managed journal marks it externalized.
func (p *Prepared) ExportBroadcasterSubmission(client *broadcaster.Client) ([]byte, broadcaster.SignedQuote, error) {
	if client == nil {
		return nil, broadcaster.SignedQuote{}, fmt.Errorf("broadcaster required")
	}
	p.mu.Lock()
	if p.intent.Broadcaster == nil {
		p.mu.Unlock()
		return nil, broadcaster.SignedQuote{}, fmt.Errorf("broadcaster quote required")
	}
	binding, err := freeze(*p.intent.Broadcaster)
	if err != nil {
		p.mu.Unlock()
		return nil, broadcaster.SignedQuote{}, err
	}
	if p.intent.Revoke != nil && len(p.revokeSignature) == 0 {
		p.mu.Unlock()
		return nil, broadcaster.SignedQuote{}, fmt.Errorf("broadcaster revocation requires owner signature")
	}
	p.mu.Unlock()
	if binding.Configuration != client.Configuration() {
		return nil, broadcaster.SignedQuote{}, fmt.Errorf("use the originally selected broadcaster")
	}
	data, err := p.Calldata(common.Address{})
	if err != nil {
		return nil, broadcaster.SignedQuote{}, err
	}
	expected, err := broadcaster.RequestHash(binding.Intent)
	if err != nil {
		return nil, broadcaster.SignedQuote{}, err
	}
	actual, err := broadcaster.RequestHash(data)
	if err != nil || actual != expected {
		return nil, broadcaster.SignedQuote{}, fmt.Errorf("final Credit calldata differs from quoted intent")
	}
	quote, err := p.ExportBroadcasterQuote()
	if err != nil {
		return nil, broadcaster.SignedQuote{}, err
	}
	if err = broadcaster.VerifyQuote(quote, *quote.Context, data, uint64(time.Now().Unix())); err != nil {
		return nil, broadcaster.SignedQuote{}, err
	}
	return data, quote, nil
}
func (p *Prepared) SubmitBroadcaster(ctx context.Context, client *broadcaster.Client) (broadcaster.Status, error) {
	p.mu.Lock()
	known := p.submission
	binding := p.intent.Broadcaster
	p.mu.Unlock()
	if client == nil || binding == nil || binding.Configuration != client.Configuration() {
		return broadcaster.Status{}, fmt.Errorf("use the originally selected broadcaster")
	}
	retry := known.State != ""
	if retry && known.RequestID != "" {
		return client.Status(ctx, known.RequestID)
	}
	var data []byte
	var quote broadcaster.SignedQuote
	var err error
	if retry {
		data, err = p.Calldata(common.Address{})
		if err == nil {
			quote, err = p.ExportBroadcasterQuote()
		}
	} else {
		data, quote, err = p.ExportBroadcasterSubmission(client)
	}
	if err != nil {
		return broadcaster.Status{}, err
	}
	if !retry {
		if _, err = client.SubmissionRequirements(ctx, quote); err != nil {
			return broadcaster.Status{}, err
		}
	}
	p.mu.Lock()
	p.submission = Submission{RequestID: known.RequestID, State: "uncertain"}
	p.mu.Unlock()
	status, err := client.Submit(ctx, append([]byte(nil), data...), quote)
	if err != nil {
		return status, err
	}
	if known.RequestID != "" && known.RequestID != status.ID {
		return status, fmt.Errorf("broadcaster changed submission ID")
	}
	p.mu.Lock()
	p.submission = Submission{RequestID: status.ID, State: status.State, TransactionHash: status.TransactionHash}
	p.mu.Unlock()
	return status, nil
}
func (p *Prepared) Submission() Submission { p.mu.Lock(); defer p.mu.Unlock(); return p.submission }
