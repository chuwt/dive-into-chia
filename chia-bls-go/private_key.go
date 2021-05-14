package bls

import (
	bls12381 "github.com/kilic/bls12-381"
	"math/big"
)

type PrivateKey struct {
	value *big.Int
}

func (pk PrivateKey) GetPublicKey() PublicKey {
	g1 := bls12381.NewG1()
	return PublicKey{
		value: g1.MulScalar(g1.New(), G1Generator(), bls12381.NewFr().FromBytes(pk.value.Bytes())),
	}
}

func (pk PrivateKey) ToBytes() []byte {
	return pk.value.Bytes()
}

func (pk PrivateKey) ToFarmerSk() PrivateKey {
	return derivePath(pk, []int{12381, 8444, 0, 0})
}

func (pk PrivateKey) ToPoolSk() PrivateKey {
	return derivePath(pk, []int{12381, 8444, 1, 0})
}

func (pk PrivateKey) ToWalletSk() PrivateKey {
	return derivePath(pk, []int{12381, 8444, 2, 0})
}

func (pk PrivateKey) ToLocalSk() PrivateKey {
	return derivePath(pk, []int{12381, 8444, 3, 0})
}
