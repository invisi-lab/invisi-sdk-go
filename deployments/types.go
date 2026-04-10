// Package deployments provides a static registry of Invisible Protocol contract
// deployments across supported chains, plus helpers for resolving per-operator
// InvisiProtocol instance addresses from the InvisiFactory on-chain.
package deployments

import "github.com/ethereum/go-ethereum/common"

// ContractDeployment holds the address and deployment metadata of a single contract.
type ContractDeployment struct {
	Address common.Address
	Block   uint64
	TxHash  common.Hash
}

// ChainDeployment holds all protocol contract addresses for a given chain.
// Addresses are zero-valued for chains that have not yet been deployed to.
type ChainDeployment struct {
	ChainID                 int64
	Network                 string
	Env                     string
	DeployedAt              string
	InvisiFactory           ContractDeployment
	VerifierRouter          ContractDeployment
	SingleSettleVerifier    ContractDeployment
	BatchSettleVerifier     ContractDeployment
	ShieldedPaymentVerifier ContractDeployment
	ProofLib                ContractDeployment
	CreationCode            []byte
	CreationCodeHash        common.Hash
}
