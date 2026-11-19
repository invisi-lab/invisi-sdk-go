package receipts

import (
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestNoteOutputsOpaqueApplicationFrames(t *testing.T) {

	decoder, _ := NewDecoder()
	_, log := fixture(t)
	ciphertexts := [][]byte{make([]byte, 2), make([]byte, 4096), make([]byte, 167)}
	var packed []byte
	for i, ciphertext := range ciphertexts {
		ciphertext[0] = []byte{1, 0x81, 0xfe}[i]
		ciphertext[1] = byte(i)
		packed = binary.BigEndian.AppendUint16(packed, uint16(len(ciphertext)))
		packed = append(packed, ciphertext...)
		clue := make([]byte, 68)
		clue[0] = byte(i)
		clue[64] = 24
		packed = append(packed, clue...)
	}
	decode := func(p []byte) (Batch, error) {
		id := common.HexToHash("1")
		tx := common.HexToHash("2")
		op := log("NoteOperation", 1, tx, []any{id, common.Address{1}}, uint8(1), big.NewInt(3), big.NewInt(99), common.Address{}, big.NewInt(0))
		ins := log("CommitmentsInserted", 2, tx, []any{big.NewInt(1)}, uint32(4), [][32]byte{{1}, {2}, {3}})
		out := log("NoteOutputs", 3, tx, []any{id, big.NewInt(1)}, uint32(4), p)
		deposit := log("PublicDeposit", 4, tx, []any{id, common.Address{1}, common.Address{2}}, big.NewInt(1))
		return decoder.Decode(Block{}, []types.Log{op, ins, out, deposit})
	}
	batch, err := decode(packed)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Outputs) != 3 {
		t.Fatal("lost output")
	}
	for i, output := range batch.Outputs {
		if len(output.Ciphertext) != len(ciphertexts[i]) || output.Ciphertext[1] != byte(i) || output.Clue[0] != byte(i) || output.LeafIndex != uint32(4+i) {
			t.Fatal("frame mapping mismatch")
		}
	}
	// A zero-length slot contains only its uint16 prefix, with no clue bytes.
	mixed := append([]byte{0, 0}, packed[2+len(ciphertexts[0])+68:]...)
	emptyBatch, err := decode(mixed)
	if err != nil || len(emptyBatch.Outputs) != 3 {
		t.Fatalf("empty slot: %v", err)
	}
	if len(emptyBatch.Outputs[0].Ciphertext) != 0 || len(emptyBatch.Outputs[0].Clue) != 0 || emptyBatch.Outputs[1].LeafIndex != 5 || emptyBatch.Outputs[1].Ciphertext[1] != 1 {
		t.Fatal("empty slot shifted delivery positions")
	}
	for _, bad := range [][]byte{nil, {0}, packed[:len(packed)-1], append(append([]byte(nil), packed...), 0), append([]byte{0, 0}, packed[2:]...), append([]byte{16, 1}, packed[2:]...)} {
		if _, err := decode(bad); err == nil {
			t.Fatal("invalid framing accepted")
		}
	}
	bad := append([]byte(nil), packed...)
	bad[2] = 0x99
	if _, err := decode(bad); err != nil {
		t.Fatal("opaque version rejected", err)
	}
	bad[2+len(ciphertexts[0])+64] = 23
	if _, err := decode(bad); err == nil {
		t.Fatal("invalid clue precision accepted")
	}
}
