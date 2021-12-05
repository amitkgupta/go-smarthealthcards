// Package ecdsa loads an ECDSA P-256 private key (*crypto/ecdsa.PrivateKey)
// from string representations of its key parameters. See
// https://spec.smarthealth.cards/#generating-and-resolving-cryptographic-keys.
package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

// LoadKey takes string representations of the d, x, and y
// paramters of an ECDSA key, and loads them as *math/big.Int
// objects using the (*math/big.Int).UnmarshalText method.
// Then it return an ECDSA private key of type
// *crypto/ecdsa.PrivateKey.
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
