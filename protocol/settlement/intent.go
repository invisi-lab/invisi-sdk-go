// Package settlement implements the public settlement wire codec for the current
// generated Protocol public operation ABIs. It contains no private note logic.
package settlement

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

//go:generate node generate.mjs
//go:embed protocol.abi.json
var protocolABIJSON string
var contractABI = mustABI()
var ErrCalldata = errors.New("noncanonical or unsupported settlement calldata")

func mustABI() abi.ABI {
	a, e := abi.JSON(strings.NewReader(protocolABIJSON))
	if e != nil {
		panic(e)
	}
	return a
}

type InputSource struct {
	TreeId     *big.Int
	MerkleRoot [32]byte
}
type Params struct {
	Token             common.Address
	Sources           []InputSource
	Nullifiers        [][32]byte
	NoteCommitments   [][32]byte
	WithdrawalAddress common.Address
	WithdrawalAmount  *big.Int
	Deadline          *big.Int
	FeeRecipient      common.Address
	FeeAmount         *big.Int
}
type Delivery struct {
	Ciphertext []byte
	Clue       []byte
}
type Permit struct {
	Deadline *big.Int
	V        uint8
	R        [32]byte
	S        [32]byte
}
type Funding struct {
	Payer     common.Address
	Amount    *big.Int
	SigNonce  *big.Int
	Deadline  *big.Int
	Signature []byte
	Permit    Permit
}
type wireRequest struct {
	Params     Params
	Deliveries []Delivery
	Funding    Funding
}
type Request struct {
	Method        string
	Members       []Request
	Shield        ShieldParams
	ShieldOutputs []ShieldOutput
	Creation      CreateCreditParams
	Draw          CreditDrawParams
	Update        UpdateCreditParams
	Revocation    RevokeCreditParams
	Proof         []byte
	Params        Params
	Deliveries    []Delivery
	Funding       Funding
}

// Decode rejects unsupported selectors, trailing bytes, nonzero padding and
// alternate offsets by requiring an exact generated-ABI encode/decode round trip.
func Decode(data []byte) (request *Request, err error) {
	defer func() {
		if recover() != nil {
			request = nil
			err = ErrCalldata
		}
	}()
	if len(data) < 4 {
		return nil, ErrCalldata
	}
	method, e := contractABI.MethodById(data[:4])
	if e != nil || !supportedMethod(method) {
		return nil, ErrCalldata
	}
	data = bytes.Clone(data)
	values, e := method.Inputs.Unpack(data[4:])
	if e != nil {
		return nil, fmt.Errorf("%w: %v", ErrCalldata, e)
	}
	request = &Request{Method: method.RawName}
	switch method.RawName {
	case "transact":
		if method.Inputs[1].Type.T == abi.SliceTy {
			members := *abi.ConvertType(values[1], new([]wireRequest)).(*[]wireRequest)
			if len(members) < 1 || len(members) > 5 {
				return nil, ErrCalldata
			}
			request.Proof = values[0].([]byte)
			for _, m := range members {
				request.Members = append(request.Members, Request{Method: "transact", Params: m.Params, Deliveries: m.Deliveries, Funding: m.Funding})
			}
			break
		}
		wire := *abi.ConvertType(values[1], new(wireRequest)).(*wireRequest)
		request.Proof = values[0].([]byte)
		request.Params = wire.Params
		request.Deliveries = wire.Deliveries
		request.Funding = wire.Funding
	case "shield":
		request.Shield = *abi.ConvertType(values[0], new(ShieldParams)).(*ShieldParams)
		request.ShieldOutputs = *abi.ConvertType(values[1], new([]ShieldOutput)).(*[]ShieldOutput)
		request.Deliveries = *abi.ConvertType(values[2], new([]Delivery)).(*[]Delivery)
		request.Funding = *abi.ConvertType(values[3], new(Funding)).(*Funding)
	case "createCredit":
		request.Proof = values[0].([]byte)
		request.Creation = *abi.ConvertType(values[1], new(CreateCreditParams)).(*CreateCreditParams)
		request.Deliveries = *abi.ConvertType(values[2], new([]Delivery)).(*[]Delivery)
	case "creditDraw":
		request.Proof = values[0].([]byte)
		request.Draw = *abi.ConvertType(values[1], new(CreditDrawParams)).(*CreditDrawParams)
		request.Deliveries = *abi.ConvertType(values[2], new([]Delivery)).(*[]Delivery)
	case "updateCredit":
		request.Proof = values[0].([]byte)
		request.Update = *abi.ConvertType(values[1], new(UpdateCreditParams)).(*UpdateCreditParams)
		request.Deliveries = *abi.ConvertType(values[2], new([]Delivery)).(*[]Delivery)
	case "revokeCredit":
		request.Revocation = RevokeCreditParams{Token: values[0].(common.Address), Owner: values[1].(common.Address), Shard: values[2].(uint64), Slot: values[3].(uint16), Deadline: values[4].(*big.Int), FeeRecipient: values[5].(common.Address), FeeAmount: values[6].(*big.Int), Signature: values[7].([]byte)}
	}
	canonical, e := Encode(request)
	if e != nil || !bytes.Equal(data, canonical) {
		return nil, ErrCalldata
	}
	return request, nil
}
func supportedMethod(method *abi.Method) bool {
	switch method.RawName {
	case "transact":
		return len(method.Inputs) == 2 && (method.Inputs[1].Type.T == abi.TupleTy || method.Inputs[1].Type.T == abi.SliceTy)
	case "shield", "createCredit", "creditDraw", "updateCredit", "revokeCredit":
		return true
	}
	return false
}
func Encode(r *Request) (encoded []byte, err error) {
	defer func() {
		if recover() != nil {
			encoded = nil
			err = ErrCalldata
		}
	}()
	if r == nil {
		return nil, ErrCalldata
	}
	switch r.Method {
	case "transact":
		for name, method := range contractABI.Methods {
			if method.RawName != "transact" {
				continue
			}
			if len(r.Members) > 0 && method.Inputs[1].Type.T == abi.SliceTy {
				members := []wireRequest{}
				for _, m := range r.Members {
					members = append(members, wireRequest{m.Params, m.Deliveries, m.Funding})
				}
				return contractABI.Pack(name, r.Proof, members)
			}
			if len(r.Members) == 0 && method.Inputs[1].Type.T == abi.TupleTy {
				return contractABI.Pack(name, r.Proof, wireRequest{r.Params, r.Deliveries, r.Funding})
			}
		}
		return nil, ErrCalldata
	case "shield":
		return contractABI.Pack("shield", r.Shield, r.ShieldOutputs, r.Deliveries, r.Funding)
	case "createCredit":
		return contractABI.Pack("createCredit", r.Proof, r.Creation, r.Deliveries)
	case "creditDraw":
		return contractABI.Pack("creditDraw", r.Proof, r.Draw, r.Deliveries)
	case "updateCredit":
		return contractABI.Pack("updateCredit", r.Proof, r.Update, r.Deliveries)
	case "revokeCredit":
		p := r.Revocation
		return contractABI.Pack("revokeCredit", p.Token, p.Owner, p.Shard, p.Slot, p.Deadline, p.FeeRecipient, p.FeeAmount, p.Signature)
	}
	return nil, ErrCalldata
}

