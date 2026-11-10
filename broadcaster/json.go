package broadcaster

import (
	"bytes"
	"encoding/json"
	"errors"
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
