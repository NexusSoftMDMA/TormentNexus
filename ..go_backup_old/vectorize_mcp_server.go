package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type vectorDoc struct {
	Index  string    `json:"index"`
	ID     string    `json:"id"`
	Vector []float64 `json:"vector"`
	Score  float64   `json:"score,omitempty"`
}

func HandleAddVector(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	doc := vectorDoc{
		Index: getString(args, "index"),
		ID:    getString(args, "id"),
	}
	vectorStr, _ :=getString(args, "vector")
	if e := json.Unmarshal([]byte(vectorStr), &doc.Vector); e != nil {
		return err("invalid vector JSON: " + e.Error())
}

	body, e := json.Marshal(doc)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.vectorize.example.com/vectors", bytes.NewReader(body))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(bodyBytes)))
}

	return ok("vector added")
}

func HandleSearchVector(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query := map[string]interface{}{
		"index":  getString(args, "index"),
		"vector": json.RawMessage(getString(args, "vector")),
		"top_k":  getInt(args, "top_k"),
	}
	body, e := json.Marshal(query)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.vectorize.example.com/search", bytes.NewReader(body))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(bodyBytes)))
}

	var results []vectorDoc
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("found %d results", len(results)))
}