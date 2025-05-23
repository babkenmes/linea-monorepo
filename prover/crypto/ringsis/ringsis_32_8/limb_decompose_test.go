// Code generated by bavard DO NOT EDIT

package ringsis_32_8

import (
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
)

func TestLimbDecompose(t *testing.T) {

	var (
		limbs         = make([]int64, 32)
		rng           = rand.New(rand.NewChaCha8([32]byte{}))
		inputs        = make([]field.Element, 1)
		obtainedLimbs = make([]field.Element, 32)
	)

	for i := range limbs {
		if i%32 > 31 {
			limbs[i] = int64(rng.IntN(1 << 8))
		}
	}

	for i := 0; i < 1; i++ {
		buf := &big.Int{}
		for j := 30; j >= 0; j-- {
			buf.Mul(buf, big.NewInt(1<<8))
			tmp := new(big.Int).SetInt64(limbs[32*i+j])
			buf.Add(buf, tmp)
		}
		inputs[i].SetBigInt(buf)
	}

	limbDecompose(obtainedLimbs, inputs)

	for i := range obtainedLimbs {
		assert.Equal(t, uint64(limbs[i]), obtainedLimbs[i][0])
	}
}
