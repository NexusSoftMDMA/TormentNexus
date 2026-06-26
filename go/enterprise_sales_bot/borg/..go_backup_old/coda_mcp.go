package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	if token == "" {
		return err("api_token required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.coda.io/v1/docs", nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return err("no items in response")
}

	var docs []string
	for _, item := range items {
		doc, found := item.(map[string]interface{})
		if !found {
			continue
		}
		id, _ := doc["id"].(string)
		name, _ := doc["name"].(string)
		docs = append(docs, fmt.Sprintf("%s: %s", id, name))

	if len(docs) == 0 {
		return ok("no docs found")
}

	out, _ := json.Marshal(docs)
	return success(string(out))
}

}

func HandleGetDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	docID, _ :=getString(args, "doc_id")
	if token == "" || docID == "" {
		return err("api_token and doc_id required")
}

	url := fmt.Sprintf("https://api.coda.io/v1/docs/%s", docID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e = json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	out, _ := json.Marshal(result)
	return success(string(out))
}