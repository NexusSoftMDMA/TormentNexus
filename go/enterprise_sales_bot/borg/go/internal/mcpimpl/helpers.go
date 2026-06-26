package mcpimpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ToolResponse struct {
	Content string `json:"content"`
	IsError bool   `json:"isError,omitempty"`
}

func getString(args map[string]interface{}, key string) (string, bool) {
	if args == nil {
		return "", false
	}
	v, ok := args[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getInt(args map[string]interface{}, key string, defaultVal ...int) (int, bool) {
	if args == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0], false
		}
		return 0, false
	}
	v, ok := args[key]
	if !ok {
		if len(defaultVal) > 0 {
			return defaultVal[0], false
		}
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	default:
		if len(defaultVal) > 0 {
			return defaultVal[0], false
		}
		return 0, false
	}
}

func getFloat(args map[string]interface{}, key string) float64 {
	if args == nil {
		return 0
	}
	v, _ := args[key]
	f, _ := v.(float64)
	return f
}

func getBool(args map[string]interface{}, key string) bool {
	if args == nil {
		return false
	}
	v, _ := args[key]
	b, _ := v.(bool)
	return b
}

func err(msg string) (ToolResponse, error) {
	return ToolResponse{Content: msg, IsError: true}, fmt.Errorf("%s", msg)
}

func ok(msg interface{}) (ToolResponse, error) {
	switch v := msg.(type) {
	case string:
		return ToolResponse{Content: v}, nil
	default:
		b, _ := json.MarshalIndent(v, "", "  ")
		return ToolResponse{Content: string(b)}, nil
	}
}

func success(msg interface{}) (ToolResponse, error) {
	switch v := msg.(type) {
	case string:
		return ToolResponse{Content: v}, nil
	default:
		b, _ := json.MarshalIndent(v, "", "  ")
		return ToolResponse{Content: string(b)}, nil
	}
}

func okJSON(data interface{}) (ToolResponse, error) {
	b, _ := json.MarshalIndent(data, "", "  ")
	return ToolResponse{Content: string(b)}, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url, contentType string, body []byte) ([]byte, error) {
	resp, err := http.Post(url, contentType, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func jsonGet(url string, target interface{}) error {
	data, err := httpGet(url)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func jsonPost(url string, body, target interface{}) error {
	b, _ := json.Marshal(body)
	data, err := httpPost(url, "application/json", b)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func contains(s []string, item string) bool {
	for _, x := range s {
		if strings.EqualFold(x, item) {
			return true
		}
	}
	return false
}

func joinStrings(elems []string, sep string) string {
	return strings.Join(elems, sep)
}
