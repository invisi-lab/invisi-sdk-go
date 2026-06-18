package deployments

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	contracts "github.com/invisi-lab/invisi-sdk-go/contracts"
)

// ResolveInstanceAddress calls InvisiFactory.getInstanceAddress on-chain to
// retrieve the deployed InvisiProtocol instance address for a given operator
// and token pair.
func ResolveInstanceAddress(
	client *ethclient.Client,
	factoryAddress common.Address,
	operator common.Address,
	token common.Address,
) (common.Address, error) {
	if client == nil {
		return common.Address{}, fmt.Errorf("client is nil")
	}
	if factoryAddress == (common.Address{}) {
		return common.Address{}, fmt.Errorf("factory address is zero")
	}

	factory, err := contracts.NewInvisiFactory(factoryAddress, client)
	if err != nil {
		return common.Address{}, fmt.Errorf("binding factory: %w", err)
	}

	addr, err := factory.GetInstanceAddress(&bind.CallOpts{}, operator, token)
	if err != nil {
		return common.Address{}, fmt.Errorf("calling getInstanceAddress: %w", err)
	}

	return addr, nil
}
