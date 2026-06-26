package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HandleSearchModels searches Hugging Face models
func HandleSearchModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	client := http.DefaultClient
	params := url.Values{}
	params.Set("search", query)
	params.Set("limit", strconv.Itoa(limit))

	reqURL := fmt.Sprintf("https://huggingface.co/api/models?%s", params.Encode())
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("API request failed: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

	var results []map[string]interface{}
	if parseErr := json.Unmarshal(body, &results); parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	if len(results) == 0 {
		return ok("No models found matching the query.")
}

	var output []string
	for i, model := range results {
		if i >= 10 { // Limit output for readability
			break
		}
		name, _ := model["modelId"].(string)
		author, _ := model["author"].(string)
		output = append(output, fmt.Sprintf("%s (by %s)", name, author))

	return ok(fmt.Sprintf("Found %d models (showing top %d):\n%s",
}
		len(results), len(output), strings.Join(output, "\n")))

}

// HandleGetModelInfo retrieves detailed information about a specific model
func HandleGetModelInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ :=getString(args, "model_id")
	if modelID == "" {
		return err("model_id parameter is required")
}

	// Validate model ID format
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+$`, modelID)
	if !matched {
		return err("model_id must be in format 'owner/model-name'")
}

	client := http.DefaultClient
	reqURL := fmt.Sprintf("https://huggingface.co/api/models/%s", url.PathEscape(modelID))

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("API request failed: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode == http.StatusNotFound {
		return err(fmt.Sprintf("Model not found: %s", modelID))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

	var modelInfo map[string]interface{}
	if parseErr := json.Unmarshal(body, &modelInfo); parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	// Extract useful information
	name, _ := modelInfo["modelId"].(string)
	author, _ := modelInfo["author"].(string)
	pipeline, _ := modelInfo["pipeline_tag"].(string)
	downloads, _ := modelInfo["downloads"].(float64)
	likes, _ := modelInfo["likes"].(float64)

	result := fmt.Sprintf("Model: %s\nAuthor: %s\nPipeline: %s\nDownloads: %d\nLikes: %d",
		name, author, pipeline, int64(downloads), int64(likes))

	if tags, found := modelInfo["tags"].([]interface{}); ok && len(tags) > 0 {
		tagList := make([]string, 0)
		for _, tag := range tags {
			if tagStr, found := tag.(string); found {
				tagList = append(tagList, tagStr)

		}
		if len(tagList) > 0 {
			result += fmt.Sprintf("\nTags: %s", strings.Join(tagList, ", "))

	}

	return ok(result)
}

}
}

// HandleSearchDatasets searches Hugging Face datasets
func HandleSearchDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	client := http.DefaultClient
	params := url.Values{}
	params.Set("search", query)
	params.Set("limit", strconv.Itoa(limit))

	reqURL := fmt.Sprintf("https://huggingface.co/api/datasets?%s", params.Encode())
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("API request failed: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

	var results []map[string]interface{}
	if parseErr := json.Unmarshal(body, &results); parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	if len(results) == 0 {
		return ok("No datasets found matching the query.")
}

	var output []string
	for i, dataset := range results {
		if i >= 10 {
			break
		}
		name, _ := dataset["id"].(string)
		author, _ := dataset["author"].(string)
		output = append(output, fmt.Sprintf("%s (by %s)", name, author))

	return ok(fmt.Sprintf("Found %d datasets (showing top %d):\n%s",
}
		len(results), len(output), strings.Join(output, "\n")))

}

// HandleGetDatasetInfo retrieves detailed information about a specific dataset
func HandleGetDatasetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	datasetID, _ :=getString(args, "dataset_id")
	if datasetID == "" {
		return err("dataset_id parameter is required")
}

	// Validate dataset ID format
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+$`, datasetID)
	if !matched {
		return err("dataset_id must be in format 'owner/dataset-name'")
}

	client := http.DefaultClient
	reqURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s", url.PathEscape(datasetID))

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("API request failed: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode == http.StatusNotFound {
		return err(fmt.Sprintf("Dataset not found: %s", datasetID))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

	var datasetInfo map[string]interface{}
	if parseErr := json.Unmarshal(body, &datasetInfo); parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	name, _ := datasetInfo["id"].(string)
	author, _ := datasetInfo["author"].(string)
	likes, _ := datasetInfo["likes"].(float64)

	result := fmt.Sprintf("Dataset: %s\nAuthor: %s\nLikes: %d",
		name, author, int64(likes))

	if tags, found := datasetInfo["tags"].([]interface{}); ok && len(tags) > 0 {
		tagList := make([]string, 0)
		for _, tag := range tags {
			if tagStr, found := tag.(string); found {
				tagList = append(tagList, tagStr)

		}
		if len(tagList) > 0 {
			result += fmt.Sprintf("\nTags: %s", strings.Join(tagList, ", "))

	}

	return ok(result)
}

}
}

// HandleListSpaces lists public spaces from Hugging Face
func HandleListSpaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	author, _ :=getString(args, "author")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	client := http.DefaultClient
	params := url.Values{}
	if author != "" {
		params.Set("author", author)

	params.Set("limit", strconv.Itoa(limit))
	params.Set("sort", "likes")
	params.Set("direction", "-1")

	reqURL := fmt.Sprintf("https://huggingface.co/api/spaces?%s", params.Encode())
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("API request failed: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

	var results []map[string]interface{}
	if parseErr := json.Unmarshal(body, &results); parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	if len(results) == 0 {
		return ok("No spaces found.")
}

	var output []string
	for i, space := range results {
		if i >= 10 {
			break
		}
		name, _ := space["id"].(string)
		author, _ := space["author"].(string)
		sdk, _ := space["sdk"].(string)
		output = append(output, fmt.Sprintf("%s by %s (SDK: %s)", name, author, sdk))

	return ok(fmt.Sprintf("Found %d spaces (showing top %d):\n%s",
}
		len(results), len(output), strings.Join(output, "\n")))
}
}