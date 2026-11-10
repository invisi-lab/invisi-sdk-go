package provenance

// ProvenanceProof contains persistable public evidence, without Note plaintext or private witnesses.
// Before submission, use the protocol decoder to bind Data to the exact native PublicInputs.
type ProvenanceProof struct {
	Format              string         `json:"format"`
	Data                string         `json:"data"`
	Context             ProofContext   `json:"context"`
	MainCircuitCapacity int            `json:"mainCircuitCapacity"`
	PublicInputs        []string       `json:"publicInputs"`
	OperationDigest     string         `json:"operationDigest"`
	Sidecars            []ProofSidecar `json:"sidecars"`
}

type ProofContext struct {
	ChainID         string `json:"chainId"`
	ProtocolAddress string `json:"protocolAddress"`
	TokenAddress    string `json:"tokenAddress"`
}
