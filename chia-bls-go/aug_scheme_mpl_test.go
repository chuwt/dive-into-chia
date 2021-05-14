package bls

import (
	"encoding/hex"
	"testing"
)

func TestAugSchemeMplTest(t *testing.T) {
	asm := new(AugSchemeMPL)

	sk := KeyGen(testSeed)

	sign := asm.Sign(sk, []byte("chuwt"))
	t.Log("signedMsg:", hex.EncodeToString(sign))

	t.Log("verify:", asm.Verify(sk.GetPublicKey(), []byte("chuwt"), sign))
}
