package ingester

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	convoy "github.com/frain-dev/convoy-go"
	convoyModels "github.com/frain-dev/convoy-go/models"
	"github.com/go-chi/chi/v5"
)

var (
	URL      = os.Getenv("CONVOY_URL")
	USERNAME = os.Getenv("CONVOY_USERNAME")
	PASSWORD = os.Getenv("CONVOY_PASSWORD")

	// GOOGLE_CLOUD_PROJECT is a user-set environment variable.
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

	// Function topic
	topic = os.Getenv("WEBHOOK_TOPIC")

	// client is a global Pub/Sub client, initialized once per instance.
	client *pubsub.Client
)

func init() {
	// err is pre-declared to avoid shadowing client.
	var err error

	// client is initialized with context.Background() because it should
	// persist between function invocations.
	client, err = pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
}

// WebhookEndpoint is a HTTP Function to receive events from the world.
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

// PushToConvoy is a Pub/Sub Triggered Function to push events to Convoy.
func PushToConvoy(ctx context.Context, m pubSubMessage) error {
	payload := []byte(string(m.Data))
	req := &convoyRequest{}
	if err := req.FromBytes(payload); err != nil {
		log.Printf("Failed to parse payload - %v", err)
		return err
	}

	// Actual push to Convoy.
	convoyClient := convoy.New()
	_, err := convoyClient.CreateAppEvent(&req.Data)

	if err != nil {
		return errors.New(fmt.Sprintf("Server Error: Failed to send event to Convoy - %+v", err))
	}

}

// HTTP Handlers

// PaystackHandler ack webhooks from https://paystack.com
// NEEDS:
// CONVOY_PAYSTACK_APP_ID
// PAYSTACK_SECRET
// TODO(subomi): Handle response and response status code properly.
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
	appID := os.Getenv("CONVOY_PAYSTACK_APP_ID")
	req := &convoyRequest{
		AppID: appID,
		Data: convoyModels.EventRequest{
			Event: event.Event,
			Data:  payload,
		},
	}

	data, err := req.ToBytes()
	if err != nil {
		log.Printf("Failed to transform to bytes")
		w.Write([]byte(`Failed to transform bytes`))
		return
	}

	m := &pubsub.Message{
		Data: data,
	}
	id, err := client.Topic(topic).Publish(r.Context(), m).Get(r.Context())
	if err != nil {
		log.Printf("topic(%s).Publish.Get: %v", topic, err)
		w.Write([]byte("Error publishing event"))
		return
	}
	fmt.Fprintf(w, "Message published: %v", id)

	w.Write([]byte("Event sent"))
}

// MonoHandler ack webhooks from https://mono.com
// TODO(subomi): Define handler
func MonoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Mono Webhooks Handler")
}

// STRUCTURES

// PaystackEvent structure of paystack webhook events.
type PaystackEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}
