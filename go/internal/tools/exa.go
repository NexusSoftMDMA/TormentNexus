package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const exaBaseURL = "https://api.exa.ai"

func getExaAPIKey() string {
	if key := os.Getenv("EXA_API_KEY"); key != "" {
		return key
	}
	return ""
}

func makeExaRequest(ctx context.Context, endpoint string, body map[string]interface{}) (map[string]interface{}, error) {
	apiKey := getExaAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("EXA_API_KEY not set")
	}

	jsonBody, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		return nil, marshalErr
	}

	reqURL := exaBaseURL + endpoint
	req, reqErr := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(jsonBody)))
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return nil, fetchErr
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exa API error: %s (status %d)", string(respBody), resp.StatusCode)
	}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

// HandleExaSearch performs a web search using the Exa API
func HandleExaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query is required")
	}

	body := map[string]interface{}{
		"query": query,
	}

	if numResults, ok := getInt(args, "num_results"); ok && numResults > 0 {
		body["numResults"] = numResults
	} else {
		body["numResults"] = 10
	}

	if includeDomains, ok := getString(args, "include_domains"); ok && includeDomains != "" {
		body["includeDomains"] = strings.Split(includeDomains, ",")
	}

	if excludeDomains, ok := getString(args, "exclude_domains"); ok && excludeDomains != "" {
		body["excludeDomains"] = strings.Split(excludeDomains, ",")
	}

	if startPublishedDate, ok := getString(args, "start_published_date"); ok && startPublishedDate != "" {
		body["startPublishedDate"] = startPublishedDate
	}

	if endPublishedDate, ok := getString(args, "end_published_date"); ok && endPublishedDate != "" {
		body["endPublishedDate"] = endPublishedDate
	}

	if useAutoprompt, ok := getBool(args, "use_autoprompt"); ok && useAutoprompt {
		body["useAutoprompt"] = true
	}

	if typeStr, ok := getString(args, "type"); ok && typeStr != "" {
		body["type"] = typeStr
	}

	result, apiErr := makeExaRequest(ctx, "/search", body)
	if apiErr != nil {
		return err(apiErr.Error())
	}

	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	return ok(string(resultJSON))
}

// HandleExaFindSimilar finds similar pages to a given URL using the Exa API
func HandleExaFindSimilar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ := getString(args, "url")
	if urlStr == "" {
		return err("url is required")
	}

	body := map[string]interface{}{
		"url": urlStr,
	}

	if numResults, ok := getInt(args, "num_results"); ok && numResults > 0 {
		body["numResults"] = numResults
	} else {
		body["numResults"] = 10
	}

	if includeDomains, ok := getString(args, "include_domains"); ok && includeDomains != "" {
		body["includeDomains"] = strings.Split(includeDomains, ",")
	}

	if excludeDomains, ok := getString(args, "exclude_domains"); ok && excludeDomains != "" {
		body["excludeDomains"] = strings.Split(excludeDomains, ",")
	}

	if startPublishedDate, ok := getString(args, "start_published_date"); ok && startPublishedDate != "" {
		body["startPublishedDate"] = startPublishedDate
	}

	if endPublishedDate, ok := getString(args, "end_published_date"); ok && endPublishedDate != "" {
		body["endPublishedDate"] = endPublishedDate
	}

	result, apiErr := makeExaRequest(ctx, "/findSimilar", body)
	if apiErr != nil {
		return err(apiErr.Error())
	}

	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	return ok(string(resultJSON))
}

// HandleExaGetContents retrieves the contents of pages by ID using the Exa API
func HandleExaGetContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	idsStr, _ := getString(args, "ids")
	if idsStr == "" {
		return err("ids is required")
	}

	ids := strings.Split(idsStr, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	body := map[string]interface{}{
		"ids": ids,
	}

	if text, ok := getBool(args, "text"); ok && text {
		body["text"] = true
	}

	if highlights, ok := getBool(args, "highlights"); ok && highlights {
		body["highlights"] = true
	}

	if summary, ok := getBool(args, "summary"); ok && summary {
		body["summary"] = true
	}

	if livecrawl, ok := getString(args, "livecrawl"); ok && livecrawl != "" {
		body["livecrawl"] = livecrawl
	}

	result, apiErr := makeExaRequest(ctx, "/contents", body)
	if apiErr != nil {
		return err(apiErr.Error())
	}

	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	return ok(string(resultJSON))
}

// HandleExaAnswer generates an answer to a query using the Exa API
func HandleExaAnswer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query is required")
	}

	body := map[string]interface{}{
		"query": query,
	}

	if model, ok := getString(args, "model"); ok && model != "" {
		body["model"] = model
	}

	if includeDomains, ok := getString(args, "include_domains"); ok && includeDomains != "" {
		body["includeDomains"] = strings.Split(includeDomains, ",")
	}

	if excludeDomains, ok := getString(args, "exclude_domains"); ok && excludeDomains != "" {
		body["excludeDomains"] = strings.Split(excludeDomains, ",")
	}

	if startPublishedDate, ok := getString(args, "start_published_date"); ok && startPublishedDate != "" {
		body["startPublishedDate"] = startPublishedDate
	}

	if endPublishedDate, ok := getString(args, "end_published_date"); ok && endPublishedDate != "" {
		body["endPublishedDate"] = endPublishedDate
	}

	if numResults, ok := getInt(args, "num_results"); ok && numResults > 0 {
		body["numResults"] = numResults
	}

	result, apiErr := makeExaRequest(ctx, "/answer", body)
	if apiErr != nil {
		return err(apiErr.Error())
	}

	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return err(marshalErr.Error())
	}

	return ok(string(resultJSON))
}