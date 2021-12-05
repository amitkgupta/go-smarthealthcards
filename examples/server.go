package main

import (
	"log"
	"net/http"
	"os"

	"github.com/amitkgupta/go-smarthealthcards/v2/ecdsa"
	"github.com/amitkgupta/go-smarthealthcards/v2/webhandlers"
)

// This example shows how to load an ECDSA private key from string
// environment variables, and use that to run a web server that
// issues SMART Health Card QR codes based on user form input and
// presents public information of the private key at
// /.well-known/jwks.json so that devices which interpret the SMART
// Health Card data in the QR codes can verify them against the issuer.
//
// This example uses "https://example.com" as the issuer, so this server
// would need to be reachable at that address serving a valid TLS
// certificate for "example.com".
func ExampleServer() {
	shcKey, err := ecdsa.LoadKey(
		os.Getenv("SMART_HEALTH_CARDS_KEY_D"),
		os.Getenv("SMART_HEALTH_CARDS_KEY_X"),
		os.Getenv("SMART_HEALTH_CARDS_KEY_Y"),
	)
	if err != nil {
		log.Fatal(err)
	}

	shcWebHandlers := webhandlers.New(shcKey, "https://example.com")

	log.Fatal(http.ListenAndServe(
		":8080",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				if responseCode, errorMessage, ok := shcWebHandlers.ProcessForm(w, r); !ok {
					http.Error(w, errorMessage, responseCode)
				}
			case http.MethodGet:
				if responseCode, errorMessage, ok := shcWebHandlers.JWKSJSON(w); !ok {
					http.Error(w, errorMessage, responseCode)
				}
			}
		}),
	))
}

func main() { ExampleServer() }
