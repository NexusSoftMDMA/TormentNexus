package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tormentnexushq/tormentnexus-go/internal/eventbus"
	"github.com/tormentnexushq/tormentnexus-go/internal/supervisor"
)

func TestHandleTormentNexusProtocol(t *testing.T) {
	s := &Server{
		supervisorManager: supervisor.NewManager(supervisor.ManagerOptions{}),
		eventBus:          eventbus.New(10),
	}

	tests := []struct {
		name           string
		uri            string
		expectedStatus int
		expectedAction string
	}{
		{
			name:           "Valid attach URI",
			uri:            "tormentnexus://attach?session=xyz123",
			expectedStatus: http.StatusOK,
			expectedAction: "attach",
		},
		{
			name:           "Missing URI",
			uri:            "",
			expectedStatus: http.StatusBadRequest,
			expectedAction: "",
		},
		{
			name:           "Invalid Scheme",
			uri:            "http://attach?session=xyz",
			expectedStatus: http.StatusBadRequest,
			expectedAction: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedAction == "attach" {
				s.supervisorManager.CreateSession("xyz123", "echo", []string{"hello"}, nil, ".", 0)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/protocol/tormentnexus?uri="+tt.uri, nil)
			w := httptest.NewRecorder()

			s.handleTormentNexusProtocol(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusOK {
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("missing data object in response: %v", resp)
				}

				action, ok := data["action"].(string)
				if !ok || action != tt.expectedAction {
					t.Errorf("expected action %s, got %v", tt.expectedAction, action)
				}
			}
		})
	}
}
