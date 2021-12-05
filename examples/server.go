package main

import (
	"log"
	"net/http"
	"os"

	"github.com/amitkgupta/go-smarthealthcards/v2/ecdsa"
	"github.com/amitkgupta/go-smarthealthcards/v2/webhandlers"
)

func main() {
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
