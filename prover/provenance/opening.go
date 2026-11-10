package provenance

type Opening struct {
	Owner  string `json:"owner"`
	Nonce  string `json:"nonce"`
	Amount string `json:"amount"`
}

func NoteOpening(s Scope, token string, o Opening) (commitment, nullifier string, err error) {
	for _, v := range []struct {
		s string
		b int
		n bool
	}{{s.ChainID, 64, false}, {s.ProtocolAddress, 160, true}, {token, 160, true}, {o.Owner, 160, true}, {o.Amount, 128, true}} {
		if e := check(v.s, v.b, v.n); e != nil {
			return "", "", e
		}
	}
	id, e := hash("2", o.Owner, o.Nonce)
	if e != nil {
		return "", "", e
	}
	c, e := hash("3", s.ChainID, s.ProtocolAddress, id, token, o.Amount)
	if e != nil {
		return "", "", e
	}
	n, e := hash("4", s.ChainID, s.ProtocolAddress, o.Nonce, c)
	return c, n, e
}
