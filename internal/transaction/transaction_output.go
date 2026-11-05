package transaction

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

// TransactionOutput is a detached final Note output, including generated change.
type TransactionOutput struct {
	output  PreparedOutput
	context NoteOutputContext
}

func transactionOutputs(context protocol.TransactionContext, outputs []PreparedOutput) []TransactionOutput {
	result := make([]TransactionOutput, len(outputs))
	for i, value := range outputs {
		if value.Note != nil {
			copy := *value.Note
			copy.Payload = append([]byte(nil), copy.Payload...)
			value.Note = &copy
		}
		if value.Amount != nil {
			value.Amount = new(big.Int).Set(value.Amount)
		}
		value.Delivery.Ciphertext = append([]byte(nil), value.Delivery.Ciphertext...)
		value.Delivery.Clue = append([]byte(nil), value.Delivery.Clue...)
		result[i] = TransactionOutput{output: value, context: NoteOutputContext{ChainID: new(big.Int).Set(context.ChainID), ProtocolAddress: context.ProtocolAddress, TokenAddress: context.TokenAddress}}
	}
	return result
}
func (n TransactionOutput) ExportPaymentNote() (string, error) {
	return exportPaymentNote(n.output.Note, n.output.Commitment)
}
func (n TransactionOutput) Commitment() common.Hash { return n.output.Commitment }
func (n TransactionOutput) Identity() common.Hash   { return n.output.Identity }
func (n TransactionOutput) Amount() *big.Int {
	if n.output.Amount == nil {
		return nil
	}
	return new(big.Int).Set(n.output.Amount)
}
func (n TransactionOutput) Change() bool { return n.output.Change }
func (p *PreparedTransaction) GetOutputs() []TransactionOutput {
	return transactionOutputs(p.operation.Plan.Context, p.operation.Plan.Outputs)
}

// GetOutputs lets a managed adapter share the exact same detached output representation.
func (o *Execution) GetOutputs() []TransactionOutput {
	return transactionOutputs(o.Plan.Context, o.Plan.Outputs)
}

func (n TransactionOutput) Context() NoteOutputContext {
	c := n.context
	if c.ChainID != nil {
		c.ChainID = new(big.Int).Set(c.ChainID)
	}
	return c
}
func (n TransactionOutput) To() string { return n.output.Address }
func (n TransactionOutput) Nonce() *big.Int {
	if n.output.Note == nil {
		return nil
	}
	return core.ElementToBigInt(n.output.Note.Nonce)
}
func (n TransactionOutput) Recipient() NoteRecipient {
	if n.output.Note == nil {
		return IdentityRecipient{Identity: n.output.Identity}
	}
	if common.IsHexAddress(n.output.Address) {
		envelope := len(n.output.Delivery.Ciphertext) > 0
		return EVMRecipient{Address: common.HexToAddress(n.output.Address), Envelope: &envelope, Nonce: n.Nonce()}
	}
	return InvisiRecipient{Address: n.output.Address, Nonce: n.Nonce()}
}
