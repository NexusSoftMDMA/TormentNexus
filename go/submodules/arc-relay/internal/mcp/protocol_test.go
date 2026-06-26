package mcp

import (
	"encoding/json"
	"testing"
)

func TestNewRequest(t *testing.T) {
	t.Run("with params", func(t *testing.T) {
		id := json.RawMessage(`1`)
		params := map[string]string{"name": "test_tool"}
		req, err := NewRequest(id, "tools/call", params)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}
		if req.JSONRPC != "2.0" {
			t.Errorf("JSONRPC = %q, want %q", req.JSONRPC, "2.0")
		}
		if req.Method != "tools/call" {
			t.Errorf("Method = %q, want %q", req.Method, "tools/call")
		}
		if string(req.ID) != "1" {
			t.Errorf("ID = %s, want 1", string(req.ID))
		}
		if req.Params == nil {
			t.Fatal("Params should not be nil")
		}

		var p map[string]string
		if err := json.Unmarshal(req.Params, &p); err != nil {
			t.Fatalf("unmarshal params: %v", err)
		}
		if p["name"] != "test_tool" {
			t.Errorf("params.name = %q, want %q", p["name"], "test_tool")
		}
	})

	t.Run("nil params", func(t *testing.T) {
		id := json.RawMessage(`"abc"`)
		req, err := NewRequest(id, "initialize", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}
		if req.Params != nil {
			t.Errorf("Params = %s, want nil", string(req.Params))
		}
	})

	t.Run("JSON round-trip", func(t *testing.T) {
		id := json.RawMessage(`42`)
		req, err := NewRequest(id, "tools/list", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Marshal error = %v", err)
		}

		var decoded Request
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}
		if decoded.JSONRPC != "2.0" {
			t.Errorf("decoded JSONRPC = %q, want %q", decoded.JSONRPC, "2.0")
		}
		if decoded.Method != "tools/list" {
			t.Errorf("decoded Method = %q, want %q", decoded.Method, "tools/list")
		}
	})
}

func TestNewResponse(t *testing.T) {
	t.Run("with result", func(t *testing.T) {
		id := json.RawMessage(`1`)
		result := ToolsListResult{Tools: []Tool{{Name: "test", Description: "a tool"}}}
		resp, err := NewResponse(id, result)
		if err != nil {
			t.Fatalf("NewResponse() error = %v", err)
		}
		if resp.JSONRPC != "2.0" {
			t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, "2.0")
		}
		if resp.Error != nil {
			t.Errorf("Error should be nil, got %+v", resp.Error)
		}
		if resp.Result == nil {
			t.Fatal("Result should not be nil")
		}

		var tools ToolsListResult
		if err := json.Unmarshal(resp.Result, &tools); err != nil {
			t.Fatalf("unmarshal result: %v", err)
		}
		if len(tools.Tools) != 1 || tools.Tools[0].Name != "test" {
			t.Errorf("unexpected result: %+v", tools)
		}
	})

	t.Run("nil result marshals to null", func(t *testing.T) {
		id := json.RawMessage(`1`)
		resp, err := NewResponse(id, nil)
		if err != nil {
			t.Fatalf("NewResponse() error = %v", err)
		}
		// nil marshals to JSON "null"
		if string(resp.Result) != "null" {
			t.Errorf("Result = %s, want null", string(resp.Result))
		}
	})
}

func TestNewErrorResponse(t *testing.T) {
	id := json.RawMessage(`99`)
	resp := NewErrorResponse(id, ErrCodeMethodNotFound, "method not found")

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, "2.0")
	}
	if string(resp.ID) != "99" {
		t.Errorf("ID = %s, want 99", string(resp.ID))
	}
	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, ErrCodeMethodNotFound)
	}
	if resp.Error.Message != "method not found" {
		t.Errorf("Error.Message = %q, want %q", resp.Error.Message, "method not found")
	}
	if resp.Result != nil {
		t.Errorf("Result should be nil for error response")
	}

	// Verify error codes are standard JSON-RPC values
	if ErrCodeParse != -32700 {
		t.Errorf("ErrCodeParse = %d, want -32700", ErrCodeParse)
	}
	if ErrCodeInvalidRequest != -32600 {
		t.Errorf("ErrCodeInvalidRequest = %d, want -32600", ErrCodeInvalidRequest)
	}
	if ErrCodeInvalidParams != -32602 {
		t.Errorf("ErrCodeInvalidParams = %d, want -32602", ErrCodeInvalidParams)
	}
	if ErrCodeInternal != -32603 {
		t.Errorf("ErrCodeInternal = %d, want -32603", ErrCodeInternal)
	}
}

func TestNotificationJSON(t *testing.T) {
	n := Notification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}

	// Notification should not have an "id" field
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if _, hasID := raw["id"]; hasID {
		t.Error("Notification JSON should not contain 'id' field")
	}
	if string(raw["jsonrpc"]) != `"2.0"` {
		t.Errorf("jsonrpc = %s, want %q", string(raw["jsonrpc"]), "2.0")
	}
	if string(raw["method"]) != `"notifications/initialized"` {
		t.Errorf("method = %s, want %q", string(raw["method"]), "notifications/initialized")
	}
}
