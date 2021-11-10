package models

import "encoding/json"

type Operation struct {
	Op     string          `json:"op"` // TODO: make enum
	Target string          `json:"target"`
	Body   json.RawMessage `json:"body,omitempty"`
}
