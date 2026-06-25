package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HandleInspectServer implements the core inspection logic.
// It attempts to connect to a server (HTTP or STDIO) and list its capabilities.
func HandleInspectServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "url")
	command, _ :=getString(args, "command")
	argsList, _ :=getString(args, "args")

	if serverURL == "" && command == "" {
		return err("Either 'url' or 'command' must be provided")
}

	var result map[string]interface{}
	var e error

	if serverURL != "" {
		result, e = inspectHTTPServer(ctx, serverURL)
	} else {
		result, e = inspectStdioServer(ctx, command, argsList)

	if e != nil {
		return err(e.Error())
}

	jsonResult, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(fmt.Sprintf("Failed to marshal result: %v", marshalErr))
}

	return ok(string(jsonResult))
}

}

// inspectHTTPServer performs a standard MCP handshake via HTTP.
func inspectHTTPServer(ctx context.Context, rawURL string) (map[string]interface{}, error) {
	// Validate URL
	parsedURL, urlErr := url.Parse(rawURL)
	if urlErr != nil {
		return nil, fmt.Errorf("invalid URL: %w", urlErr)
}

	// Construct the initialization request
	// Standard MCP JSON-RPC 2.0 initialization
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "mcpjam-go-inspector",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		},
	}

	jsonBody, jsonErr := json.Marshal(initReq)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", jsonErr)
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", parsedURL.String(), strings.NewReader(string(jsonBody)))
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return nil, fmt.Errorf("failed to fetch server: %w", fetchErr)
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
}

	// Parse the JSON-RPC response
	var response map[string]interface{}
	parseErr := json.Unmarshal(body, &response)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC response: %w", parseErr)
}

	// Check for error in response
	if respVal, found := response["error"]; found {
		return nil, fmt.Errorf("server returned JSON-RPC error: %v", respVal)
}

	// Extract server info
	result := map[string]interface{}{
		"status": "connected",
		"server": response["result"],
	}

	// If capabilities are present, try to list tools/resources/prompts
	if serverResult, found := response["result"].(map[string]interface{}); found {
		if capabilities, found := serverResult["capabilities"].(map[string]interface{}); found {
			result["capabilities"] = capabilities
		}
	}

	return result, nil
}

// inspectStdioServer attempts to start a local process and communicate via STDIO.
func inspectStdioServer(ctx context.Context, command, argsStr string) (map[string]interface{}, error) {
	if command == "" {
		return nil, fmt.Errorf("command is required for STDIO inspection")
}

	// Parse arguments
	var cmdArgs []string
	if argsStr != "" {
		// Simple space splitting for demo purposes; real implementation might use shell parsing
		cmdArgs = strings.Fields(argsStr)

	cmd := exec.CommandContext(ctx, command, cmdArgs...)
	
	// Setup pipes
	stdin, stdinErr := cmd.StdinPipe()
	if stdinErr != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", stdinErr)
}

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", stdoutErr)
}

	stderr, stderrErr := cmd.StderrPipe()
	if stderrErr != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", stderrErr)
}

	// Start process
	startErr := cmd.Start()
	if startErr != nil {
		return nil, fmt.Errorf("failed to start command: %w", startErr)
}

	// Read stderr in a goroutine to prevent blocking
	go func() {
		buf := make([]byte, 1024)
		for {
			n, readErr := stderr.Read(buf)
			if readErr != nil {
				return
			}
			if n > 0 {
				// Log to console or ignore for this tool
				// fmt.Fprintf(os.Stderr, "STDERR: %s", buf[:n])

		}
	}()

	// Send initialization request
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "mcpjam-go-inspector",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{},
		},
	}

	jsonBody, jsonErr := json.Marshal(initReq)
	if jsonErr != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to marshal request: %w", jsonErr)
}

	// Write to stdin
	_, writeErr := stdin.Write(append(jsonBody, '\n'))
	if writeErr != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to write to stdin: %w", writeErr)
}

	stdin.Close()

	// Read response from stdout
	// Note: In a real implementation, we'd handle line-delimited JSON or specific framing
	// Here we assume a single response for simplicity or read until EOF
	respBytes := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, readErr := stdout.Read(buf)
		if n > 0 {
			respBytes = append(respBytes, buf[:n]...)

		if readErr != nil {
			break
		}
	}

	// Wait for process to exit
	waitErr := cmd.Wait()
	if waitErr != nil && waitErr.Error() != "signal: killed" {
		// Non-zero exit might be expected if the server shuts down after init
		// We proceed with the data we have
	}

	if len(respBytes) == 0 {
		return nil, fmt.Errorf("no response received from STDIO server")
}

	var response map[string]interface{}
	parseErr := json.Unmarshal(respBytes, &response)
	if parseErr != nil {
		// Try to find JSON in the output if it's mixed with logs
		jsonRegex := regexp.MustCompile(`\{.*"jsonrpc".*\}`)
		match := jsonRegex.Find(respBytes)
		if match == nil {
			return nil, fmt.Errorf("failed to parse JSON response: %w, raw: %s", parseErr, string(respBytes))
}

		parseErr = json.Unmarshal(match, &response)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse extracted JSON: %w", parseErr)

	}

	if respVal, found := response["error"]; found {
		return nil, fmt.Errorf("server returned JSON-RPC error: %v", respVal)
}

	return map[string]interface{}{
}
		"status": "connected",
		"server": response["result"],
	}, nil
}

}
}
}

