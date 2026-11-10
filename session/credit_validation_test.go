package session

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
	"github.com/invisi-lab/invisi-sdk-go/rpc/contracts"
	"math/big"
	"testing"
)

type creditValidationProvider struct {
	sessionProvider
	revoked   bool
	allowance *big.Int
}

func (p creditValidationProvider) CallContract(ctx context.Context, message ethereum.CallMsg, block *big.Int) ([]byte, error) {
	a, _ := protocol.ProtocolABI()
	if len(message.Data) >= 4 {
		if method, e := a.MethodById(message.Data[:4]); e == nil {
			switch method.RawName {
			case "creditAllocation":
				return method.Outputs.Pack(uint64(1), uint16(1))
			case "creditRevocationBitmap":
				bits := new(big.Int)
				if p.revoked {
					bits.SetUint64(1)
				}
				return method.Outputs.Pack(bits)
			}
		}
	}
	erc20, _ := contracts.IERC20MetaData.GetAbi()
	if method, e := erc20.MethodById(message.Data[:4]); e == nil {
		switch method.RawName {
		case "balanceOf":
			return method.Outputs.Pack(big.NewInt(1000))
		case "allowance":
			value := big.NewInt(1000)
			if p.allowance != nil {
				value = p.allowance
			}
			return method.Outputs.Pack(value)
		}
	}
	return p.sessionProvider.CallContract(ctx, message, block)
}
func TestManagedRevocationUsesOriginalReferenceAndCurrentFeeAllowance(t *testing.T) {
	ctx := context.Background()
	wallet, token, _ := testSession(t)
	base := wallet.options.Provider.(sessionProvider)
	wallet.options.Provider = creditValidationProvider{sessionProvider: base}
	c := credit.Context{ProtocolContext: wallet.coreContext(), TokenAddress: token, Timestamp: base.head.Time}
	p := credit.RevokeParams{Token: token, Owner: wallet.wallet.OwnerAddress(), Shard: 1, Slot: 0, Deadline: new(big.Int).SetUint64(base.head.Time + 300), FeeRecipient: common.HexToAddress("0x1234"), FeeAmount: big.NewInt(1)}
	prepared, err := credit.PrepareRevoke(c, p)
	if err != nil {
		t.Fatal(err)
	}
	managed, err := wallet.Credit()
	if err != nil {
		t.Fatal(err)
	}
	operation, err := managed.Prepare(ctx, "revoke", prepared)
	if err != nil {
		t.Fatal("revocation required latest merchant state", err)
	}
	if _, err = operation.Calldata(ctx, p.Owner); err != nil {
		t.Fatal(err)
	}
	wallet.options.Provider = creditValidationProvider{sessionProvider: base, revoked: true}
	if _, err = managed.Prepare(ctx, "revoked", prepared); err == nil {
		t.Fatal("accepted revoked authority")
	}
	wallet.options.Provider = creditValidationProvider{sessionProvider: base, allowance: big.NewInt(0)}
	if _, err = managed.Prepare(ctx, "unapproved-fee", prepared); err == nil {
		t.Fatal("accepted unapproved owner fee")
	}
	contract, err := contracts.NewInvisiProtocolCaller(wallet.scope.ProtocolAddress, wallet.options.Provider)
	if err != nil {
		t.Fatal(err)
	}
	allocation, err := contract.CreditAllocation(&bind.CallOpts{Context: ctx}, p.Owner, p.Token)
	if err != nil || allocation.Shard != 1 {
		t.Fatal(err)
	}
	wallet.options.Provider = creditValidationProvider{sessionProvider: base}
	if err = wallet.options.Store.Update(ctx, wallet.scope, func(state *WalletState) error {
		state.Invalidated = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = managed.Prepare(ctx, "invalidated", prepared); err == nil {
		t.Fatal("accepted authorization with invalidated wallet checkpoint")
	}
}
