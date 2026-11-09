package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// Object fields must be unique so host parsers cannot interpret different transcripts.
func objectFields(raw []byte) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("wallet JSON object required")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return nil, err
		}
		name, ok := key.(string)
		if !ok {
			return nil, errors.New("invalid wallet field")
		}
		if _, ok = fields[name]; ok {
			return nil, errors.New("duplicate wallet JSON field")
		}
		var value json.RawMessage
		if err = d.Decode(&value); err != nil {
			return nil, err
		}
		fields[name] = value
	}
	if _, err = d.Token(); err != nil {
		return nil, err
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, errors.New("trailing wallet JSON")
	}
	return fields, nil
}
func (r *Request) UnmarshalJSON(raw []byte) error {
	fields, err := objectFields(raw)
	if err != nil {
		return err
	}
	for _, field := range []string{"version", "id", "ownerAddress", "method", "params"} {
		if _, ok := fields[field]; !ok {
			return errors.New("missing wallet request field")
		}
	}
	if len(fields) != 5 {
		return errors.New("unknown wallet request field")
	}
	type plain Request
	var q plain
	if err = Decode(raw, &q); err != nil {
		return err
	}
	if err = ValidateRequest(Request(q)); err != nil {
		return err
	}
	*r = Request(q)
	return nil
}
func (r *Response) UnmarshalJSON(raw []byte) error {
	fields, err := objectFields(raw)
	if err != nil {
		return err
	}
	for _, field := range []string{"version", "id", "ownerAddress", "method"} {
		if _, ok := fields[field]; !ok {
			return errors.New("missing wallet response field")
		}
	}
	_, hasResult := fields["result"]
	failure, hasError := fields["error"]
	if len(fields) != 5 || hasResult == hasError || hasError && bytes.Equal(bytes.TrimSpace(failure), []byte("null")) {
		return errors.New("wallet response requires exactly one result or error")
	}
	type plain Response
	var response plain
	if err = Decode(raw, &response); err != nil {
		return err
	}
	if hasError {
		if _, err = objectFields(failure); err != nil {
			return err
		}
	}
	*r = Response(response)
	return nil
}
