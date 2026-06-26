package ai

import (
	"context"
	"fmt"
	"testing"
)

type MockClient struct {
	Response *LLMResponse
	Error    error
}

func (m *MockClient) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	return m.Response, m.Error
}

func TestWaterfallClient_SuccessOnFirst(t *testing.T) {
	c1 := &MockClient{Response: &LLMResponse{Content: "success 1", Provider: "mock1"}}
	c2 := &MockClient{Response: &LLMResponse{Content: "success 2", Provider: "mock2"}}
	wc := NewWaterfallClient(nil, c1, c2)
	resp, err := wc.GenerateText(context.Background(), "test-model", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "success 1" {
		t.Fatalf("expected 'success 1', got '%s'", resp.Content)
	}
}

func TestWaterfallClient_CascadeOn429(t *testing.T) {
	c1 := &MockClient{Error: fmt.Errorf("HTTP 429 Too Many Requests")}
	c2 := &MockClient{Response: &LLMResponse{Content: "success 2", Provider: "mock2"}}
	wc := NewWaterfallClient(nil, c1, c2)
	resp, err := wc.GenerateText(context.Background(), "test-model", nil)
	if err != nil {
		t.Fatalf("expected no error after cascade, got %v", err)
	}
	if resp.Content != "success 2" {
		t.Fatalf("expected 'success 2', got '%s'", resp.Content)
	}
}

func TestWaterfallClient_CascadeOn500(t *testing.T) {
	c1 := &MockClient{Error: fmt.Errorf("HTTP 500 Internal Server Error")}
	c2 := &MockClient{Response: &LLMResponse{Content: "success 2", Provider: "mock2"}}
	wc := NewWaterfallClient(nil, c1, c2)
	resp, err := wc.GenerateText(context.Background(), "test-model", nil)
	if err != nil {
		t.Fatalf("expected no error after cascade, got %v", err)
	}
	if resp.Content != "success 2" {
		t.Fatalf("expected 'success 2', got '%s'", resp.Content)
	}
}

func TestWaterfallClient_FailsOn400(t *testing.T) {
	c1 := &MockClient{Error: fmt.Errorf("HTTP 400 Bad Request")}
	c2 := &MockClient{Response: &LLMResponse{Content: "success 2", Provider: "mock2"}}
	wc := NewWaterfallClient(nil, c1, c2)
	_, err := wc.GenerateText(context.Background(), "test-model", nil)
	if err == nil {
		t.Fatalf("expected error on 400, got nil")
	}
}

func TestWaterfallClient_FailsIfAllFail(t *testing.T) {
	c1 := &MockClient{Error: fmt.Errorf("HTTP 500")}
	c2 := &MockClient{Error: fmt.Errorf("HTTP 503")}
	wc := NewWaterfallClient(nil, c1, c2)
	_, err := wc.GenerateText(context.Background(), "test-model", nil)
	if err == nil {
		t.Fatalf("expected error when all tiers fail, got nil")
	}
}
