package prover

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// bundleJSON is the same proof exchange format as the TypeScript SDK.
type bundleJSON struct {
	FormatVersion  int           `json:"formatVersion"`
	CircuitName    string        `json:"circuitName"`
	Target         string        `json:"target"`
	BBVersion      string        `json:"bbVersion"`
	ArtifactSHA256 string        `json:"artifactSha256"`
	VKSHA256       string        `json:"vkSha256"`
	Proof          hexutil.Bytes `json:"proofHex"`
	ProofSHA256    string        `json:"proofSha256"`
	PublicInputs   []common.Hash `json:"publicInputs"`
}

func (b ProofBundle) MarshalJSON() ([]byte, error) {
	return json.Marshal(bundleJSON{b.FormatVersion, b.CircuitName, b.Target, b.BBVersion, b.ArtifactSHA256, b.VKSHA256, b.Proof, b.ProofSHA256, b.PublicInputs})
}

func (b *ProofBundle) UnmarshalJSON(data []byte) error {
	var wire bundleJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	*b = ProofBundle{wire.FormatVersion, wire.CircuitName, wire.Target, wire.BBVersion, wire.ArtifactSHA256, wire.VKSHA256, wire.Proof, wire.ProofSHA256, wire.PublicInputs}
	return nil
}
