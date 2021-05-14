package bls

import (
	bls12381 "github.com/kilic/bls12-381"
	"math/big"
)

type PublicKey struct {
	value *bls12381.PointG1
}

func (pk *PublicKey) FingerPrint() string {
	return new(big.Int).SetBytes(Hash256(bls12381.NewG1().ToCompressed(pk.value))[:4]).String()
}

func (pk *PublicKey) ToBytes() []byte {
	return bls12381.NewG1().ToCompressed(pk.value)
}

func (pk *PublicKey) ToG1() *bls12381.PointG1 {
	return pk.value
}
