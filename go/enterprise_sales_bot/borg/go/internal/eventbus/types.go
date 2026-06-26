package eventbus

import (
	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

type A2ASignalPayload struct {
	Message orchestration.A2AMessage `json:"message"`
}

type UserActivityPayload struct {
	LastActivityTime int64 `json:"lastActivityTime"`
	ActiveEditor     *struct {
		Uri string `json:"uri"`
	} `json:"activeEditor,omitempty"`
}
