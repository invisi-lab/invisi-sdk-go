package primitives

import "github.com/invisi-lab/invisi-sdk-go/primitives/internal/cryptography/fmd"

type PreparedClue = fmd.PreparedClue

func DeriveClueKey(key []byte) ([]byte, error) { return fmd.DeriveClueKey(key) }

func ValidateClueKey(key []byte) error { return fmd.ValidateClueKey(key) }

func CreateClue(key []byte) ([]byte, error) { return fmd.CreateClue(key) }

func CreateClueFromSeed(key, seed []byte) ([]byte, error) { return fmd.CreateClueFromSeed(key, seed) }

func PrepareClue(clue []byte) (*PreparedClue, error) { return fmd.PrepareClue(clue) }

func DetectClue(key, clue []byte, checkedBits int) (bool, error) {
	return fmd.DetectClue(key, clue, checkedBits)
}

func DetectPreparedClue(key []byte, prepared *PreparedClue, checkedBits int) (bool, error) {
	return fmd.DetectPreparedClue(key, prepared, checkedBits)
}
