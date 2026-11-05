package transaction

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
)

var OutputSealKey = func() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return key
}()

func (n NoteOutput) MarshalJSON() ([]byte, error) {
	if n.saved == nil {
		return nil, errors.New("outputs must be created with NewNoteOutput")
	}
	raw, e := json.Marshal(n.saved)
	if e != nil {
		return nil, e
	}
	mac := hmac.New(sha256.New, OutputSealKey)
	mac.Write(raw)
	return json.Marshal(struct {
		Material json.RawMessage
		Seal     []byte
	}{raw, mac.Sum(nil)})
}

func (n *NoteOutput) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Material json.RawMessage
		Seal     []byte
	}
	if e := json.Unmarshal(raw, &wire); e != nil {
		return e
	}
	mac := hmac.New(sha256.New, OutputSealKey)
	mac.Write(wire.Material)
	if !hmac.Equal(wire.Seal, mac.Sum(nil)) {
		return errors.New("outputs must be created with NewNoteOutput; invalid output snapshot")
	}
	var saved OutputIntent
	if e := json.Unmarshal(wire.Material, &saved); e != nil {
		return e
	}
	n.saved = &saved
	return nil
}
