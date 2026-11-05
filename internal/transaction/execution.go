package transaction

import (
	"encoding/json"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/invisi-lab/invisi-sdk-go/core"
	"github.com/invisi-lab/invisi-sdk-go/prover"
)

type Execution struct {
	mu              sync.Mutex
	Runtime         *Runtime
	Plan            Plan
	Status          string
	Authorization   *core.WalletAuthorizationWitness
	Bundle          *prover.ProofBundle
	Data            []byte
	TransactionHash common.Hash
	WorkerID        string
}

func (o *Execution) ID() string { return o.Plan.ID }

func (o *Execution) String() string { return "Operation(" + o.Plan.ID + ",[private plan redacted])" }

func (o *Execution) GoString() string { return o.String() }

func (o *Execution) MarshalJSON() ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return json.Marshal(struct {
		ID     string `json:"id"`
		Kind   string `json:"kind"`
		Status string `json:"status"`
	}{o.Plan.ID, PublicOperationKind(o.Plan), o.Status})
}

func HashCode(code []byte) common.Hash { return crypto.Keccak256Hash(code) }

func ClonePlan(plan Plan) (Plan, error) {
	raw, e := json.Marshal(plan)
	if e != nil {
		return Plan{}, e
	}
	var result Plan
	e = json.Unmarshal(raw, &result)
	return result, e
}

type Confirmation struct {
	Status          string
	TransactionHash common.Hash
}
