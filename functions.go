package ingester

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	convoy "github.com/frain-dev/convoy-go"
	convoyModels "github.com/frain-dev/convoy-go/models"
	"github.com/go-chi/chi/v5"
)

var (
	URL      = os.Getenv("CONVOY_URL")
	USERNAME = os.Getenv("CONVOY_USERNAME")
	PASSWORD = os.Getenv("CONVOY_PASSWORD")
)

func WebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	// Build Router.
	router := chi.NewRouter()

	router.Route("/v1", func(v1Router chi.Router) {

		v1Router.Post("/paystack", PaystackHandler)
		v1Router.Post("/mono", MonoHandler)
	})

	// Serve Request.
	router.ServeHTTP(w, r)
}

// HTTP Handlers

// PaystackHandler ack webhooks from https://paystack.com
// NEEDS:
// CONVOY_PAYSTACK_APP_ID
// PAYSTACK_SECRET
func PaystackHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Could not read payload"))
		return
	}

	var event PaystackEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		w.Write([]byte("Server Error: could not unmarshal payload"))
		return
	}

	// Verify Payload.
	verifierOpts := &VerifierOptions{
		Header:      "X-Paystack-Signature",
		Hash:        "SHA512",
		Secret:      os.Getenv("PAYSTACK_SECRET"),
		IPWhitelist: []string{"52.31.139.75", "52.49.173.169", "52.214.14.220"},
	}

	verifier := NewVerifier(verifierOpts)
	if err != nil {
		w.Write([]byte("Server Error: Could not create verifier"))
	}

	ok, err := verifier.VerifyRequest(r)
	if err != nil {
		w.Write([]byte("Server Error: Failed to verify request"))
		return
	}

	if !ok {
		w.Write([]byte("Bad Request: Could not verify request"))
		return
	}

	// Push to Convoy.
	// TODO(subomi): Build in some form of retries here to ensure reliability.
	convoyClient := convoy.NewWithCredentials(URL, USERNAME, PASSWORD)
	appID := os.Getenv("CONVOY_PAYSTACK_APP_ID")
	_, err = convoyClient.CreateAppEvent(appID, &convoyModels.EventRequest{
		Event: event.Event,
		Data:  []byte(`<insert-real-data>`),
	})

	// TODO(subomi): This is critical. Add logs here.
	if err != nil {
		w.Write([]byte("Server Error: Failed to send event to Convoy"))
		return
	}

}

// MonoHandler ack webhooks from https://mono.com
// TODO(subomi): Define handler
func MonoHandler(w http.ResponseWriter, r *http.Request) {
}

// STRUCTURES

// PaystackEvent structure of paystack webhook events.
type PaystackEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}
