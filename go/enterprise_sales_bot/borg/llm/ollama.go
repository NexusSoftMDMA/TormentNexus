package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaClient handles communication with a local Ollama instance
type OllamaClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type OllamaModel struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
}

type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// ListModels retrieves all available local models
func (c *OllamaClient) ListModels() ([]OllamaModel, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/tags", c.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ollama: %w", err)
	}
	defer resp.Body.Close()

	var tagsResp OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	return tagsResp.Models, nil
}
