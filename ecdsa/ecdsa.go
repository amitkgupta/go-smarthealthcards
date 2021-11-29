package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

func LoadKey(d, x, y string) (*ecdsa.PrivateKey, error) {
	dInt := new(big.Int)
	if err := dInt.UnmarshalText([]byte(d)); err != nil {
		return nil, err
	}

	xInt := new(big.Int)
	if err := xInt.UnmarshalText([]byte(x)); err != nil {
		return nil, err
	}

	yInt := new(big.Int)
	if err := yInt.UnmarshalText([]byte(y)); err != nil {
		return nil, err
	}

	pkey := ecdsa.PrivateKey{
		D: dInt,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     xInt,
			Y:     yInt,
		},
	}

	return &pkey, nil
}
