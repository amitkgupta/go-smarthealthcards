package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

func GenerateKey() {
	pkey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	d, _ := pkey.D.MarshalText()
	x, _ := pkey.PublicKey.X.MarshalText()
	y, _ := pkey.PublicKey.Y.MarshalText()

	println(string(d), string(x), string(y))

	D := new(big.Int)
	if err := D.UnmarshalText(d); err != nil {
		panic(err)
	}

	X := new(big.Int)
	if err := X.UnmarshalText(x); err != nil {
		panic(err)
	}

	Y := new(big.Int)
	if err := Y.UnmarshalText(y); err != nil {
		panic(err)
	}

	PKEY := ecdsa.PrivateKey{
		D: D,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     X,
			Y:     Y,
		},
	}

	println(pkey.Equal(&PKEY))
}
