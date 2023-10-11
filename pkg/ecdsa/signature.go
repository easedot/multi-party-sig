package ecdsa

import (
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"log"
)

type Signature struct {
	R curve.Point
	S curve.Scalar
}

// EmptySignature returns a new signature with a given curve, ready to be unmarshalled.
func EmptySignature(group curve.Curve) Signature {
	return Signature{R: group.NewPoint(), S: group.NewScalar()}
}

func (sig Signature) RecoverCode() byte {
	overflow := sig.R.XOverflow()
	pubKeyRecoveryCode := byte(overflow<<1) | byte(sig.R.YOddBit())
	log.Printf("RecoveryCode:%v", pubKeyRecoveryCode)
	if sig.S.IsOverHalfOrder() {
		//pubKeyRecoveryCode ^= 0x01
	}
	log.Printf("RecoveryCode1:%v", pubKeyRecoveryCode)
	return pubKeyRecoveryCode
}

// Verify is a custom signature format using curve data.
// aG = PublicKey
// r=Random
// sig.R =rG    sig.S=(h+ax)/r
func (sig Signature) Verify(aG curve.Point, hash []byte) bool {
	group := aG.Curve()

	sInv := group.NewScalar().Set(sig.S).Invert() //=1/s

	h := curve.FromHash(group, hash)
	hG := h.ActOnBase()

	x := sig.R.XScalar()
	xaG := x.Act(aG)

	rG := sInv.Act(hG.Add(xaG)) // (hG+xaG)/s

	return rG.Equal(sig.R)
}

//// Verify is a custom signature format using curve data.
//func (sig Signature) Verify(X curve.Point, hash []byte) bool {
//	group := X.Curve()
//
//	m := curve.FromHash(group, hash)
//	sInv := group.NewScalar().Set(sig.S).Invert()
//
//	mG := m.ActOnBase()
//
//	r := sig.R.XScalar()
//	rX := r.Act(X)
//	R2 := mG.Add(rX)
//
//	R2 = sInv.Act(R2)
//
//	return R2.Equal(sig.R)
//}
