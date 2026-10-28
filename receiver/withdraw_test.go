package receiver

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/payer"
	"math/big"
	"testing"
)

func unsignedFixture(t *testing.T) BuildUnsignedWithdrawalCalldataParams {
	t.Helper()
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	token := common.HexToAddress("0x2222222222222222222222222222222222222222")
	c := payer.TransactionContext{ChainID: big.NewInt(1), ProtocolAddress: common.HexToAddress("0x3333333333333333333333333333333333333333"), TokenAddress: token}
	note := core.Note{NoteSecret: core.ElementFromUint64(123), ReceivingAddress: core.BigIntToElement(owner.Big()), TokenAddress: core.BigIntToElement(token.Big()), Amount: core.ElementFromUint64(100), Nonce: core.ElementFromUint64(42)}
	nf, err := core.ComputeNoteNullifier(core.ElementFromUint64(1), note.NoteSecret, note.ReceivingAddress, note.TokenAddress, note.Amount, note.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	params := payer.CrossTreeTransactParams{Sources: []payer.InputSource{{TreeId: big.NewInt(1), MerkleRoot: payer.Word(big.NewInt(1))}}, Nullifiers: [][32]byte{nf.Bytes()}, WithdrawalAddress: owner, WithdrawalAmount: big.NewInt(100), Deadline: big.NewInt(1000)}
	return BuildUnsignedWithdrawalCalldataParams{Context: c, Proof: []byte{0x12, 0x34}, Params: params, SourceNotes: []core.Note{note}}
}
func TestUnsignedWithdrawalCurrentBinding(t *testing.T) {
	p := unsignedFixture(t)
	r, err := BuildUnsignedWithdrawalCalldata(p)
	if err != nil {
		t.Fatal(err)
	}
	a, err := payer.ProtocolABI()
	if err != nil {
		t.Fatal(err)
	}
	method, err := a.MethodById(r.Data[:4])
	if err != nil || method.RawName != "transact" {
		t.Fatal("unsigned withdrawal must use current transact")
	}
}
func TestUnsignedWithdrawalRejectsTampering(t *testing.T) {
	tests := map[string]func(*BuildUnsignedWithdrawalCalldataParams){
		"redirected recipient": func(p *BuildUnsignedWithdrawalCalldataParams) {
			p.Params.WithdrawalAddress = common.HexToAddress("0x4444444444444444444444444444444444444444")
		},
		"partial amount": func(p *BuildUnsignedWithdrawalCalldataParams) { p.Params.WithdrawalAmount = big.NewInt(99) },
		"nullifier":      func(p *BuildUnsignedWithdrawalCalldataParams) { p.Params.Nullifiers[0] = payer.Word(big.NewInt(2)) },
		"token": func(p *BuildUnsignedWithdrawalCalldataParams) {
			p.SourceNotes[0].TokenAddress = core.ElementFromUint64(7)
		},
		"private output": func(p *BuildUnsignedWithdrawalCalldataParams) {
			p.Params.OutputCommitments = [][32]byte{payer.Word(big.NewInt(3))}
		},
		"funding": func(p *BuildUnsignedWithdrawalCalldataParams) {
			p.Funding = payer.FundingParams{Payer: p.Context.ProtocolAddress, Amount: big.NewInt(1)}
		},
		"no notes":  func(p *BuildUnsignedWithdrawalCalldataParams) { p.SourceNotes = nil },
		"zero note": func(p *BuildUnsignedWithdrawalCalldataParams) { p.SourceNotes[0].Amount = core.ElementFromUint64(0) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			p := unsignedFixture(t)
			mutate(&p)
			if _, err := BuildUnsignedWithdrawalCalldata(p); err == nil {
				t.Fatal("accepted unauthorized unsigned spend")
			}
		})
	}
}
func TestInvoiceUsesNoteSecret(t *testing.T) {
	invoice, err := GenerateInvoice(GenerateInvoiceParams{MasterKey: core.ElementFromUint64(3), InvoiceID: core.ElementFromUint64(4), ReceivingAddress: common.HexToAddress("0x1111111111111111111111111111111111111111")})
	if err != nil {
		t.Fatal(err)
	}
	secret, err := core.DeriveNoteSecret(core.ElementFromUint64(3), core.ElementFromUint64(4))
	if err != nil {
		t.Fatal(err)
	}
	if !secret.Equal(&invoice.NoteSecret) {
		t.Fatal("wrong invoice secret")
	}
	identity, err := core.ComputeIdentityCommitment(secret, addressToElement(invoice.ReceivingAddress))
	if err != nil {
		t.Fatal(err)
	}
	if identity.Bytes() != [32]byte(invoice.IdentityCommitment) {
		t.Fatal("wrong invoice identity")
	}
}
