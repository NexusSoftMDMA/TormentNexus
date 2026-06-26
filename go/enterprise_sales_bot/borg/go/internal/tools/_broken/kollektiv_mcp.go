package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func HandleUploadDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filename, _ :=getString(args, "filename")
	if filename == "" {
		return err("filename is required")
}

	authToken, _ :=getString(args, "auth_token")
	if authToken == "" {
		return err("auth_token is required")
}

	content, _ :=getString(args, "content")
	filePath, _ :=getString(args, "file_path")
	var fileContent []byte
	var readErr error

	if content != "" {
		fileContent = []byte(content)
	} else if filePath != "" {
		absPath, pathErr := filepath.Abs(filePath)
		if pathErr != nil {
			return err(fmt.Sprintf("invalid file path: %v", pathErr))
}

		if _, statErr := os.Stat(absPath); statErr != nil {
			return err(fmt.Sprintf("file not found: %v", statErr))
}

		fileContent, readErr = os.ReadFile(absPath)
		if readErr != nil {
			return err(fmt.Sprintf("failed to read file: %v", readErr))

	} else {
		return err("either content or file_path is required")
}

	apiURL := "https://mcp.thekollektiv.ai/api/upload"
	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(fileContent)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Filename", filename)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to send request: %v", fetchErr))
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("upload failed with status %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(fmt.Sprintf("Successfully uploaded %s to Kollektiv knowledge base", filename))
}

}

func HandleSearchKnowledgeBase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	authToken, _ :=getString(args, "auth_token")
	if authToken == "" {
		return err("auth_token is required")
}

	limit := 5
	if _, found := args["limit"]; found {
		limit = getInt(args, "limit")
		if limit <= 0 {
			limit = 5
		}
	}

	apiURL := "https://mcp.thekollektiv.ai/api/search"
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	q := url.Values{}
	q.Set("q", query)
	q.Set("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to send request: %v", fetchErr))
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("search failed with status %d: %s", resp.StatusCode, string(respBody)))
}

	var searchResults struct {
		Results []struct {
			Document string  `json:"document"`
			Content  string  `json:"content"`
			Score    float64 `json:"score"`
		} `json:"results"`
	}
	parseErr := json.Unmarshal(respBody, &searchResults)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse search results: %v", parseErr))
}

	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("Found %d results for query '%s':\n\n", len(searchResults.Results), query))
	for i, res := range searchResults.Results {
		resultText.WriteString(fmt.Sprintf("%d. Document: %s (Score: %.2f)\n", i+1, res.Document, res.Score))
		resultText.WriteString(fmt.Sprintf(" Content: %s\n\n", res.Content))

	return ok(resultText.String())
}

}

func HandleListDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	authToken, _ :=getString(args, "auth_token")
	if authToken == "" {
		return err("auth_token is required")
}

	apiURL := "https://mcp.thekollektiv.ai/api/documents"
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to send request: %v", fetchErr))
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to list documents with status %d: %s", resp.StatusCode, string(respBody)))
}

	var documents struct {
		Documents []struct {
			ID         string `json:"id"`
			Filename   string `json:"filename"`
			Size       int64  `json:"size"`
			UploadedAt string `json:"uploaded_at"`
		} `json:"documents"`
	}
	parseErr := json.Unmarshal(respBody, &documents)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse documents list: %v", parseErr))
}

	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("Found %d documents in your Kollektiv knowledge base:\n\n", len(documents.Documents)))
	for i, doc := range documents.Documents {
		sizeKB := float64(doc.Size) / 1024
		resultText.WriteString(fmt.Sprintf("%d. %s (ID: %s)\n", i+1, doc.Filename, doc.ID))
		resultText.WriteString(fmt.Sprintf(" Size: %.2f KB | Uploaded: %s\n\n", sizeKB, doc.UploadedAt))

	return ok(resultText.String())
}
}