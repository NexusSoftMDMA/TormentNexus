package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func HandleExaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required parameter: query")
}

	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return err("EXA_API_KEY environment variable not set")
}

	typ, _ :=getString(args, "type")
	if typ == "" {
		typ = "neural"
	}
	numResults, _ :=getInt(args, "numResults")
	if numResults <= 0 {
		numResults = 10
	}
	useAutoparse, _ :=getBool(args, "useAutoparse")
	if !useAutoparse {
		useAutoparse = true
	}
	contentsText, _ :=getString(args, "contentsText")
	contentsTextBool := false
	if contentsText != "" {
		contentsTextBool = true
	}
	contentsURL, _ :=getString(args, "contentsURL")
	contentsURLBool := false
	if contentsURL != "" {
		contentsURLBool = true
	}
	reqBody := map[string]interface{}{
		"query":         query,
		"type":          typ,
		"numResults":    numResults,
		"useAutoparse":  useAutoparse,
		"contents": map[string]interface{}{
			"text": map[string]interface{}{
				"maxCharacters": 1000,
			},
			"url": map[string]interface{}{
				"maxCharacters": 1000,
			},
		},
	}
	bodyBytes, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.exa.ai/search", strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("exa API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return err(decErr.Error())
}

	resultsRaw, found := result["results"].([]interface{})
	if !found {
		return err("unexpected response format: missing results")
}

	var sb strings.Builder
	for i, r := range resultsRaw {
		item, found := r.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := item["title"].(string)
		resURL, _ := item["url"].(string)
		pubDate, _ := item["publishedDate"].(string)
		author, _ := item["author"].(string)
		score, _ := item["score"].(float64)
		highlights, _ := item["highlights"].([]interface{})
		highlightScore, _ := item["highlightScore"].(float64)
		sb.WriteString(fmt.Sprintf("--- Result %d ---\n", i+1))
		if title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", title))

		if resURL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", resURL))

		if pubDate != "" {
			sb.WriteString(fmt.Sprintf("Published: %s\n", pubDate))

		if author != "" {
			sb.WriteString(fmt.Sprintf("Author: %s\n", author))

		if score > 0 {
			sb.WriteString(fmt.Sprintf("Score: %.4f\n", score))

		if highlightScore > 0 {
			sb.WriteString(fmt.Sprintf("Highlight Score: %.4f\n", highlightScore))

		if len(highlights) > 0 {
			sb.WriteString("Highlights:\n")
			for _, h := range highlights {
				if hs, found := h.(string); found {
					sb.WriteString(fmt.Sprintf(" - %s\n", hs))

			}
		}
		sb.WriteString("\n")

	if sb.Len() == 0 {
		return ok("No results found.")
}

	return ok(sb.String())
}

}
}
}
}
}
}
}
}

func HandleExaGetContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urls, _ :=getString(args, "urls")
	if urls == "" {
		return err("missing required parameter: urls")
}

	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return err("EXA_API_KEY environment variable not set")
}

	urlList := strings.Split(urls, ",")
	for i := range urlList {
		urlList[i] = strings.TrimSpace(urlList[i])

	var filtered []string
	for _, u := range urlList {
		if u != "" {
			filtered = append(filtered, u)

	}
	if len(filtered) == 0 {
		return err("no valid URLs provided")
}

	reqBody := map[string]interface{}{
		"ids": filtered,
		"contents": map[string]interface{}{
			"text": map[string]interface{}{
				"maxCharacters": 1000,
			},
			"url": map[string]interface{}{
				"maxCharacters": 1000,
			},
		},
	}
	bodyBytes, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.exa.ai/contents", strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("exa API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return err(decErr.Error())
}

	resultsRaw, found := result["results"].([]interface{})
	if !found {
		return err("unexpected response format: missing results")
}

	var sb strings.Builder
	for i, r := range resultsRaw {
		item, found := r.(map[string]interface{})
		if !found {
			continue
		}
		resURL, _ := item["url"].(string)
		title, _ := item["title"].(string)
		author, _ := item["author"].(string)
		pubDate, _ := item["publishedDate"].(string)
		text, _ := item["text"].(string)
		sb.WriteString(fmt.Sprintf("--- Content %d ---\n", i+1))
		if title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", title))

		if resURL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", resURL))

		if pubDate != "" {
			sb.WriteString(fmt.Sprintf("Published: %s\n", pubDate))

		if author != "" {
			sb.WriteString(fmt.Sprintf("Author: %s\n", author))

		if text != "" {
			sb.WriteString(fmt.Sprintf("Text:\n%s\n", text))

		sb.WriteString("\n")

	if sb.Len() == 0 {
		return ok("No contents returned.")
}

	return ok(sb.String())
}

}
}
}
}
}
}
}
}

