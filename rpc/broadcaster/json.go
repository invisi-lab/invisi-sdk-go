package broadcaster

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func (s *SignedQuote) UnmarshalJSON(raw []byte) error {
	type wireQuote SignedQuote
	var value wireQuote
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&value); e != nil {
		return e
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return e
	}
	if _, present := fields["provenanceRequired"]; present && !value.ProvenanceRequired {
		return errors.New("provenanceRequired must be true when present")
	}
	*s = SignedQuote(value)
	return nil
}

func (q *Quote) UnmarshalJSON(raw []byte) error {
	type wire Quote
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&value); e != nil {
		return e
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return e
	}
	if (len(fields) != 7 && !(len(fields) == 8 && value.Schedule != nil)) || (value.Execution != "direct" && value.Execution != "aggregate") {
		return errors.New("invalid broadcaster quote execution or fields")
	}
	*q = Quote(value)
	return nil
}

func (p *MemberProof) UnmarshalJSON(raw []byte) error {
	type wire MemberProof
	var value wire
	if err := strictPublicObject(raw, &value); err != nil {
		return err
	}
	*p = MemberProof(value)
	return nil
}

func (p *Submission) UnmarshalJSON(raw []byte) error {
	type wire Submission
	var value wire
	if err := strictPublicObject(raw, &value); err != nil {
		return err
	}
	*p = Submission(value)
	return nil
}

func strictPublicObject(raw []byte, value any) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("public envelope must be an object")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("trailing public envelope data")
	}
	return nil
}
