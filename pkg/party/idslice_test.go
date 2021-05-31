package party

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taurusgroup/cmp-ecdsa/pkg/math/curve"
)

func TestIDSlice_Lagrange(t *testing.T) {
	N := 10
	//T := 8
	allIDs := RandomPartyIDs(N)
	coefs := make([]*curve.Scalar, N)
	for i, id := range allIDs {
		coefs[i] = allIDs.Lagrange(id)
	}
	sum := curve.NewScalar()
	for _, c := range coefs {
		sum.Add(sum, c)
	}
	assert.True(t, sum.Equal(curve.NewScalar().SetInt64(1)))
}