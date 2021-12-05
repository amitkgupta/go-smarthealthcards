// Package jws creates a compact serialization of a JSON Web Signature (JWS)
// with the ECDSA P-256 SHA-256 signing algorithm and DEFLATE compression of
// the payload and creates a serialization of a JSON Web Key Set representing
// the public key of an ECDSA P-256 key. See
// https://spec.smarthealth.cards/#health-cards-are-encoded-as-compact-serialization-json-web-signatures-jws,
// https://spec.smarthealth.cards/#health-cards-are-small,
// and
// https://spec.smarthealth.cards/#determining-keys-associated-with-an-issuer
package jws

import (
	"bytes"
	"compress/flate"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

const (
	algorithm = "ES256"
	curve     = "P-256"
	keyType   = "EC"
)

type header struct {
	Algorithm string `json:"alg"`
	Zip       string `json:"zip"`
	KeyID     string `json:"kid"`
}

// SignAndSerialize compresses the given payload, signs it with the given key,
// and returns the resulting enoded JSON Web Signature (JWS). See:
// https://datatracker.ietf.org/doc/html/rfc7515#appendix-A.3
func SignAndSerialize(payload []byte, key *ecdsa.PrivateKey) (string, error) {
	h := header{
		Algorithm: algorithm,
		Zip:       "DEF",
		KeyID:     kid(key),
	}

	hBytes, err := json.Marshal(&h)
	if err != nil {
		return "", err
	}

	hB64String := base64.RawURLEncoding.EncodeToString(hBytes)

	pBuf := new(bytes.Buffer)
	if zw, err := flate.NewWriter(pBuf, flate.BestCompression); err != nil {
		return "", err
	} else {
		if _, err = zw.Write(payload); err != nil {
			return "", err
		}
		if err = zw.Close(); err != nil {
			return "", err
		}
	}

	pB64String := base64.RawURLEncoding.EncodeToString(pBuf.Bytes())

	signingInput := []byte(hB64String + "." + pB64String)

	r, s, err := sign(key, signingInput)
	if err != nil {
		return "", err
	}

	sigB64String := base64.RawURLEncoding.EncodeToString(
		append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...),
	)

	return hB64String + "." + pB64String + "." + sigB64String, nil
}

func sign(key *ecdsa.PrivateKey, payload []byte) (*big.Int, *big.Int, error) {
	hash := make([]byte, 32)
	for i, b := range sha256.Sum256(payload) {
		hash[i] = b
	}
	return ecdsa.Sign(rand.Reader, key, hash)
}

func xtos(key *ecdsa.PrivateKey) string {
	return base64.RawURLEncoding.EncodeToString(key.PublicKey.X.FillBytes(make([]byte, 32)))
}

func ytos(key *ecdsa.PrivateKey) string {
	return base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.FillBytes(make([]byte, 32)))
}

func kid(key *ecdsa.PrivateKey) string {
	jwkString := fmt.Sprintf(
		`{"crv":"%s","kty":"%s","x":"%s","y":"%s"}`,
		curve,
		keyType,
		xtos(key),
		ytos(key),
	)

	hash := make([]byte, 32)
	for i, b := range sha256.Sum256([]byte(jwkString)) {
		hash[i] = b
	}

	return base64.RawURLEncoding.EncodeToString(hash)
}

// JWKSJSON takes an *crypto/ecdsa.PrivateKey and returns
// the JSON serialization of the JSON Web Key Set (JWKS)
// representing the unique publid identifying information
// of the private key.
func JWKSJSON(key *ecdsa.PrivateKey) ([]byte, error) {
	return json.Marshal(jwks{
		Keys: []jwk{
			{
				KeyType:   keyType,
				KeyID:     kid(key),
				Use:       "sig",
				Algorithm: algorithm,
				Curve:     curve,
				X:         xtos(key),
				Y:         ytos(key),
			},
		},
	})
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Curve     string `json:"crv"`
	X         string `json:"x"`
	Y         string `json:"y"`
}
