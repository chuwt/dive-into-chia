package bls

import (
	"fmt"
	bls12381 "github.com/kilic/bls12-381"
)

var (
	AugSchemeDst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_AUG_")
)

type AugSchemeMPL struct{}

func (asm *AugSchemeMPL) Sign(sk PrivateKey, message []byte) []byte {
	pk := sk.GetPublicKey()

	newMessage := append(pk.ToBytes(), message...)

	g2 := coreSignMpl(sk.ToBytes(), newMessage, AugSchemeDst)

	return bls12381.NewG2().ToCompressed(g2)
}

func (asm *AugSchemeMPL) Verify(pk PublicKey, message []byte, sig []byte) bool {
	return coreVerifyMpl(
		pk,
		append(pk.ToBytes(), message...),
		sig,
		AugSchemeDst,
	)
}

func (asm *AugSchemeMPL) Aggregate() {

}

func (asm *AugSchemeMPL) AggregateVerify() {

}

func coreSignMpl(sk, message, dst []byte) *bls12381.PointG2 {
	g2Map := bls12381.NewG2()

	q, _ := g2Map.HashToCurve(message, dst)

	return g2Map.MulScalar(g2Map.New(), q, bls12381.NewFr().FromBytes(sk))
}

func coreVerifyMpl(pk PublicKey, message []byte, sig, dst []byte) bool {

	g2Map := bls12381.NewG2()
	q, _ := g2Map.HashToCurve(message, dst)

	// 校验
	signature, err := bls12381.NewG2().FromCompressed(sig)
	if err != nil {
		fmt.Println(len(sig))
		fmt.Println(err)
		return false
	}

	engine := bls12381.NewEngine()

	g1Neg := new(bls12381.PointG1)
	g1Neg = bls12381.NewG1().Neg(g1Neg, G1Generator())

	engine = engine.AddPair(pk.ToG1(), q)
	engine = engine.AddPair(g1Neg, signature)

	return engine.Check()
}

func coreAggregate() {

}
