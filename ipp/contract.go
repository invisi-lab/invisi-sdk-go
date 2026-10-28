package ipp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// strictObject accepts exact wire keys and rejects retired or misspelled fields.
// Open extension maps (metadata, paymaster context and circuit witnesses) remain open.
func strictObject(data []byte, dst interface{}) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if fields == nil {
		return fmt.Errorf("ipp: data must be an object")
	}
	typ := reflect.TypeOf(dst).Elem()
	allowed := make(map[string]bool, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		allowed[strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]] = true
	}
	for key := range fields {
		if !allowed[key] {
			return fmt.Errorf("ipp: unsupported field %q", key)
		}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func requireStrings(values ...string) error {
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("ipp: required field is empty")
		}
	}
	return nil
}

func validatePublicRequest(value interface{}) error {
	switch p := value.(type) {
	case InfoRequest:
		return requireStrings(p.IdentityCommitment)
	case ProposeRequest:
		return requireStrings(p.IdentityCommitment, p.Amount)
	case AuthorizeRequest:
		return requireStrings(p.SpenderIdentityCommitment, p.OwnerIdentityCommitment, p.Amount, p.Expiry)
	case NotifyRequest:
		return requireStrings(p.Commitment, p.Action)
	case WithdrawRequest:
		if err := requireStrings(p.Network, p.Operator, p.TokenAddress, p.TreeID); err != nil {
			return err
		}
		if len(p.NoteCommitments) == 0 {
			return fmt.Errorf("ipp: note_commitments must be nonempty")
		}
		return requireStrings(p.NoteCommitments...)
	case RelayRequest:
		return validateRelayRequest(p)
	default:
		return fmt.Errorf("ipp: unsupported request type")
	}
}

var relayAddress = regexp.MustCompile(`^0[xX][0-9a-fA-F]{40}$`)
var relayCalldata = regexp.MustCompile(`^0[xX](?:[0-9a-fA-F]{2})+$`)

var proofCircuits = map[RelayOperation][]string{
	RelayTransact:     {"transact_user_slots14", "transact_user_slots56", "transact_user_slots112", "transact_user_slots168", "transact_aggregate_14_4", "transact_aggregate_14_8", "transact_aggregate_14_12"},
	RelayCreateCredit: {"create_credit"},
	RelayCreditDraw:   {"credit_draw_slots1", "credit_draw_slots12", "credit_draw_slots50", "credit_draw_slots120"},
	RelayCancelCredit: {"credit_cancel"},
	RelayTopUpCredit:  {"credit_top_up"},
}

