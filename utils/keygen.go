package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

func main() {
	pkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	d, err := pkey.D.MarshalText()
	if err != nil {
		panic(err)
	}

	x, err := pkey.PublicKey.X.MarshalText()
	if err != nil {
		panic(err)
	}

	y, err := pkey.PublicKey.Y.MarshalText()
	if err != nil {
		panic(err)
	}

	fmt.Printf("export SMART_HEALTH_CARDS_KEY_D=%s\n", string(d))
	fmt.Printf("export SMART_HEALTH_CARDS_KEY_X=%s\n", string(x))
	fmt.Printf("export SMART_HEALTH_CARDS_KEY_Y=%s\n", string(y))
}
