package gas

// GasReference is a synchronous pre-proof reference, not a transaction gas limit.
// Use the application's geth transaction pipeline for final estimation and sending.
type GasReference struct {
	GasUnits uint64 `json:"gasUnits"`
}

type GasBenchmarkShape struct {
	Kind             string `json:"kind"`
	Target           string `json:"target"`
	Capacity         int    `json:"capacity"`
	NoteInputs       int    `json:"noteInputs"`
	NoteOutputs      int    `json:"noteOutputs"`
	SourceTrees      int    `json:"sourceTrees"`
	CreditInputs     int    `json:"creditInputs"`
	CreditOutputs    int    `json:"creditOutputs"`
	CreditRemainders int    `json:"creditRemainders"`
	Funding          string `json:"funding"`
	Permit           bool   `json:"permit"`
	DeliveryBytes    int    `json:"deliveryBytes"`
}

type GasBenchmarkDeployment struct {
	ChainID         string `json:"chainId"`
	ProtocolAddress string `json:"protocolAddress"`
	RuntimeCodeHash string `json:"runtimeCodeHash"`
	TokenAddress    string `json:"tokenAddress"`
}

type GasBenchmarkCircuit struct {
	CircuitName    string `json:"circuitName"`
	Target         string `json:"target"`
	ArtifactSHA256 string `json:"artifactSha256"`
	VKSHA256       string `json:"vkSha256"`
}

type GasBenchmarkSample struct {
	GasUsed         string `json:"gasUsed"`
	TransactionHash string `json:"transactionHash"`
	BlockNumber     string `json:"blockNumber"`
	BlockHash       string `json:"blockHash"`
	CalldataBytes   int    `json:"calldataBytes"`
}

type GasBenchmarkFingerprints struct {
	ProtocolSourceSHA256    string `json:"protocolSourceSha256"`
	DirtyDiffSHA256         string `json:"dirtyDiffSha256"`
	CompiledArtifactsSHA256 string `json:"compiledArtifactsSha256"`
	CreditManifestSHA256    string `json:"creditManifestSha256"`
	UserManifestSHA256      string `json:"userManifestSha256"`
	SDKSourceSHA256         string `json:"sdkSourceSha256"`
}

type GasBenchmarkToolchain struct {
	Solidity      string `json:"solidity"`
	Barretenberg  string `json:"barretenberg"`
	Noir          string `json:"noir"`
	EVMVersion    string `json:"evmVersion"`
	OptimizerRuns int    `json:"optimizerRuns"`
	Anvil         string `json:"anvil"`
}

type GasBenchmarkRuntimeContract struct {
	Address         string `json:"address"`
	RuntimeCodeHash string `json:"runtimeCodeHash"`
}

type GasBenchmarkSource struct {
	ProtocolCommit   string                                 `json:"protocolCommit"`
	Fingerprints     *GasBenchmarkFingerprints              `json:"fingerprints,omitempty"`
	Toolchain        *GasBenchmarkToolchain                 `json:"toolchain,omitempty"`
	BlockGasLimit    string                                 `json:"blockGasLimit,omitempty"`
	RuntimeContracts map[string]GasBenchmarkRuntimeContract `json:"runtimeContracts,omitempty"`
	MeasuredAt       string                                 `json:"measuredAt"`
	Chain            string                                 `json:"chain"`
	Hardfork         string                                 `json:"hardfork"`
	Token            string                                 `json:"token"`
	Verifier         string                                 `json:"verifier"`
	Assumptions      []string                               `json:"assumptions"`
}

type GasBenchmarkMeasurement struct {
	ID      string               `json:"id"`
	Shape   GasBenchmarkShape    `json:"shape"`
	Circuit *GasBenchmarkCircuit `json:"circuit"`
	Samples []GasBenchmarkSample `json:"samples"`
}

type GasBenchmarkProfile struct {
	FormatVersion int                       `json:"formatVersion"`
	ID            string                    `json:"id"`
	Deployment    GasBenchmarkDeployment    `json:"deployment"`
	Source        GasBenchmarkSource        `json:"source"`
	Measurements  []GasBenchmarkMeasurement `json:"measurements"`
}

type GasEstimateRequest struct {
	Deployment GasBenchmarkDeployment
	Shape      GasBenchmarkShape
	Circuit    *GasBenchmarkCircuit
}
