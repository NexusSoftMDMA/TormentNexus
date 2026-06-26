package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	s := &Server{}
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.handleHealth)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != "OK\n" {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "OK\n")
	}
}

func TestHandleGenerateQuote(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil) // Dependencies aren't strictly needed for this endpoint since it just calculates a quote based on URL parameters

	req, err := http.NewRequest("GET", "/api/v1/quote?company_size=enterprise", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"quote":50000,"tier":"enterprise"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
