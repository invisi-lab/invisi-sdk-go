package fmd

// Decaf377 encoding follows decaf377 0.10.1 min_curve/element.rs exactly.
// The underlying Edwards arithmetic and field square roots use gnark-crypto's
// BLS12-377 twisted Edwards implementation, with the Rust Decaf generator.
// The codec composition is unaudited; scalar multiplication and big.Int do not
// offer a constant-time guarantee. It is not a hardened secret-key backend.
import (
	"fmt"
	"math/big"

	decafq "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/twistededwards"
)

var decafScalarOrder = decafMustInteger("2111115437357092606062206234695386632838870926408408195193685246394721360383")
var decafGenerator = func() twistededwards.PointAffine {
	// gnark extended multiplication assumes curve parameters were initialized.
	_ = twistededwards.GetEdwardsCurve()
	var p twistededwards.PointAffine
	p.X.SetString("4959445789346820725352484487855828915252512307947624787834978378872129235627")
	p.Y.SetString("6060471950081851567114691557659790004756535011754163002297540472747064943288")
	return p
}()

func decafMustInteger(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid internal Decaf constant")
	}
	return n
}
func decafReadLE(input []byte) *big.Int {
	rev := make([]byte, len(input))
	for i, v := range input {
		rev[len(input)-1-i] = v
	}
	return new(big.Int).SetBytes(rev)
}
func decafWriteLE(n *big.Int) [32]byte {
	var b [32]byte
	n.FillBytes(b[:])
	for i := 0; i < 16; i++ {
		b[i], b[31-i] = b[31-i], b[i]
	}
	return b
}
func decafEncodeField(n decafq.Element) [32]byte {
	b := n.Bytes()
	for i := 0; i < 16; i++ {
		b[i], b[31-i] = b[31-i], b[i]
	}
	return b
}
func decafNegative(n *decafq.Element) bool { return n.Bits()[0]&1 != 0 }
func decafAbs(n decafq.Element) decafq.Element {
	if decafNegative(&n) {
		n.Neg(&n)
	}
	return n
}
func decafDecode(input []byte) (twistededwards.PointAffine, error) {
	var p twistededwards.PointAffine
	if len(input) != 32 || input[31]>>5 != 0 || input[0]&1 != 0 {
		return p, fmt.Errorf("invalid canonical Decaf377 encoding")
	}
	var canonical [32]byte
	for i, v := range input {
		canonical[31-i] = v
	}
	var s, ss, u1, u2, v, tmp, one, twoSU1 decafq.Element
	if err := s.SetBytesCanonical(canonical[:]); err != nil {
		return p, fmt.Errorf("invalid canonical Decaf377 field: %w", err)
	}
	one.SetOne()
	ss.Square(&s)
	u1.Sub(&one, &ss)
	u2.Square(&u1)
	tmp.SetUint64(12084)
	tmp.Mul(&tmp, &ss)
	u2.Sub(&u2, &tmp)
	tmp.Square(&u1)
	tmp.Mul(&tmp, &u2)
	if tmp.IsZero() {
		return p, fmt.Errorf("invalid Decaf377 square root")
	}
	tmp.Inverse(&tmp)
	if v.Sqrt(&tmp) == nil {
		return p, fmt.Errorf("invalid Decaf377 square root")
	}
	twoSU1.Double(&s)
	twoSU1.Mul(&twoSU1, &u1)
	tmp.Mul(&twoSU1, &v)
	if decafNegative(&tmp) {
		v.Neg(&v)
	}
	tmp.Square(&v)
	tmp.Mul(&tmp, &u2)
	p.X.Mul(&twoSU1, &tmp)
	tmp.Add(&one, &ss)
	tmp.Mul(&tmp, &v)
	p.Y.Mul(&tmp, &u1)
	if !p.IsOnCurve() {
		return twistededwards.PointAffine{}, fmt.Errorf("invalid Decaf377 point")
	}
	return p, nil
}
func decafEncode(p *twistededwards.PointAffine) ([32]byte, error) {
	if p.X.IsZero() {
		return [32]byte{}, nil
	}
	var t, u1, u2, u3, v, tmp, amd decafq.Element
	t.Mul(&p.X, &p.Y)
	u1.Add(&p.X, &t)
	tmp.Sub(&p.X, &t)
	u1.Mul(&u1, &tmp)
	amd.SetInt64(-3022)
	tmp.Square(&p.X)
	tmp.Mul(&tmp, &amd)
	tmp.Mul(&tmp, &u1)
	if tmp.IsZero() {
		return [32]byte{}, fmt.Errorf("degenerate Decaf377 point")
	}
	tmp.Inverse(&tmp)
	if v.Sqrt(&tmp) == nil {
		return [32]byte{}, fmt.Errorf("invalid Decaf377 point square root")
	}
	u2.Mul(&v, &u1)
	u2 = decafAbs(u2)
	u3.Sub(&u2, &t)
	tmp.Mul(&amd, &v)
	tmp.Mul(&tmp, &u3)
	tmp.Mul(&tmp, &p.X)
	tmp = decafAbs(tmp)
	return decafEncodeField(tmp), nil
}
func decafMultiply(p *twistededwards.PointAffine, scalar *big.Int) twistededwards.PointAffine {
	var out twistededwards.PointAffine
	out.ScalarMultiplication(p, scalar)
	return out
}
