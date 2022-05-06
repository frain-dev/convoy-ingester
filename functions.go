package ingester

import (
	"context"
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
	// GOOGLE_CLOUD_PROJECT is a user-set environment variable.
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

	// Function topic
	topic = os.Getenv("WEBHOOK_TOPIC")

	// client is a global Pub/Sub client, initialized once per instance.
	client *pubsub.Client

	// Configuration Storage
	configStore *Configuration

	// Configuration Environment Variable
	CONFIG_ENV = "CONVOY_INGESTER_CONFIG"

	// Providers Store
	providerStore = make(ProviderStore)
)

func init() {
	// err is pre-declared to avoid shadowing client.
	var err error

	// Set environment to prevent the init function from running in our tests.
	env := os.Getenv("ENV")

	// client is initialized with context.Background() because it should
	// persist between function invocations.
	if env == "prod" {
		client, err = pubsub.NewClient(context.Background(), projectID)
		if err != nil {
			log.Fatalf("pubsub.NewClient: %v", err)
		}

		// Setup configStore
		if err = LoadConfig(CONFIG_ENV); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// TODO(subomi): Initialize providers registry once per instance.
		if err = LoadProviderStore(); err != nil {
			log.Fatalf("Failed to setup provider store: %v", err)
		}
	}
}

// WebhookEndpoint is a HTTP Function to receive events from the world.
func WebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	// Build Router.
	router := chi.NewRouter()

	router.Route("/v1", func(v1Router chi.Router) {

		// TODO(subomi): Use middleware to set provider in the request context.
		v1Router.Post("/webhooks/{provider}", WebhooksHandler)
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

	return nil
}

// HTTP Handlers
func WebhooksHandler(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider := providerStore[providerName]

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Bad Request: Could not read payload"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = provider.VerifyRequest(r, payload)
	if err != nil {
		errMsg := fmt.Sprintf("Bad Request: Could not verify request -  %s", err)
		w.Write([]byte(errMsg))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Push to Convoy.
	event := fmt.Sprintf("%s.event", providerName)
	req := &convoyRequest{
		Data: convoyModels.EventRequest{
			AppID: provider.AppID,
			Event: event,
			Data:  payload,
		},
	}

	data, err := req.ToBytes()
	if err != nil {
		w.Write([]byte("Bad Request: Failed to transform bytes"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := &pubsub.Message{
		Data: data,
	}

	id, err := client.Topic(topic).Publish(r.Context(), m).Get(r.Context())
	if err != nil {
		w.Write([]byte("Bad Request: Error publishing event"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Message published: %v\n", id)
	w.Write([]byte("Event sent"))
	w.WriteHeader(http.StatusOK)
}
