// Package qrcode creates one or more QR codes in PNG format encoding the JWS
// of a SMART Health Card such that smart devices such as iPhones can scan the
// QR code(s) and load the SMART Health Card information in applications such
// as the Wallet and Health apps for the iPhone. See
// https://spec.smarthealth.cards/#every-health-card-can-be-embedded-in-a-qr-code
// and
// https://spec.smarthealth.cards/#encoding-chunks-as-qr-codes.
package qrcode

import (
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

const maxSingleChunkSize = 1195 // https://spec.smarthealth.cards/#chunking
const maxMultipleChunkSize = 1191

// Encode takes the content to be encoded, breaks it into chunks if necessary,
// and encodes each chunk as per the SMART Health Card spec, see:
// https://spec.smarthealth.cards/#encoding-chunks-as-qr-codes.
//
// Each encoded chunk is then encoded as a QR code in PNG format and
// represented as a byte slice.
func Encode(content string) ([][]byte, error) {
	numChunks := 1
	if len(content) > maxSingleChunkSize {
		if len(content)%maxMultipleChunkSize == 0 {
			numChunks = len(content) / maxMultipleChunkSize
		} else {
			numChunks = (len(content) / maxMultipleChunkSize) + 1
		}
	}

	pngs := make([][]byte, numChunks)
	for i := 1; i <= numChunks; i++ {
		var err error
		if pngs[i-1], err = shcContent(i, numChunks, content[(i-1)*len(content)/numChunks:i*len(content)/numChunks]); err != nil {
			return nil, err
		}
	}
	return pngs, nil
}

func shcContent(c int, n int, content string) ([]byte, error) {
	shcContent := "shc:/"

	if n != 1 {
		shcContent += fmt.Sprintf("%d/%d/", c, n)
	}

	for _, r := range content {
		shcContent += fmt.Sprintf("%02d", r-45)
	}

	q, err := qrcode.NewWithForcedVersion(shcContent, 22, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	return q.PNG(512)
}
