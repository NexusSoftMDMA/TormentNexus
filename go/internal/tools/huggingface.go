//go:build ignore
// +build ignore

package tools

/**
 * @file  * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Hugging Face Hub API.
 * Replaces ` *
 * Uses the Hugging Face Hub REST API (https:// * Improvements over original:
 *  - No SSE connection overhead.
 *  - Supports: model search/details, dataset search/details, space search,
 *              inference (text generation, classification, embeddings),
 *              and model card retrieval.
 *  - Context-aware with timeout; uses HF_TOKEN for authenticated access.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	hfAPIBase       = "https://	hfInferenceBase = "https://api-inference.)

func hfToken() string {
	if t := os.Getenv("HF_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("HUGGINGFACE_TOKEN")
}

func hfGet(ctx context.Context, path string, params map[string]string) (interface{}, error) {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}

	fullURL := hfAPIBase + path
	if len(q) > 0 {
		fullURL += "?" + q.Encode()
	}

	req, e := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if e != nil {
		return nil, e
	}
	if token := hfToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "TormentNexus/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Hugging Face API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

func hfInfer(ctx context.Context, modelID string, payload interface{}) (interface{}, error) {
	token := hfToken()
	if token == "" {
		return nil, fmt.Errorf("HF_TOKEN environment variable is required for inference")
	}

	data, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", hfInferenceBase+"/"+modelID, bytes.NewBuffer(data))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Hugging Face inference error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// HandleHFSearchModels searches for models on Hugging Face Hub.
// Tool: hf_search_models
func HandleHFSearchModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")

	params := map[string]string{"limit": "20"}
	if query != "" {
		params["search"] = query
	}
	if task, _ := getString(args, "pipeline_tag", "task"); task != "" {
		params["pipeline_tag"] = task
	}
	if sort, _ := getString(args, "sort"); sort != "" {
		params["sort"] = sort
	} else {
		params["sort"] = "downloads"
	}

	result, e := hfGet(ctx, "/models", params)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleHFGetModel retrieves details about a specific model.
// Tool: hf_get_model
func HandleHFGetModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ := getString(args, "model_id", "model", "id")
	if modelID == "" {
		return err("model_id parameter is required (e.g., 'meta-llama/Llama-3.2-1B')")
	}

	result, e := hfGet(ctx, "/models/"+modelID, nil)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleHFSearchDatasets searches for datasets on Hugging Face Hub.
// Tool: hf_search_datasets
func HandleHFSearchDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")

	params := map[string]string{"limit": "20", "sort": "downloads"}
	if query != "" {
		params["search"] = query
	}

	result, e := hfGet(ctx, "/datasets", params)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleHFTextGeneration runs text generation inference on a model.
// Tool: hf_text_generation
func HandleHFTextGeneration(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ := getString(args, "model_id", "model")
	if modelID == "" {
		modelID = "gpt2"
	}

	inputText, _ := getString(args, "inputs", "text", "prompt")
	if inputText == "" {
		return err("inputs (text/prompt) parameter is required")
	}

	payload := map[string]interface{}{
		"inputs": inputText,
	}

	params := map[string]interface{}{}
	if maxLen := getInt(args, "max_new_tokens", "max_length"); maxLen > 0 {
		params["max_new_tokens"] = maxLen
	}
	if temp, ok := args["temperature"].(float64); ok {
		params["temperature"] = temp
	}
	if len(params) > 0 {
		payload["parameters"] = params
	}

	result, e := hfInfer(ctx, modelID, payload)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleHFClassification runs text classification on a model.
// Tool: hf_classify_text
func HandleHFClassification(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ := getString(args, "model_id", "model")
	if modelID == "" {
		modelID = "distilbert-base-uncased-finetuned-sst-2-english"
	}

	inputText, _ := getString(args, "inputs", "text")
	if inputText == "" {
		return err("inputs (text) parameter is required")
	}

	result, e := hfInfer(ctx, modelID, map[string]interface{}{"inputs": inputText})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleHFEmbeddings generates embeddings for text using a model.
// Tool: hf_embeddings
func HandleHFEmbeddings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelID, _ := getString(args, "model_id", "model")
	if modelID == "" {
		modelID = "sentence-transformers/all-MiniLM-L6-v2"
	}

	inputText, _ := getString(args, "inputs", "text")
	if inputText == "" {
		return err("inputs (text) parameter is required")
	}

	result, e := hfInfer(ctx, modelID, map[string]interface{}{"inputs": inputText})
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleHFSearchSpaces searches for Spaces on Hugging Face Hub.
// Tool: hf_search_spaces
func HandleHFSearchSpaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	params := map[string]string{"limit": "20", "sort": "likes"}
	if query != "" {
		params["search"] = query
	}

	result, e := hfGet(ctx, "/spaces", params)
	if e != nil {
		return err(e.Error())
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}
