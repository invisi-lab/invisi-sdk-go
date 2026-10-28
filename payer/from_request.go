package payer

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/ipp"
	"math/big"
)

type BuildShieldFromRequestParams struct {
	Request               ipp.IPPRequest
	Nonce, NoteCommitment common.Hash
	Amount                *big.Int
	ProtocolAddress       common.Address
	Deadline              *big.Int
	Ciphertext            []byte
	Funding               FundingParams
	Caller                *common.Address
}

func BuildShieldFromRequest(p BuildShieldFromRequestParams) (*Calldata, error) {
	if err := ipp.ValidateIPPRequest(p.Request); err != nil {
		return nil, err
	}
	allowed := false
	for _, t := range p.Request.Payload.PaymentTypes {
		if t == ipp.PaymentTypeOpen {
			allowed = true
		}
	}
	if !allowed {
		return nil, fmt.Errorf("request does not accept shield payments")
	}
	if err := CheckUint(p.Amount, 128, "amount"); err != nil {
		return nil, err
	}
	r := p.Request.Payload
	if r.Amount != nil {
		n, ok := new(big.Int).SetString(*r.Amount, 10)
		if !ok || n.Cmp(p.Amount) != 0 {
			return nil, fmt.Errorf("request amount mismatch")
		}
	}
	chain, err := ipp.CAIP2ToChainID(p.Request.Network)
	if err != nil {
		return nil, err
	}
	identity, err := core.HexToElement(r.IdentityCommitment)
	if err != nil {
		return nil, err
	}
	if !common.IsHexAddress(r.TokenAddress) {
		return nil, fmt.Errorf("invalid request token address")
	}
	token := common.HexToAddress(r.TokenAddress)
	valid, err := VerifyNoteCommitment(VerifyNoteCommitmentParams{ChainID: chain, IdentityCommitment: identity.Bytes(), TokenAddress: token, Amount: p.Amount, Nonce: p.Nonce, NoteCommitment: p.NoteCommitment})
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("noteCommitment verification failed")
	}
	return BuildShieldCalldata(BuildShieldCalldataParams{Context: TransactionContext{ChainID: chain, ProtocolAddress: p.ProtocolAddress, TokenAddress: token},
		Params:  ShieldParams{OutputCommitments: [][32]byte{p.NoteCommitment}, Deadline: p.Deadline},
		Outputs: []ShieldOutput{{IdentityCommitment: identity.Bytes(), Amount: p.Amount, Nonce: p.Nonce}}, Ciphertexts: [][]byte{p.Ciphertext}, Funding: p.Funding, Caller: p.Caller})
}
