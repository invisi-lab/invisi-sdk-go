package core

import "fmt"

// MaxCiphertextBytes is the SDK's fixed envelope resource bound.
const MaxCiphertextBytes = 4096

// ValidatePayload checks the length-delimited raw application bytes, without interpreting them.
func ValidatePayload(payload []byte) error {
	if len(payload) > 65535 {
		return fmt.Errorf("payload length exceeds uint16")
	}
	return nil
}

// PayloadDataV1 is the memo-only application layout; financial fields belong to the envelope.
type PayloadDataV1 struct {
	Memo []byte `json:"memo,omitempty"`
}
type payloadCodecV1 struct{}

var FundsV1 = payloadCodecV1{}

func (payloadCodecV1) Encode(data PayloadDataV1) ([]byte, error) {
	if err := ValidateMemo(data.Memo); err != nil {
		return nil, err
	}
	return append([]byte{1}, data.Memo...), nil
}
func (payloadCodecV1) Decode(payload []byte) (PayloadDataV1, error) {
	if len(payload) < 1 || payload[0] != 1 {
		return PayloadDataV1{}, fmt.Errorf("unsupported payload version")
	}
	if err := ValidateMemo(payload[1:]); err != nil {
		return PayloadDataV1{}, err
	}
	return PayloadDataV1{Memo: CloneMemo(payload[1:])}, nil
}
