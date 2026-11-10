package broadcaster

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
)

const ProvenanceRequestHashFormat = "invisi-broadcaster-provenance-intent-v1"

func sidecarParams(data []byte, required bool, sidecars [][]provenance.ProofSidecar) (map[string]any, error) {
	if len(sidecars) > 1 {
		return nil, errors.New("at most one sidecar list required")
	}
	p := map[string]any{"data": "0x" + common.Bytes2Hex(data)}
	var selected []provenance.ProofSidecar
	if len(sidecars) == 1 {
		selected = sidecars[0]
	}
	if !required && len(selected) > 0 {
		return nil, errors.New("quote does not accept provenance evidence")
	}
	if required {
		normalized, e := provenance.ProofSidecarsToWire(selected)
		if e != nil {
			return nil, e
		}
		p["proofSidecars"] = normalized
	}
	return p, nil
}