// HandleListTools lists available tools from a connected server.
// This is a simplified version that assumes the server is already known or passed via URL.
func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "url")
	if serverURL == "" {
		return err("URL is required to list tools")
}

	// Reuse the HTTP inspection logic but target the tools endpoint or list method
	// Since MCP doesn't have a direct "list tools" HTTP endpoint without the protocol handshake,
	// we will perform the handshake and then call "tools/list" if supported.
	
	// 1. Initialize
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{"name": "mcpjam-go-inspector", "version": "1.0.0"},
			"capabilities":  map[string]interface{}{},
		},
	}
	
	jsonBody, jsonErr := json.Marshal(initReq)
	if jsonErr != nil {
		return err(fmt.Sprintf("Marshal error: %v", jsonErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", serverURL, strings.NewReader(string(jsonBody)))
	if reqErr != nil {
		return err(fmt.Sprintf("Request creation error: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Fetch error: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Read error: %v", readErr))
}

	var initResp map[string]interface{}
	if parseErr := json.Unmarshal(body, &initResp); parseErr != nil {
		return err(fmt.Sprintf("Parse error: %v", parseErr))
}

	if _, found := initResp["error"]; found {
		return err("Initialization failed")
}

	// 2. Call tools/list
	listReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}
	
	listBody, listJsonErr := json.Marshal(listReq)
	if listJsonErr != nil {
		return err(fmt.Sprintf("List marshal error: %v", listJsonErr))
}

	req2, req2Err := http.NewRequestWithContext(ctx, "POST", serverURL, strings.NewReader(string(listBody)))
	if req2Err != nil {
		return err(fmt.Sprintf("List request creation error: %v", req2Err))
}

	req2.Header.Set("Content-Type", "application/json")

	resp2, fetch2Err := client.Do(req2)
	if fetch2Err != nil {
		return err(fmt.Sprintf("List fetch error: %v", fetch2Err))
}

	defer resp2.Body.Close()

	body2, read2Err := io.ReadAll(resp2.Body)
	if read2Err != nil {
		return err(fmt.Sprintf("List read error: %v", read2Err))
}

	var listResp map[string]interface{}
	if parse2Err := json.Unmarshal(body2, &listResp); parse2Err != nil {
		return err(fmt.Sprintf("List parse error: %v", parse2Err))
}

	if _, found := listResp["error"]; found {
		return err("Server does not support tools/list or returned an error")
}

	result, found := listResp["result"]
	if !found {
		return err("No result in tools/list response")
}

	jsonOut, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(fmt.Sprintf("Final marshal error: %v", marshalErr))
}

	return ok(string(jsonOut))
}

// HandleCheckHealth performs a simple connectivity check.
func HandleCheckHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "url")
	if serverURL == "" {
		return err("URL is required")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("Request creation error: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Connection failed: %v", fetchErr))
}

	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode != http.StatusOK {
		status = fmt.Sprintf("unhealthy (status: %d)", resp.StatusCode)

	return ok(fmt.Sprintf("Server at %s is %s", serverURL, status))
}

}

// HandleParseLog parses a log file or string to extract JSON-RPC messages.
func HandleParseLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	logContent, _ :=getString(args, "content")
	logPath, _ :=getString(args, "path")

	var content string
	if logPath != "" {
		fileContent, readErr := os.ReadFile(logPath)
		if readErr != nil {
			return err(fmt.Sprintf("Failed to read file: %v", readErr))
}

		content = string(fileContent)
	} else {
		content = logContent
	}

	if content == "" {
		return err("No content or path provided")
}

	// Split by newlines and try to parse each line as JSON
	lines := strings.Split(content, "\n")
	var messages []map[string]interface{}
	jsonRegex := regexp.MustCompile(`\{.*"jsonrpc".*\}`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try direct parse first
		var msg map[string]interface{}
		if parseErr := json.Unmarshal([]byte(line), &msg); parseErr == nil {
			if _, found := msg["jsonrpc"]; found {
				messages = append(messages, msg)
				continue
			}
		}

		// Try regex extraction for multi-line or noisy logs
		match := jsonRegex.FindString(line)
		if match != "" {
			var msg map[string]interface{}
			if parseErr := json.Unmarshal([]byte(match), &msg); parseErr == nil {
				if _, found := msg["jsonrpc"]; found {
					messages = append(messages, msg)

			}
		}
	}

	// Sort by ID if present
	sort.Slice(messages, func(i, j int) bool {
		idI := 0
		idJ := 0
		if val, found := messages[i]["id"]; found {
			if intVal, found := val.(float64); found {
				idI = int(intVal)

		}
		if val, found := messages[j]["id"]; found {
			if intVal, found := val.(float64); found {
				idJ = int(intVal)

		}
		return idI < idJ
	})

	jsonOut, marshalErr := json.Marshal(messages)
	if marshalErr != nil {
		return err(fmt.Sprintf("Failed to marshal messages: %v", marshalErr))
}

	return ok(string(jsonOut))
}
}
}
}