package ingester

import (
	"bytes"
	"encoding/json"

	convoyModels "github.com/frain-dev/convoy-go/models"
)

type publishRequest struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type pubSubMessage struct {
	Data []byte `json:"data"`
}

type convoyRequest struct {
	Data convoyModels.EventRequest `json:"data"`
}

func (c *convoyRequest) ToBytes() ([]byte, error) {
	var buf *bytes.Buffer

	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(c.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *convoyRequest) FromBytes(b []byte) error {
	source := bytes.NewReader(b)
	decoder := json.NewDecoder(source)

	if err := decoder.Decode(c.Data); err != nil {
		return err
	}

	return nil
}
