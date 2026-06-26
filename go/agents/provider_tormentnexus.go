package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TormentNexusControlPlaneProvider struct {
	BaseURL string
}

func NewTormentNexusProvider() *TormentNexusControlPlaneProvider {
	return &TormentNexusControlPlaneProvider{
		BaseURL: "http://127.0.0.1:4000",
	}
}

func (p *TormentNexusControlPlaneProvider) Chat(ctx context.Context, messages []Message, tools []Tool) (Message, error) {
	// Re-map messages to the format expected by the /api/agent/chat endpoint
	type payloadMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	var history []payloadMsg
	for _, msg := range messages {
		history = append(history, payloadMsg{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"message": "", // We send the full history
		"history": history,
	})
	if err != nil {
		return Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/api/agent/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return Message{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to contact TormentNexus Control Plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Message{}, fmt.Errorf("TormentNexus API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Content  string `json:"content"`
			Provider string `json:"provider"`
			Model    string `json:"model"`
		} `json:"data"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Message{}, fmt.Errorf("failed to parse TormentNexus response: %w", err)
	}

	if !result.Success {
		return Message{}, fmt.Errorf("TormentNexus rejected chat: %s", result.Error)
	}

	return Message{
		Role:    RoleAssistant,
		Content: result.Data.Content,
	}, nil
}

func (p *TormentNexusControlPlaneProvider) Stream(ctx context.Context, messages []Message, tools []Tool, chunkChan chan<- string) error {
	// Fallback to synchronous chat if streaming isn't perfectly supported on the sidecar yet
	msg, err := p.Chat(ctx, messages, tools)
	if err != nil {
		return err
	}
	chunkChan <- msg.Content
	close(chunkChan)
	return nil
}

func (p *TormentNexusControlPlaneProvider) GetModelName() string {
	return "tormentnexus-router-active"
}
