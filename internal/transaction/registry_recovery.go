package transaction

import (
	"bytes"
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/invisi-labs/invisi-sdk-go/protocol"
	"github.com/invisi-labs/invisi-sdk-go/rpc/contracts"
)

// RegistrationSnapshot holds public data for confirming a submitted registration after a restart.
type RegistrationSnapshot struct {
	ChainID         string
	RegistryAddress common.Address
	TransactionHash common.Hash
	Calldata        []byte
}

func (o *RegistrationOperation) ConfirmationSnapshot() (RegistrationSnapshot, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.transactionHash == (common.Hash{}) {
		return RegistrationSnapshot{}, errors.New("registration is not tracked")
	}
	return RegistrationSnapshot{
		ChainID: o.context.ChainID.String(), RegistryAddress: o.context.RegistryAddress,
		TransactionHash: o.transactionHash, Calldata: bytes.Clone(o.data),
	}, nil
}

// RestoreRegistrationConfirmation verifies the original submission without preparing a new call.
func (r *AddressRegistry) RestoreRegistrationConfirmation(ctx context.Context, wallet *Wallet, snapshot RegistrationSnapshot) (*RegistrationOperation, error) {
	if wallet == nil || snapshot.ChainID != r.context.ChainID.String() ||
		snapshot.RegistryAddress != r.context.RegistryAddress || snapshot.TransactionHash == (common.Hash{}) {
		return nil, errors.New("invalid registration snapshot deployment, wallet or transaction")
	}
	if err := r.ValidateDeployment(ctx); err != nil {
		return nil, err
	}
	parsed, err := contracts.InvisiAddressRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	method, err := parsed.MethodById(snapshot.Calldata)
	if err != nil || (method.Name != "registerKeys" && method.Name != "updateKeys") {
		return nil, errors.New("invalid registration snapshot calldata")
	}
	values, err := method.Inputs.Unpack(snapshot.Calldata[4:])
	if err != nil {
		return nil, err
	}
	request := abi.ConvertType(values[0], new(contracts.IAddressRegistryRegistration)).(*contracts.IAddressRegistryRegistration)
	signature := values[1].([]byte)
	if new(big.Int).SetBytes(request.WalletIdentity[:]).Cmp(wallet.ownerAddress.Big()) != 0 {
		return nil, errors.New("registration snapshot owner mismatch")
	}
	registration := protocol.AddressRegistration{
		OwnerAddress: request.WalletIdentity, Keys: request.Keys, Nonce: request.Nonce, Deadline: request.Deadline,
	}
	update := method.Name == "updateKeys"
	call, err := protocol.BuildRegistrationCalldata(r.context, registration, signature, protocol.RegistrationOptions{Update: update})
	if err != nil || !bytes.Equal(snapshot.Calldata, call.Data) {
		return nil, errors.New("registration snapshot calldata mismatch")
	}
	operation := &RegistrationOperation{
		lifecycle: &TransactionLifecycle{}, registry: r, wallet: wallet,
		options: RegistrationPreparation{Update: update}, registration: registration, context: r.context,
		status: "uncertain", data: call.Data, transactionHash: snapshot.TransactionHash, direct: len(signature) == 0,
	}
	return operation.Reconcile(ctx)
}
