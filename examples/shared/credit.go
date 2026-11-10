package shared

import (
	"github.com/invisi-lab/invisi-sdk-go/credit"
	"github.com/invisi-lab/invisi-sdk-go/primitives"
	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func CreditDeliveries(snapshot credit.DeliveryState) ([]protocol.Delivery, error) {
	commitment, err := snapshot.State.Commitment()
	if err != nil {
		return nil, err
	}
	result := []protocol.Delivery{}
	for _, role := range []credit.Role{credit.Owner, credit.Spender} {
		identity := snapshot.State.OwnerWalletIdentity
		keys := snapshot.OwnerReceivingKeys
		if role == credit.Spender {
			identity = snapshot.State.SpenderWalletIdentity
			keys = snapshot.SpenderReceivingKeys
		}
		cipher, err := credit.Encrypt(snapshot, credit.DeliveryContext{ProtocolContext: snapshot.State.ProtocolContext, TokenAddress: snapshot.State.TokenAddress, Commitment: commitment, RecipientIdentity: identity}, role)
		if err != nil {
			return nil, err
		}
		clue, err := primitives.CreateClue(keys.ClueKey)
		if err != nil {
			return nil, err
		}
		result = append(result, protocol.Delivery{Ciphertext: cipher, Clue: clue})
	}
	return result, nil
}