func HandleExaFindSimilar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlArg, _ :=getString(args, "url")
	if urlArg == "" {
		return err("missing required parameter: url")
}

	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return err("EXA_API_KEY environment variable not set")
}

	numResults, _ :=getInt(args, "numResults")
	if numResults <= 0 {
		numResults = 10
	}
	contentsDomain, _ :=getString(args, "contentsDomain")
	contentsDomainBool := false
	if contentsDomain != "" {
		contentsDomainBool = true
	}
	reqBody := map[string]interface{}{
		"url":        urlArg,
		"numResults": numResults,
		"contents": map[string]interface{}{
			"text": map[string]interface{}{
				"maxCharacters": 1000,
			},
			"url": map[string]interface{}{
				"maxCharacters": 1000,
			},
		},
	}
	bodyBytes, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.exa.ai/findSimilar", strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("exa API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return err(decErr.Error())
}

	resultsRaw, found := result["results"].([]interface{})
	if !found {
		return err("unexpected response format: missing results")
}

	var sb strings.Builder
	for i, r := range resultsRaw {
		item, found := r.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := item["title"].(string)
		resURL, _ := item["url"].(string)
		pubDate, _ := item["publishedDate"].(string)
		author, _ := item["author"].(string)
		score, _ := item["score"].(float64)
		sb.WriteString(fmt.Sprintf("--- Similar %d ---\n", i+1))
		if title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", title))

		if resURL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", resURL))

		if pubDate != "" {
			sb.WriteString(fmt.Sprintf("Published: %s\n", pubDate))

		if author != "" {
			sb.WriteString(fmt.Sprintf("Author: %s\n", author))

		if score > 0 {
			sb.WriteString(fmt.Sprintf("Score: %.4f\n", score))

		sb.WriteString("\n")

	if sb.Len() == 0 {
		return ok("No similar results found.")
}

	return ok(sb.String())
}

}
}
}
}
}
}

func HandleExaAnswer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required parameter: query")
}

	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return err("EXA_API_KEY environment variable not set")
}

	urlArg, _ :=getString(args, "url")
	urls, _ :=getString(args, "urls")
	reqBody := map[string]interface{}{
		"query": query,
	}
	if urlArg != "" {
		reqBody["url"] = urlArg
	}
	if urls != "" {
		urlList := strings.Split(urls, ",")
		for i := range urlList {
			urlList[i] = strings.TrimSpace(urlList[i])

		var filtered []string
		for _, u := range urlList {
			if u != "" {
				filtered = append(filtered, u)

		}
		if len(filtered) > 0 {
			reqBody["urls"] = filtered
		}
	}
	reqBody["contents"] = map[string]interface{}{
		"text": map[string]interface{}{
			"maxCharacters": 1000,
		},
		"url": map[string]interface{}{
			"maxCharacters": 1000,
		},
	}
	bodyBytes, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", "https://api.exa.ai/answer", strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("exa API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return err(decErr.Error())
}

	answer, _ := result["answer"].(string)
	if answer == "" {
		return ok("No answer returned.")
}

	return ok(answer)
}

}
}

func HandleExaCustomSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required parameter: query")
}

	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return err("EXA_API_KEY environment variable not set")
}

	params := url.Values{}
	params.Set("query", query)
	params.Set("type", "neural")
	params.Set("contents", `{"text":{"maxCharacters":1000},"url":{"maxCharacters":1000}}`)
	typ, _ :=getString(args, "type")
	if typ != "" {
		params.Set("type", typ)

	numResults, _ :=getInt(args, "numResults")
	if numResults > 0 {
		params.Set("numResults", strconv.Itoa(numResults))

	useAutoparse, _ :=getBool(args, "useAutoparse")
	if useAutoparse {
		params.Set("useAutoparse", "true")

	startDate, _ :=getString(args, "startDate")
	if startDate != "" {
		params.Set("startPublishedDate", startDate)

	endDate, _ :=getString(args, "endDate")
	if endDate != "" {
		params.Set("endPublishedDate", endDate)

	includeDomains, _ :=getString(args, "includeDomains")
	if includeDomains != "" {
		params.Set("includeDomains", includeDomains)

	excludeDomains, _ :=getString(args, "excludeDomains")
	if excludeDomains != "" {
		params.Set("excludeDomains", excludeDomains)

	category, _ :=getString(args, "category")
	if category != "" {
		params.Set("category", category)

	near, _ :=getString(args, "near")
	if near != "" {
		params.Set("near", near)

	endpoint := "https://api.exa.ai/search?" + params.Encode()
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)
	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("exa API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return err(decErr.Error())
}

	resultsRaw, found := result["results"].([]interface{})
	if !found {
		return err("unexpected response format: missing results")
}

	var sb strings.Builder
	for i, r := range resultsRaw {
		item, found := r.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := item["title"].(string)
		resURL, _ := item["url"].(string)
		pubDate, _ := item["publishedDate"].(string)
		author, _ := item["author"].(string)
		score, _ := item["score"].(float64)
		sb.WriteString(fmt.Sprintf("--- Result %d ---\n", i+1))
		if title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", title))

		if resURL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", resURL))

		if pubDate != "" {
			sb.WriteString(fmt.Sprintf("Published: %s\n", pubDate))

		if author != "" {
			sb.WriteString(fmt.Sprintf("Author: %s\n", author))

		if score > 0 {
			sb.WriteString(fmt.Sprintf("Score: %.4f\n", score))

		sb.WriteString("\n")

	if sb.Len() == 0 {
		return ok("No results found.")
}

	return ok(sb.String())
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}