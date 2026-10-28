package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func fmdHex(t testing.TB, s string) []byte {
	t.Helper()
	b, e := hex.DecodeString(s)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

type rustFMDVector struct {
	DetectionKey, ClueKey, SeedHex, Clue, OtherDetectionKey string
	OtherMatches                                            []bool
	Mutations                                               []struct {
		Position int
		Matches  []bool
	}
}

func fmdVectors(t testing.TB) []rustFMDVector {
	t.Helper()
	raw, e := os.ReadFile("testdata/fmd-rust-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var data struct{ Vectors []rustFMDVector }
	if e = json.Unmarshal(raw, &data); e != nil {
		t.Fatal(e)
	}
	return data.Vectors
}
func TestFMDIndependentRustVectors(t *testing.T) {
	for _, v := range fmdVectors(t) {
		key, ck, clue := fmdHex(t, v.DetectionKey), fmdHex(t, v.ClueKey), fmdHex(t, v.Clue)
		got, e := DeriveClueKey(key)
		if e != nil || !bytes.Equal(got, ck) {
			t.Fatalf("derive %s: %x %v", v.DetectionKey, got, e)
		}
		got, e = CreateClueFromSeed(ck, fmdHex(t, v.SeedHex))
		if e != nil || !bytes.Equal(got, clue) {
			t.Fatalf("Rust deterministic clue mismatch: %x %v", got, e)
		}
		prepared, e := PrepareClue(clue)
		if e != nil {
			t.Fatal(e)
		}
		for bits := 1; bits <= 24; bits++ {
			matched, e := DetectPreparedClue(key, prepared, bits)
			if e != nil || !matched {
				t.Fatalf("target prefix %d: %v %v", bits, matched, e)
			}
			matched, e = DetectPreparedClue(fmdHex(t, v.OtherDetectionKey), prepared, bits)
			if e != nil || matched != v.OtherMatches[bits-1] {
				t.Fatalf("wrong-key prefix %d: %v %v", bits, matched, e)
			}
		}
	}
}
func TestFMDOriginalContextAndCiphertextBytes(t *testing.T) {
	for _, v := range fmdVectors(t)[:4] {
		for _, m := range v.Mutations {
			clue := fmdHex(t, v.Clue)
			clue[m.Position] ^= 1
			prepared, e := PrepareClue(clue)
			if e != nil {
				t.Fatal(e)
			}
			for bits := 1; bits <= 24; bits++ {
				got, e := DetectPreparedClue(fmdHex(t, v.DetectionKey), prepared, bits)
				if e != nil || got != m.Matches[bits-1] {
					t.Fatalf("mutation at %d prefix%d: %v %v", m.Position, bits, got, e)
				}
			}
		}
	}
}
func TestFMDCanonicalRustPointCorpus(t *testing.T) {
	raw, e := os.ReadFile("testdata/decaf377-clue-keys.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus []struct {
		Key   string
		Valid bool
	}
	if e = json.Unmarshal(raw, &corpus); e != nil {
		t.Fatal(e)
	}
	if len(corpus) != 322 {
		t.Fatal("unexpected independent corpus")
	}
	for _, v := range corpus {
		if e := ValidateClueKey(fmdHex(t, v.Key[2:])); (e == nil) != v.Valid {
			t.Fatalf("point %s valid%v: %v", v.Key, v.Valid, e)
		}
	}
}
func TestFMDInputBoundariesAndPreparedIsolation(t *testing.T) {
	v := fmdVectors(t)[0]
	key, ck := fmdHex(t, v.DetectionKey), fmdHex(t, v.ClueKey)
	for _, bad := range [][]byte{nil, make([]byte, 31), make([]byte, 33), make([]byte, 32), bytes.Repeat([]byte{255}, 32), fmdHex(t, "ffd93fc39aee5ab9fe8a3cc4afa3935200ec0d9747132d9855298ba657d9aa04")} {
		if _, e := DeriveClueKey(bad); e == nil {
			t.Fatal("invalid detection key accepted")
		}
		if _, e := DetectClue(bad, fmdHex(t, v.Clue), 24); e == nil {
			t.Fatal("invalid detection key accepted in detection")
		}
	}
	for _, bits := range []int{-1, 0, 25, 256} {
		if _, e := DetectClue(key, fmdHex(t, v.Clue), bits); e == nil {
			t.Fatal("invalid checkedBits accepted")
		}
	}
	for _, size := range []int{0, 67, 69} {
		if _, e := PrepareClue(make([]byte, size)); e == nil {
			t.Fatal("invalid clue length accepted")
		}
	}
	if _, e := CreateClueFromSeed(ck, make([]byte, 31)); e == nil {
		t.Fatal("invalid seed length accepted")
	}
	for _, precision := range []byte{0, 1, 8, 12, 23, 25, 255} {
		clue := fmdHex(t, v.Clue)
		clue[64] = precision
		if got, e := DetectClue(key, clue, 1); e != nil || got {
			t.Fatal("changed precision not fail-closed", e)
		}
	}
	for _, offset := range []int{0, 32} {
		for _, fill := range []byte{0, 255} {
			clue := fmdHex(t, v.Clue)
			copy(clue[offset:offset+32], bytes.Repeat([]byte{fill}, 32))
			if got, e := DetectClue(key, clue, 24); e != nil || got {
				t.Fatal("invalid point/scalar not fail-closed", e)
			}
		}
	}
	raw := fmdHex(t, v.Clue)
	prepared, e := PrepareClue(raw)
	if e != nil {
		t.Fatal(e)
	}
	clear(raw)
	if got, e := DetectPreparedClue(key, prepared, 24); e != nil || !got {
		t.Fatal("caller mutation affected prepared clue", e)
	}
	if _, e := DetectPreparedClue(key, nil, 24); e == nil {
		t.Fatal("nil prepared accepted")
	}
	for i := 0; i < 2; i++ {
		clue, e := CreateClue(ck)
		if e != nil || len(clue) != 68 || clue[64] != 24 {
			t.Fatal("secure creation", e)
		}
		if got, e := DetectClue(key, clue, 24); e != nil || !got {
			t.Fatal("random clue detection", e)
		}
	}
}
func BenchmarkFMDCreate(b *testing.B) {
	v := fmdVectors(b)[0]
	ck := fmdHex(b, v.ClueKey)
	seed := fmdHex(b, v.SeedHex)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := CreateClueFromSeed(ck, seed); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkFMDDetect24(b *testing.B) {
	v := fmdVectors(b)[0]
	key := fmdHex(b, v.DetectionKey)
	prepared, e := PrepareClue(fmdHex(b, v.Clue))
	if e != nil {
		b.Fatal(e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got, e := DetectPreparedClue(key, prepared, 24); e != nil || !got {
			b.Fatal(e)
		}
	}
}
func BenchmarkFMDWrongKey(b *testing.B) {
	v := fmdVectors(b)[0]
	key := fmdHex(b, v.OtherDetectionKey)
	prepared, e := PrepareClue(fmdHex(b, v.Clue))
	if e != nil {
		b.Fatal(e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := DetectPreparedClue(key, prepared, 24); e != nil {
			b.Fatal(e)
		}
	}
}
