package replayer

import (
	"bytes"
	"encoding/json"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v7"
)

type operation struct {
	Op     string          `json:"op"` // TODO: make enum
	Target string          `json:"target"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type searchOperation operation
type indexOperation operation

func (s searchOperation) Do(esClient es.Client) error {
	var body bytes.Buffer
	if len(s.Body) > 0 {
		body.Write(s.Body)
	}

	_, err := esClient.Search(
		esClient.Search.WithIndex(s.Target),
		esClient.Search.WithBody(&body),
	)

	if err != nil {
		return fmt.Errorf("search operation failed: %w", err)
	}

	return nil
}

func (i indexOperation) Do(esClient es.Client) error {
	var body bytes.Buffer
	if len(i.Body) > 0 {
		body.Write(i.Body)
	}

	_, err := esClient.Index(
		i.Target,
		&body,
	)

	if err != nil {
		return fmt.Errorf("index operation failed: %w", err)
	}

	return nil
}
