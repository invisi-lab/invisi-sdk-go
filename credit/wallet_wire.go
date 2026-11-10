package credit

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	wire "github.com/invisi-lab/invisi-sdk-go/wallet/protocol"
)

func (p *Prepared) ExportWalletRequest(id string) (wire.Request, error) {
	request, err := p.AuthorizationRequest()
	if err != nil {
		return wire.Request{}, err
	}
	p.mu.Lock()
	var identity fr.Element
	switch {
	case p.intent.Create != nil:
		identity = primitives.BigIntToElement(p.intent.Create.Owner.Big())
	case p.intent.Update != nil:
		identity = primitives.BigIntToElement(p.intent.Update.Owner.Big())
	case p.intent.Draw != nil && len(p.intent.Inputs) > 0:
		identity = p.intent.Inputs[0].State.SpenderWalletIdentity
	default:
		p.mu.Unlock()
		return wire.Request{}, fmt.Errorf("Credit wallet identity unavailable")
	}
	p.mu.Unlock()
	return wire.Prepare(id, common.BigToAddress(primitives.ElementToBigInt(identity)), wire.Authorize, wire.Authorization(request))
}
func (p *Prepared) WithWalletResponse(q wire.Request, r wire.Response) error {
	witness, err := wire.AcceptAuthorization(q, r)
	if err != nil {
		return err
	}
	public := append([]byte{4}, witness.PublicKeyX[:]...)
	public = append(public, witness.PublicKeyY[:]...)
	// WithAuthority rechecks the current frozen statement, including broadcaster repricing.
	return p.WithAuthority(Authority{PublicKey: public, Signature: witness.Signature[:]})
}
