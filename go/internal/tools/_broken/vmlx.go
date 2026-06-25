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
	"strconv"
	"strings"
	"time"
	"regexp"
	"sort"
)

// HandleHealthCheck returns a simple health status.
func HandleHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("vmlx is healthy")
}

// HandleServeModel starts a vmlx server for the given model.
// Required args:
//   - model: string, the model identifier (e.g., "mlx-community/Qwen3-8B-4bit")
// Optional args:
//   - port: int, port to bind (default 8000)
func HandleServeModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("missing required argument: model")
	}
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 8000
	}

	// Build command: vmlx serve <model> --port <port>
	cmd := exec.CommandContext(ctx, "vmlx", "serve", model, "--port", strconv.Itoa(port))
	// Redirect stdout/stderr to files for debugging (optional)
	stdoutPath := filepath.Join(os.TempDir(), "vmlx_stdout.log")
	stderrPath := filepath.Join(os.TempDir(), "vmlx_stderr.log")
	stdoutFile, _ := os.Create(stdoutPath)
	stderrFile, _ := os.Create(stderrPath)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	if startErr := cmd.Start(); startErr != nil {
		return err(fmt.Sprintf("failed to start vmlx: %v", startErr))
	}
	// Do not wait; assume it starts successfully.
	return ok(fmt.Sprintf("vmlx serving model %s on port %d (pid %d)", model, port, cmd.Process.Pid))
}

// HandleGenerate sends a chat completion request to a running vmlx server.
// Required args:
//   - prompt: string, user prompt.
// Optional args:
//   - max_tokens: int, maximum tokens to generate (default 256)
//   - temperature: float64, sampling temperature (default 0.7)
func HandleGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing required argument: prompt")
	}
	maxTokens, _ :=getInt(args, "max_tokens")
	if maxTokens == 0 {
		maxTokens = 256
	}
	temperature, _ :=getString(args, "temperature") // we accept string to keep getString signature
	// Build request payload
	payload := map[string]interface{}{
		"model": "local",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": maxTokens,
		"temperature": func() interface{} {
			if temperature != "" {
				if f, convErr := strconv.ParseFloat(temperature, 64); convErr == nil {
					return f
				}
			}
			return 0.7
		}(),
		"stream": false,
	}
	// Perform HTTP request
	respBody, apiErr := requestJSON(ctx, "POST", "http://localhost:8000/v1/chat/completions", payload)
	if apiErr != nil {
		return err(fmt.Sprintf("request error: %v", apiErr))
	}
	text, extractErr := extractChatResponse(respBody)
	if extractErr != nil {
		return err(fmt.Sprintf("response parse error: %v", extractErr))
	}
	return ok(text)
}

// HandleListModels runs `vmlx list` and returns the raw output.
func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "vmlx", "list")
	output, listErr := cmd.CombinedOutput()
	if listErr != nil {
		return err(fmt.Sprintf("failed to list models: %v", listErr))
	}
	return ok(string(output))
}

// HandleReadConfig reads a JSON configuration file and returns its pretty‑printed content.
// Required args:
//   - path: string, absolute or relative path to the config file.
func HandleReadConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing required argument: path")
	}
	absPath, absErr := filepath.Abs(path)
	if absErr != nil {
		return err(fmt.Sprintf("invalid path: %v", absErr))
	}
	data, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return err(fmt.Sprintf("cannot read file: %v", readErr))
	}
	// Ensure valid JSON
	var js interface{}
	if jsonErr := json.Unmarshal(data, &js); jsonErr != nil {
		return err(fmt.Sprintf("invalid JSON: %v", jsonErr))
	}
	pretty, _ := json.MarshalIndent(js, "", "  ")
	return ok(string(pretty))
}

// requestJSON performs an HTTP request with a JSON body and returns the decoded response.
func requestJSON(ctx context.Context, method, rawURL string, body interface{}) (interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		b, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, fmt.Errorf("json marshal error: %w", marshalErr)
}

		reqBody = strings.NewReader(string(b))

	req, reqErr := http.NewRequestWithContext(ctx, method, rawURL, reqBody)
	if reqErr != nil {
		return nil, fmt.Errorf("request creation error: %w", reqErr)
}

	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	resp, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("http error: %w", respErr)
}

	defer resp.Body.Close()
	respBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read body error: %w", readErr)
}

	if len(respBytes) == 0 {
		return nil, fmt.Errorf("empty response")
}

	var result interface{}
	if jsonErr := json.Unmarshal(respBytes, &result); jsonErr != nil {
		// Return raw string if not JSON
		return string(respBytes), nil
	}
	return result, nil
}

}

// extractChatResponse pulls the generated text from a vmlx chat completion response.
func extractChatResponse(resp interface{}) (string, error) {
	m, found := resp.(map[string]interface{})
	if !found {
		return "", fmt.Errorf("unexpected response format")
}

	choices, found := m["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in response")
}

	firstChoice, found := choices[0].(map[string]interface{})
	if !found {
		return "", fmt.Errorf("invalid choice structure")
}

	message, found := firstChoice["message"].(map[string]interface{})
	if !found {
		return "", fmt.Errorf("missing message field")
}

	content, found := message["content"].(string)
	if !found {
		return "", fmt.Errorf("content not a string")
}

	return content, nil
}