package bls

import (
	bls12381 "github.com/kilic/bls12-381"
	"math/big"
)

func KeyGen(seed []byte) PrivateKey {
	L := 48
	okm := extractExpand(L, append(seed, 0), []byte("BLS-SIG-KEYGEN-SALT-"), []byte{0, byte(L)})

	return PrivateKey{new(big.Int).Mod(new(big.Int).SetBytes(okm), bls12381.NewG1().Q())}
}
