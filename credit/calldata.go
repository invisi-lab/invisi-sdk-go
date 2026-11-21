package credit

import (
	_ "embed"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/invisi-labs/invisi-sdk-go/primitives"
	"github.com/invisi-labs/invisi-sdk-go/protocol"
)

//go:embed data/abi.json
var abiJSON string

//go:embed data/current-abi.json
var currentABIJSON string

func ContractABI() (abi.ABI, error)       { return abi.JSON(strings.NewReader(currentABIJSON)) }
func LegacyContractABI() (abi.ABI, error) { return abi.JSON(strings.NewReader(abiJSON)) }
func pack(name string, args ...any) ([]byte, error) {
	a, err := ContractABI()
	if err != nil {
		return nil, err
	}
	return a.Pack(name, args...)
}
func (p CreateParams) Calldata(c Context, proof []byte, deliveries []protocol.Delivery) ([]byte, error) {
	if _, _, err := p.Statement(c, deliveries); err != nil {
		return nil, err
	}
	encoded, err := encodeProof(proof)
	if err != nil {
		return nil, err
	}
	p.FeeAmount, _ = normalizeFee(p.FeeRecipient, p.FeeAmount)
	return packCreate(p, encoded, deliveries)
}
func (p UpdateParams) Calldata(c Context, proof []byte, deliveries []protocol.Delivery) ([]byte, error) {
	if _, _, err := p.Statement(c, deliveries); err != nil {
		return nil, err
	}
	encoded, err := encodeProof(proof)
	if err != nil {
		return nil, err
	}
	p.FeeAmount, _ = normalizeFee(p.FeeRecipient, p.FeeAmount)
	return pack("updateCredit", encoded, p, deliveries)
}
func (p DrawParams) Calldata(c Context, proof []byte, deliveries []protocol.Delivery) ([]byte, error) {
	if _, _, err := p.Statement(c, deliveries); err != nil {
		return nil, err
	}
	encoded, err := encodeProof(proof)
	if err != nil {
		return nil, err
	}
	p.FeeAmount, _ = normalizeFee(p.FeeRecipient, p.FeeAmount)
	return pack("creditDraw", encoded, p, deliveries)
}

type RevokeParams struct {
	Token, Owner common.Address
	GrantID      *big.Int `abi:"grantId"`
	Deadline     *big.Int
	FeeRecipient common.Address
	FeeAmount    *big.Int
}

func (p RevokeParams) validate(c Context) error {
	if _, err := normalizeFee(p.FeeRecipient, p.FeeAmount); err != nil {
		return err
	}
	if err := c.validate(p.Token, p.Deadline); err != nil {
		return err
	}
	if p.Owner == (common.Address{}) || !ValidGrantID(p.GrantID) {
		return fmt.Errorf("invalid Credit owner or revocation namespace")
	}
	return nil
}
func (p RevokeParams) TypedData(c Context) (apitypes.TypedData, error) {
	if err := p.validate(c); err != nil {
		return apitypes.TypedData{}, err
	}
	fee, _ := normalizeFee(p.FeeRecipient, p.FeeAmount)
	chain := ethmath.HexOrDecimal256(*primitives.ElementToBigInt(c.ChainID))
	return apitypes.TypedData{Types: apitypes.Types{"EIP712Domain": {{Name: "name", Type: "string"}, {Name: "version", Type: "string"}, {Name: "chainId", Type: "uint256"}, {Name: "verifyingContract", Type: "address"}}, "RevokeCredit": {{Name: "token", Type: "address"}, {Name: "owner", Type: "address"}, {Name: "grantId", Type: "uint128"}, {Name: "deadline", Type: "uint256"}, {Name: "feeRecipient", Type: "address"}, {Name: "feeAmount", Type: "uint256"}}}, PrimaryType: "RevokeCredit", Domain: apitypes.TypedDataDomain{Name: "InvisibleProtocol", Version: "1", ChainId: &chain, VerifyingContract: common.BigToAddress(primitives.ElementToBigInt(c.ProtocolAddress)).Hex()}, Message: apitypes.TypedDataMessage{"token": p.Token.Hex(), "owner": p.Owner.Hex(), "grantId": p.GrantID.String(), "deadline": p.Deadline.String(), "feeRecipient": p.FeeRecipient.Hex(), "feeAmount": fee.String()}}, nil
}

// Calldata permits a direct owner call or an externally supplied EOA/ERC-1271 authorization.
// Contract settlement authenticates signatures and the current grant status; no ZK state is required.
func (p RevokeParams) Calldata(c Context, signature []byte, caller common.Address) ([]byte, error) {
	if err := p.validate(c); err != nil {
		return nil, err
	}
	if len(signature) == 0 && caller != p.Owner {
		return nil, fmt.Errorf("direct revocation requires owner caller")
	}
	fee, _ := normalizeFee(p.FeeRecipient, p.FeeAmount)
	return pack("revokeCredit", p.Token, p.Owner, p.GrantID, p.Deadline, p.FeeRecipient, fee, signature)
}

func encodeProof(proof []byte) ([]byte, error) {
	if len(proof) == 0 {
		return nil, fmt.Errorf("empty Credit proof")
	}
	schema, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return nil, err
	}
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}
	return (abi.Arguments{{Type: schema}, {Type: bytesType}}).Pack(uint8(0), proof)
}

func packCreate(p CreateParams, proof []byte, deliveries []protocol.Delivery) ([]byte, error) {
	if p.CreationABI == "legacy" {
		return nil, fmt.Errorf("historical creation ABI cannot encode current grant ID")
	}
	if p.CreationABI != "" && p.CreationABI != "initial-payment" {
		return nil, fmt.Errorf("unsupported creation ABI")
	}
	if p.HasInitialPayment == nil {
		no := false
		p.HasInitialPayment = &no
	}
	a, err := abi.JSON(strings.NewReader(currentABIJSON))
	if err != nil {
		return nil, err
	}
	total := p.TotalAmount
	if total == nil {
		total = new(big.Int)
	}
	args := struct {
		Token, Owner      common.Address
		GrantID           *big.Int `abi:"grantId"`
		CreditCommitment  common.Hash
		Deadline          *big.Int
		FeeRecipient      common.Address
		FeeAmount         *big.Int
		HasInitialPayment bool
		TotalAmount       *big.Int
		NoteCommitment    common.Hash
	}{p.Token, p.Owner, p.GrantID, p.CreditCommitment, p.Deadline, p.FeeRecipient, p.FeeAmount, *p.HasInitialPayment, total, p.NoteCommitment}
	return a.Pack("createCredit", proof, args, deliveries)
}