func validateRelayRequest(p RelayRequest) error {
	switch p.Operation {
	case RelayShield, RelayTransact, RelayCreateCredit, RelayCreditDraw, RelayCancelCredit, RelayTopUpCredit:
	default:
		return fmt.Errorf("ipp: unsupported relay operation %q", p.Operation)
	}
	switch p.Mode {
	case "calldata":
		if err := requireStrings(p.Network); err != nil {
			return err
		}
		if !relayAddress.MatchString(p.ProtocolAddress) || strings.Trim(p.ProtocolAddress[2:], "0") == "" {
			return fmt.Errorf("ipp: relay protocol_address must be a nonzero address")
		}
		if !relayCalldata.MatchString(p.Calldata) {
			return fmt.Errorf("ipp: malformed relay calldata")
		}
		if p.Witness != nil || p.Circuit != "" {
			return fmt.Errorf("ipp: unsupported calldata relay fields")
		}
	case "witness", "prove":
		if p.ProtocolAddress != "" || p.Calldata != "" || (p.Mode == "prove" && p.Network != "") {
			return fmt.Errorf("ipp: unsupported proof relay fields")
		}
		if p.Mode == "witness" {
			if err := requireStrings(p.Network); err != nil {
				return err
			}
		}
		found := false
		for _, circuit := range proofCircuits[p.Operation] {
			if circuit == p.Circuit {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ipp: unsupported circuit for relay operation")
		}
		raw, err := json.Marshal(p.Witness)
		if err != nil {
			return fmt.Errorf("ipp: invalid relay witness: %w", err)
		}
		var witness map[string]json.RawMessage
		if err = json.Unmarshal(raw, &witness); err != nil || witness == nil {
			return fmt.Errorf("ipp: relay witness must be an object")
		}
	default:
		return fmt.Errorf("ipp: unsupported relay mode %q", p.Mode)
	}
	return nil
}

func acceptedStatus(status string) error {
	if status != "accepted" && status != "rejected" {
		return fmt.Errorf("ipp: invalid response status %q", status)
	}
	return nil
}

func requiredPointers(values ...*string) error {
	for _, value := range values {
		if value == nil || *value == "" {
			return fmt.Errorf("ipp: required response field is missing")
		}
	}
	return nil
}

func validatePublicResponse(value interface{}) error {
	switch p := value.(type) {
	case ProposeResponse:
		if err := acceptedStatus(p.Status); err != nil {
			return err
		}
		if p.Status == "accepted" {
			return requiredPointers(p.Nonce, p.NoteCommitment)
		}
	case AuthorizeResponse:
		if err := acceptedStatus(p.Status); err != nil {
			return err
		}
		if p.Status == "accepted" {
			return requiredPointers(p.Nonce, p.CreditCommitment, p.CreditSharedKeyHash, p.CreditSharedKey)
		}
	case WithdrawResponse:
		if err := acceptedStatus(p.Status); err != nil {
			return err
		}
		if p.Status != "accepted" {
			return nil
		}
		if p.Witness == nil {
			return fmt.Errorf("ipp: withdraw witness is required")
		}
		w := p.Witness
		if err := requireStrings(w.MerkleRoot, w.TotalAmount); err != nil {
			return err
		}
		if len(w.MerkleProofs) == 0 {
			return fmt.Errorf("ipp: Merkle proofs are required")
		}
		for _, proof := range w.MerkleProofs {
			if err := requireStrings(proof.LeafIndex); err != nil {
				return err
			}
			if len(proof.MerklePath) != 16 {
				return fmt.Errorf("ipp: Merkle path must have 16 fields")
			}
			if err := requireStrings(proof.MerklePath...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *IPPPayloadV1) UnmarshalJSON(data []byte) error {
	type wire IPPPayloadV1
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = IPPPayloadV1(next)
	return nil
}

func (v *IPPVoucherV1) UnmarshalJSON(data []byte) error {
	type wire IPPVoucherV1
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = IPPVoucherV1(next)
	return nil
}

func (v *InfoRequest) UnmarshalJSON(data []byte) error {
	type wire InfoRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = InfoRequest(next)
	return nil
}

func (v *ProposeRequest) UnmarshalJSON(data []byte) error {
	type wire ProposeRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = ProposeRequest(next)
	return nil
}

func (v *AuthorizeRequest) UnmarshalJSON(data []byte) error {
	type wire AuthorizeRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = AuthorizeRequest(next)
	return nil
}

func (v *NotifyRequest) UnmarshalJSON(data []byte) error {
	type wire NotifyRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = NotifyRequest(next)
	return nil
}

func (v *WithdrawRequest) UnmarshalJSON(data []byte) error {
	type wire WithdrawRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = WithdrawRequest(next)
	return nil
}

func (v *ProposeResponse) UnmarshalJSON(data []byte) error {
	type wire ProposeResponse
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = ProposeResponse(next)
	return nil
}

func (v *AuthorizeResponse) UnmarshalJSON(data []byte) error {
	type wire AuthorizeResponse
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = AuthorizeResponse(next)
	return nil
}

func (v *WithdrawResponse) UnmarshalJSON(data []byte) error {
	type wire WithdrawResponse
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = WithdrawResponse(next)
	return nil
}

func (v *WithdrawWitness) UnmarshalJSON(data []byte) error {
	type wire WithdrawWitness
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = WithdrawWitness(next)
	return nil
}

func (v *MerkleMembershipProof) UnmarshalJSON(data []byte) error {
	type wire MerkleMembershipProof
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = MerkleMembershipProof(next)
	return nil
}

func (v *RelayRequest) UnmarshalJSON(data []byte) error {
	type wire RelayRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	allowed := map[string]bool{"mode": true, "operation": true}
	switch next.Mode {
	case "calldata":
		for _, key := range []string{"network", "protocol_address", "calldata"} {
			allowed[key] = true
		}
	case "witness":
		for _, key := range []string{"network", "witness", "circuit"} {
			allowed[key] = true
		}
	case "prove":
		for _, key := range []string{"witness", "circuit"} {
			allowed[key] = true
		}
	}
	for key := range fields {
		if !allowed[key] {
			return fmt.Errorf("ipp: unsupported relay field %q", key)
		}
	}
	if err := validateRelayRequest(RelayRequest(next)); err != nil {
		return err
	}
	*v = RelayRequest(next)
	return nil
}

func (v *RelayCalldataRequest) UnmarshalJSON(data []byte) error {
	type wire RelayCalldataRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	if next.Mode != "calldata" {
		return fmt.Errorf("ipp: invalid relay mode")
	}
	relay := RelayRequest{Mode: next.Mode, Operation: next.Operation, Network: next.Network, ProtocolAddress: next.ProtocolAddress, Calldata: next.Calldata}
	if err := validateRelayRequest(relay); err != nil {
		return err
	}
	*v = RelayCalldataRequest(next)
	return nil
}

func (v *RelayWitnessRequest) UnmarshalJSON(data []byte) error {
	type wire RelayWitnessRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	if next.Mode != "witness" {
		return fmt.Errorf("ipp: invalid relay mode")
	}
	relay := RelayRequest{Mode: next.Mode, Operation: next.Operation, Network: next.Network, Witness: next.Witness, Circuit: next.Circuit}
	if err := validateRelayRequest(relay); err != nil {
		return err
	}
	*v = RelayWitnessRequest(next)
	return nil
}

func (v *RelayProveRequest) UnmarshalJSON(data []byte) error {
	type wire RelayProveRequest
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	if next.Mode != "prove" {
		return fmt.Errorf("ipp: invalid relay mode")
	}
	relay := RelayRequest{Mode: next.Mode, Operation: next.Operation, Witness: next.Witness, Circuit: next.Circuit}
	if err := validateRelayRequest(relay); err != nil {
		return err
	}
	*v = RelayProveRequest(next)
	return nil
}

func (v *RelayResponse) UnmarshalJSON(data []byte) error {
	type wire RelayResponse
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = RelayResponse(next)
	return nil
}

func (v *RelayProveResponse) UnmarshalJSON(data []byte) error {
	type wire RelayProveResponse
	var next wire
	if err := strictObject(data, &next); err != nil {
		return err
	}
	*v = RelayProveResponse(next)
	return nil
}
