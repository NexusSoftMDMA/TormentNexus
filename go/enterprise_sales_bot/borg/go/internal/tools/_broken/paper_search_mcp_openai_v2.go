package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// internal HTTP client with timeout
var http.DefaultClient = http.DefaultClient

// openAI request/response structures
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float32         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

// callOpenAI sends a prompt to the OpenAI chat completion endpoint and returns the model's reply.
func callOpenAI(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
}

	reqBody := openAIRequest{
		Model:    "gpt-3.5-turbo-1106",
		Messages: []openAIMessage{{Role: "user", Content: prompt}},
	}
	payload, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return "", jsonErr
	}

	endpoint := "https://api.openai.com/v1/chat/completions"
	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if reqErr != nil {
		return "", reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return "", apiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openAI API returned status %d", resp.StatusCode)
}

	var apiResp openAIResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&apiResp)
	if decodeErr != nil {
		return "", decodeErr
	}
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
}

	return apiResp.Choices[0].Message.Content, nil
}

// HandleSearchPapers searches for academic papers using OpenAI.
// Expected args:
// - query (string): the search query.
// - top_k (int, optional): number of papers to return (default 5).
func HandleSearchPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
}

	topK, _ :=getInt(args, "top_k")
	if topK <= 0 {
		topK = 5
	}

	prompt := fmt.Sprintf(`You are an academic research assistant. Provide a list of up to %d academic papers related to the query "%s". For each paper, give:
- Title
- Authors
- Year
- A concise abstract (2-3 sentences)

Return the list in plain text, each paper separated by a blank line.`, topK, query)

	response, apiErr := callOpenAI(prompt)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(response)
}

// HandleSummarizePaper asks OpenAI to summarize a given paper.
// Expected args:
// - title (string): title of the paper.
// - abstract (string): abstract of the paper.
// - max_words (int, optional): maximum words for the summary (default 150).
func HandleSummarizePaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	abstract, _ :=getString(args, "abstract")
	if title == "" || abstract == "" {
		return err("missing required arguments: title and/or abstract")
}

	maxWords, _ :=getInt(args, "max_words")
	if maxWords <= 0 {
		maxWords = 150
	}

	prompt := fmt.Sprintf(`Summarize the following academic paper in no more than %d words.

Title: %s

Abstract: %s

Provide a clear, concise summary suitable for a non‑specialist audience.`, maxWords, title, abstract)

	summary, apiErr := callOpenAI(prompt)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(summary)
}

// HandleGetPaperMetadata asks OpenAI to generate metadata for a paper based on a short description.
// Expected args:
// - description (string): brief description of the paper's content.
func HandleGetPaperMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	desc, _ :=getString(args, "description")
	if desc == "" {
		return err("missing required argument: description")
}

	prompt := fmt.Sprintf(`Based on the following description of an academic paper, generate structured metadata in JSON format with the fields:
{
  "title": string,
  "authors": [string],
  "year": integer,
  "journal": string,
  "doi": string
}

Description: %s`, desc)

	metaJSON, apiErr := callOpenAI(prompt)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(metaJSON)
}

// HandlePaperCitation generates a citation string for a paper.
// Expected args:
// - title (string)
// - authors (string) – comma‑separated list
// - year (int)
// - journal (string, optional)
// - doi (string, optional)
// - style (string, optional) – e.g., "APA", "MLA" (default "APA")
func HandlePaperCitation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	authors, _ :=getString(args, "authors")
	year, _ :=getInt(args, "year")
	if title == "" || authors == "" || year <= 0 {
		return err("missing required arguments: title, authors, and/or year")
}

	journal, _ :=getString(args, "journal")
	doi, _ :=getString(args, "doi")
	style, _ :=getString(args, "style")
	if style == "" {
		style = "APA"
	}

	prompt := fmt.Sprintf(`Create a %s citation for the following paper.

Title: %s
Authors: %s
Year: %d
Journal: %s
DOI: %s

Provide only the citation string.`, style, title, authors, year, journal, doi)

	citation, apiErr := callOpenAI(prompt)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(citation)
}