// DecodeCalls returns logical operations in their original execution order.
// Aggregate members retain native shapes; callers must authenticate mined receipts separately.
func DecodeCalls(data []byte) ([]Request, error) {
	r, err := Decode(data)
	if err != nil {
		return nil, err
	}
	if len(r.Members) == 0 {
		return []Request{*r}, nil
	}
	first := 0
	homogeneous, small := true, true
	seen := map[[32]byte]bool{}
	for _, member := range r.Members {
		n, m := len(member.Params.Nullifiers), len(member.Params.NoteCommitments)
		c := n + m
		if c < 1 || c > 10 || n == 0 && m < 2 {
			return nil, ErrCalldata
		}
		if first == 0 {
			first = c
		}
		homogeneous = homogeneous && first == c
		small = small && c <= 5
		for _, values := range [][][32]byte{member.Params.Nullifiers, member.Params.NoteCommitments} {
			for _, v := range values {
				if seen[v] {
					return nil, ErrCalldata
				}
				seen[v] = true
			}
		}
	}
	count := len(r.Members)
	if !(homogeneous && (count == 1 || count == 2 || count == 5) || small && count >= 2) {
		return nil, ErrCalldata
	}
	return r.Members, nil
}

// RequestHash matches the SDK canonical method-separated intent. Proof and the
// funding and revocation signatures are cleared; all other calldata bytes remain bound.
func RequestHash(data []byte) (common.Hash, error) {
	r, e := DecodeSingle(data)
	if e != nil {
		return common.Hash{}, e
	}
	return Hash(r)
}
func Hash(r *Request) (hash common.Hash, err error) {
	defer func() {
		if recover() != nil {
			hash = common.Hash{}
			err = ErrCalldata
		}
	}()
	if r == nil {
		return common.Hash{}, ErrCalldata
	}
	unsigned := *r
	unsigned.Proof = nil
	unsigned.Funding = r.Funding
	unsigned.Funding.Signature = nil
	unsigned.Revocation.Signature = nil
	canonical, e := Encode(&unsigned)
	if e != nil {
		return common.Hash{}, e
	}
	stringType, _ := abi.NewType("string", "", nil)
	bytesType, _ := abi.NewType("bytes", "", nil)
	args := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: bytesType}}
	encoded, e := args.Pack("InvisiBroadcasterIntentV1", r.Method, canonical)
	if e != nil {
		return common.Hash{}, e
	}
	return crypto.Keccak256Hash(encoded), nil
}

// DecodeSingle excludes aggregates at single-operation relay boundaries.
func DecodeSingle(data []byte) (*Request, error) {
	r, err := Decode(data)
	if err != nil {
		return nil, err
	}
	if len(r.Members) != 0 {
		return nil, ErrCalldata
	}
	return r, nil
}
