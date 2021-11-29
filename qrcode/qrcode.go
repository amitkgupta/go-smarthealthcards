package qrcode

import (
	"errors"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

var JWSTooLargeError = errors.New("JWS too large, QR chunking not currently implemented.")

func Encode(content string) ([]byte, error) {
	if len(content) > 1195 {
		return nil, JWSTooLargeError
	}

	shcContent := "shc:/"
	for _, r := range content {
		shcContent += fmt.Sprintf("%02d", r-45)
	}

	q, err := qrcode.NewWithForcedVersion(shcContent, 22, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	return q.PNG(512)
}
