package jws

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"

	"github.com/amitkgupta/personal-website/smarthealthcards/ecdsa"
)

type header struct {
	Algorithm string `json:"alg"`
	Zip       string `json:"zip"`
	KeyID     string `json:"kid"`
}

func SignAndSerialize(payload []byte, key ecdsa.Key) (string, error) {
	h := header{
		Algorithm: "ES256",
		Zip:       "DEF",
		KeyID:     key.Kid(),
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

	r, s, err := key.Sign(signingInput)
	if err != nil {
		return "", err
	}

	sigB64String := base64.RawURLEncoding.EncodeToString(
		append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...),
	)

	return hB64String + "." + pB64String + "." + sigB64String, nil
}
