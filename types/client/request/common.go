package request

import "encoding/json"

type Common struct {
	RequestID int             `json:"request_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}
