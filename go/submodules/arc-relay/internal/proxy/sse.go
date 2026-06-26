package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/comma-compliance/arc-relay/internal/mcp"
)

// parseHTTPResponse reads an MCP response from an HTTP response body,
// handling both plain JSON and SSE (text/event-stream) formats.
func parseHTTPResponse(httpResp *http.Response) (*mcp.Response, error) {
	contentType := httpResp.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "text/event-stream") {
		return parseSSEResponse(httpResp.Body)
	}

	// Plain JSON
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var resp mcp.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing JSON response: %w (body: %s)", err, truncate(body, 200))
	}
	return &resp, nil
}

// parseSSEResponse reads the first "message" event from an SSE stream
// and parses its data field as an MCP JSON-RPC response.
func parseSSEResponse(r io.Reader) (*mcp.Response, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line length
	var dataLines []string
	var eventType string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			dataLines = append(dataLines, data)
			continue
		}

		// Empty line = end of event
		if line == "" && len(dataLines) > 0 {
			if eventType == "message" || eventType == "" {
				combined := strings.Join(dataLines, "\n")
				var resp mcp.Response
				if err := json.Unmarshal([]byte(combined), &resp); err != nil {
					return nil, fmt.Errorf("parsing SSE data as JSON-RPC: %w (data: %s)", err, truncate([]byte(combined), 200))
				}
				return &resp, nil
			}
			// Reset for next event
			dataLines = nil
			eventType = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading SSE stream: %w", err)
	}

	// If we collected data lines but didn't hit an empty line (end of stream)
	if len(dataLines) > 0 {
		combined := strings.Join(dataLines, "\n")
		var resp mcp.Response
		if err := json.Unmarshal([]byte(combined), &resp); err != nil {
			return nil, fmt.Errorf("parsing final SSE data: %w", err)
		}
		return &resp, nil
	}

	return nil, fmt.Errorf("no message event found in SSE stream")
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
