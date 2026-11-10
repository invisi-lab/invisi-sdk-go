package protocol

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/broadcaster"
	"github.com/invisi-lab/invisi-sdk-go/provenance"
	"math/big"
	"reflect"
)

// ValidateProvenanceProof 将公开恢复包重新解码成原生语句，防止数据与证明分离。
func ValidateProvenanceProof(packet provenance.ProvenanceProof, trust provenance.TrustContext) error {
	if packet.Format != "invisi-provenance-submission-v1" || len(packet.Sidecars) != 1 || !sameSourceNumber(packet.Context.ChainID, trust.ChainID) || !sameSourceNumber(packet.Context.ProtocolAddress, trust.ProtocolAddress) {
		return errors.New("provenance packet deployment or shape mismatch")
	}
	chain, e := provenance.Uint(packet.Context.ChainID, 256)
	if e != nil {
		return e
	}
	if !common.IsHexAddress(packet.Context.ProtocolAddress) || !common.IsHexAddress(packet.Context.TokenAddress) {
		return errors.New("invalid provenance packet context")
	}
	data := common.FromHex(packet.Data)
	if len(packet.Data) != 2+len(data)*2 {
		return errors.New("invalid provenance packet calldata")
	}
	intent, e := broadcaster.DecodeIntent(data)
	if e != nil {
		return e
	}
	if intent.Method != "transact" {
		return errors.New("provenance requires private transact calldata")
	}
	contract, e := ProtocolABI()
	if e != nil {
		return e
	}
	m, e := contract.MethodById(data[:4])
	if e != nil {
		return e
	}
	args, e := m.Inputs.Unpack(data[4:])
	if e != nil {
		return e
	}
	request := *abi.ConvertType(args[1], new(TransactRequest)).(*TransactRequest)
	c := TransactionContext{ChainID: chain, ProtocolAddress: common.HexToAddress(packet.Context.ProtocolAddress), TokenAddress: common.HexToAddress(packet.Context.TokenAddress)}
	normalized, e := NormalizeTransactRequest(c, request.Params, request.Deliveries, request.Funding)
	if e != nil {
		return e
	}
	words, e := TransactUserPublicInputs(c, request.Params, request.Deliveries, request.Funding)
	if e != nil {
		return e
	}
	values := make([]string, len(words))
	for i, w := range words {
		values[i] = common.Hash(w).Hex()
	}
	if len(request.Params.Nullifiers) == 0 || normalized.Capacity != packet.MainCircuitCapacity || !reflect.DeepEqual(values, packet.PublicInputs) {
		return errors.New("provenance native statement mismatch")
	}
	digest, e := provenance.OperationDigest(trust.Scope, normalized.Capacity, values)
	if e != nil {
		return e
	}
	if packet.OperationDigest != digest {
		return errors.New("provenance operation digest mismatch")
	}
	sidecar := packet.Sidecars[0]
	if sidecar.OperationIndex != "0" {
		return errors.New("unsupported provenance operation index")
	}
	if e = provenance.ValidateProofBundle(sidecar.ProofBundle, trust); e != nil {
		return e
	}
	nullifiers := make([]string, len(request.Params.Nullifiers))
	for i, n := range request.Params.Nullifiers {
		nullifiers[i] = common.Hash(n).Hex()
	}
	input, e := provenance.InputDigest(nullifiers)
	if e != nil {
		return e
	}
	pi := sidecar.ProofBundle.PublicInputs
	hi := new(big.Int).Rsh(common.HexToHash(digest).Big(), 128)
	lo := new(big.Int).And(common.HexToHash(digest).Big(), new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	if !sameSourceNumber(pi[2], c.TokenAddress.Hex()) || !sameSourceNumber(pi[6], new(big.Int).SetInt64(int64(len(nullifiers))).String()) || !sameSourceNumber(pi[7], hi.String()) || !sameSourceNumber(pi[8], lo.String()) || !sameSourceNumber(pi[9], input) {
		return errors.New("provenance proof statement mismatch")
	}
	return nil
}
func sameSourceNumber(a, b string) bool {
	x, e := provenance.Uint(a, 256)
	y, f := provenance.Uint(b, 256)
	return e == nil && f == nil && x.Cmp(y) == 0
}
func SubmitProvenanceProof(ctx context.Context, client *provenance.Client, packet provenance.ProvenanceProof) (provenance.Result, error) {
	if client == nil {
		return nil, errors.New("provenance client required")
	}
	if e := ValidateProvenanceProof(packet, client.Trust); e != nil {
		return nil, e
	}
	return client.SubmitSourceProof(ctx, packet.Data, packet.Sidecars)
}
