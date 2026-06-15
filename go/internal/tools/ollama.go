//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

// getOllamaHost returns the Ollama host, defaulting to http://localhost:11434.
func getOllamaHost() string {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	return host
}

// HandleListLocalModels lists all locally installed Ollama models.
func HandleListLocalModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/api/tags", getOllamaHost())

	req, errNew := http.NewRequestWithContext(ctx, "GET", url, nil)
	if errNew != nil {
		return errResponseOllama(errNew)
	}

	resp, errDo := client.Do(req)
	if errDo != nil {
		return ok("[ERROR] Ollama Server non accessibile\n\nVerifica installazione e porta 11434.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ok(fmt.Sprintf("[ERROR] Ollama returned status %d", resp.StatusCode))
	}

	bodyBytes, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return errResponseOllama(errRead)
	}

	var data struct {
		Models []struct {
			Name       string    `json:"name"`
			Size       int64     `json:"size"`
			ModifiedAt time.Time `json:"modified_at"`
		} `json:"models"`
	}

	if errUnmarshal := json.Unmarshal(bodyBytes, &data); errUnmarshal != nil {
		return errResponseOllama(errUnmarshal)
	}

	if len(data.Models) == 0 {
		return ok("[ERROR] Nessun modello Ollama trovato.\n\nScarica un modello con: ollama pull llama3.2")
	}

	var sb bytes.Buffer
	sb.WriteString("[MODELLI] LLM Locali Disponibili\n\n")
	for _, m := range data.Models {
		sizeGB := float64(m.Size) / (1024 * 1024 * 1024)
		sb.WriteString(fmt.Sprintf("- %s\n", m.Name))
		sb.WriteString(fmt.Sprintf("  Dimensione: %.1f GB\n", sizeGB))
		sb.WriteString(fmt.Sprintf("  Aggiornato: %s\n\n", m.ModifiedAt.Format(time.RFC3339)))
	}
	sb.WriteString(fmt.Sprintf("[TOTALE] Modelli disponibili: %d\n", len(data.Models)))
	sb.WriteString("[PRIVACY] Tutti i modelli vengono eseguiti localmente, nessun dato inviato al cloud")

	return ok(sb.String())
}

// HandleLocalLLMChat sends a chat message to a local Ollama model.
func HandleLocalLLMChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ := getString(args, "message")
	if message == "" {
		return err("message parameter is required")
	}

	model, _ := getString(args, "model")
	temperature := 0.7
	if tempVal, exists := args["temperature"]; exists {
		if f, okF := tempVal.(float64); okF {
			temperature = f
		}
	}

	// Auto-select model if none specified
	if model == "" {
		client := &http.Client{Timeout: 5 * time.Second}
		url := fmt.Sprintf("%s/api/tags", getOllamaHost())
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			var listData struct {
				Models []struct {
					Name string `json:"name"`
				} `json:"models"`
			}
			if b, errRead := io.ReadAll(resp.Body); errRead == nil {
				_ = json.Unmarshal(b, &listData)
				if len(listData.Models) > 0 {
					model = listData.Models[0].Name
				}
			}
		}
	}

	if model == "" {
		return ok(`{"success": false, "error": "No models available. Download one via 'ollama pull llama3.2'"}`)
	}

	type ChatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type ChatRequest struct {
		Model    string        `json:"model"`
		Messages []ChatMessage `json:"messages"`
		Options  struct {
			Temperature float64 `json:"temperature"`
		} `json:"options"`
		Stream bool `json:"stream"`
	}

	reqBody := ChatRequest{
		Model:    model,
		Messages: []ChatMessage{{Role: "user", Content: message}},
		Stream:   false,
	}
	reqBody.Options.Temperature = temperature

	bBytes, errMarshal := json.Marshal(reqBody)
	if errMarshal != nil {
		return errResponseOllama(errMarshal)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	url := fmt.Sprintf("%s/api/chat", getOllamaHost())
	req, errNew := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bBytes))
	if errNew != nil {
		return errResponseOllama(errNew)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, errDo := client.Do(req)
	if errDo != nil {
		return ok(fmt.Sprintf(`{"success": false, "error": "Ollama server not accessible: %v"}`, errDo))
	}
	defer resp.Body.Close()

	bodyBytes, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return errResponseOllama(errRead)
	}

	var chatResp struct {
		Model   string `json:"model"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if errUnmarshal := json.Unmarshal(bodyBytes, &chatResp); errUnmarshal != nil {
		return ok(fmt.Sprintf(`{"success": false, "error": "Failed to parse Ollama response: %s"}`, string(bodyBytes)))
	}

	resultMap := map[string]interface{}{
		"success":      true,
		"response":     chatResp.Message.Content,
		"model_used":   chatResp.Model,
		"user_message": message,
		"privacy_note": "All processing done locally - no data sent to cloud",
	}

	finalBytes, _ := json.MarshalIndent(resultMap, "", "  ")
	return ok(string(finalBytes))
}

// HandleOllamaHealthCheck checks the Ollama server health.
func HandleOllamaHealthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := getOllamaHost()

	req, errNew := http.NewRequestWithContext(ctx, "GET", url, nil)
	if errNew != nil {
		return errResponseOllama(errNew)
	}

	resp, errDo := client.Do(req)
	healthy := errDo == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	var status, message string
	if healthy {
		status = "HEALTHY"
		message = "Ollama server is running and responsive"
	} else {
		status = "UNHEALTHY"
		message = "Ollama server is down or unreachable"
	}

	resultMap := map[string]interface{}{
		"status":     status,
		"server_url": url,
		"message":    message,
	}

	finalBytes, _ := json.MarshalIndent(resultMap, "", "  ")
	return ok(string(finalBytes))
}

// HandleSystemResourceCheck checks system specs.
func HandleSystemResourceCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sysInfo := map[string]interface{}{
		"os":          runtime.GOOS,
		"architecture": runtime.GOARCH,
		"cpus":        runtime.NumCPU(),
		"compiler":    runtime.Compiler,
	}

	finalBytes, _ := json.MarshalIndent(sysInfo, "", "  ")
	return ok(string(finalBytes))
}

func errResponseOllama(err error) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: fmt.Sprintf(`{"success": false, "error": %q}`, err.Error())}},
		IsError: true,
	}, nil
}
