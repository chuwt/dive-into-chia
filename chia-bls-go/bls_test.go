package bls

import (
	"encoding/hex"
	"testing"
)

var testSeed, _ = hex.DecodeString("" +
	"76f3109f1a2142fdefcc6a666d3f321b37ce9690d28103ccbb1a654af2c0a469" +
	"00aac14fab9f0ce4851cf1a1fe8beaf7d34c9ceb008849d5b7e9bc78ef0ec649")

/*
Fingerprint: 563730848

Farmer public key (m/12381/8444/0/0): b69b74794fa16c4569af42401615948094ad076795627d88f08e5f1626ec3e2dda47376db481dd3ecdf0585b960b80cf

Pool public key (m/12381/8444/1/0): 8b417b4310ecb7fd68e8c39e0fa0e334edd3c8c93eca9985a3f398846f9429142993196416199436718f3ec26609e618
*/

func TestBls(t *testing.T) {
	masterSk := KeyGen(testSeed)
	masterPk := masterSk.GetPublicKey()

	t.Log("masterSk:", hex.EncodeToString(masterSk.ToBytes()))
	t.Log("masterSk:", hex.EncodeToString(masterPk.ToBytes()))
	t.Log("fingerPrint:", masterPk.FingerPrint())

	t.Log("")

	farmerSk := masterSk.ToFarmerSk()
	farmerPk := farmerSk.GetPublicKey()
	t.Log("farmerSk:", hex.EncodeToString(farmerSk.ToBytes()))
	t.Log("farmerPk:", hex.EncodeToString(farmerPk.ToBytes()))

	t.Log("")

	poolSk := masterSk.ToPoolSk()
	poolPk := poolSk.GetPublicKey()
	t.Log("poolSk:", hex.EncodeToString(poolSk.ToBytes()))
	t.Log("poolPk:", hex.EncodeToString(poolPk.ToBytes()))
}
