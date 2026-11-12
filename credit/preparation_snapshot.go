package credit

import "math/big"

func snapshotPayload(payload []byte) []byte {
	if payload == nil {
		return nil
	}
	return append([]byte{}, payload...)
}
func snapshotAmount(amount *big.Int) *big.Int {
	if amount == nil {
		return nil
	}
	return new(big.Int).Set(amount)
}
func (f *Fee) clone() *Fee {
	if f == nil {
		return nil
	}
	return &Fee{Recipient: f.Recipient, Amount: snapshotAmount(f.Amount)}
}
func (p UpdatePatch) Clone() UpdatePatch {
	if p.SingleLimit != nil {
		value := *p.SingleLimit
		p.SingleLimit = &value
	}
	if p.DailyLimit != nil {
		value := *p.DailyLimit
		p.DailyLimit = &value
	}
	if p.MonthlyLimit != nil {
		value := *p.MonthlyLimit
		p.MonthlyLimit = &value
	}
	if p.Expiry != nil {
		value := *p.Expiry
		p.Expiry = &value
	}
	if p.Unlimited != nil {
		value := *p.Unlimited
		p.Unlimited = &value
	}
	p.Payload = snapshotPayload(p.Payload)
	return p
}
func (p CreatePreparation) Clone() CreatePreparation {
	p.AvailableCredit = snapshotAmount(p.AvailableCredit)
	p.Payload = snapshotPayload(p.Payload)
	p.Fee = p.Fee.clone()
	return p
}
func (p UpdatePreparation) Clone() UpdatePreparation {
	p.UpdatePatch = p.UpdatePatch.Clone()
	p.Fee = p.Fee.clone()
	return p
}
func (p DrawPreparation) Clone() DrawPreparation {
	p.Payload = snapshotPayload(p.Payload)
	p.Fee = p.Fee.clone()
	if p.Credits != nil {
		members := make([]DrawMember, len(p.Credits))
		for i, member := range p.Credits {
			member.Amount = snapshotAmount(member.Amount)
			member.Payload = snapshotPayload(member.Payload)
			members[i] = member
		}
		p.Credits = members
	}
	return p
}
func (p RevokePreparation) Clone() RevokePreparation { p.Fee = p.Fee.clone(); return p }
