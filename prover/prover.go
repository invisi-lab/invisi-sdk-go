package prover

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type ProofStatement struct {
	CircuitName, Target, ArtifactSHA256, VKSHA256 string
	ExpectedPublicInputs                          []common.Hash
}

type ProofBundle struct {
	FormatVersion                                            int
	CircuitName, Target, BBVersion, ArtifactSHA256, VKSHA256 string
	Proof                                                    []byte
	ProofSHA256                                              string
	PublicInputs                                             []common.Hash
}

type Progress struct{ Phase string }

type Options struct{ OnProgress func(Progress) }

type Prover interface {
	Prove(context.Context, ProvingRequest, Options) (ProofBundle, error)
	Verify(context.Context, ProofBundle, ProofStatement, Options) (bool, error)
}
