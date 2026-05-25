package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHypercodeProtocol(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name           string
		uri            string
		expectedStatus int
		expectedAction string
	}{
		{
			name:           "Valid attach URI (no supervisor)",
			uri:            "hypercode://attach?session=xyz123",
			expectedStatus: http.StatusServiceUnavailable,
			expectedAction: "",
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
			req := httptest.NewRequest(http.MethodGet, "/api/protocol/hypercode?uri="+tt.uri, nil)
			w := httptest.NewRecorder()

			s.handleHypercodeProtocol(w, req)

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
					t.Fatalf("missing data object in response")
				}

				action, ok := data["action"].(string)
				if !ok || action != tt.expectedAction {
					t.Errorf("expected action %s, got %v", tt.expectedAction, action)
				}
			}
		})
	}
}
