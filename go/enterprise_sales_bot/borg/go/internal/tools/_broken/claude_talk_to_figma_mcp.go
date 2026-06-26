package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var http.DefaultClient = http.DefaultClient

func getFileKey(args map[string]interface{}) string {
	if key, found := args["fileKey"].(string); ok && key != "" {
		return key
	}
	return os.Getenv("FIGMA_FILE_KEY")
}

func figmaGet(ctx context.Context, path string) ([]byte, error) {
	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("FIGMA_ACCESS_TOKEN not set")
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.figma.com/v1"+path, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("create request: %w", reqErr)
}

	req.Header.Set("X-Figma-Token", token)
	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return nil, fmt.Errorf("request failed: %w", apiErr)
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Figma API error: %s %s", resp.Status, string(body))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return nil, fmt.Errorf("read response: %w", e)
}

	return body, nil
}

func HandleGetDocumentInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey := getFileKey(args)
	if fileKey == "" {
		return err("fileKey is required (provide in args or set FIGMA_FILE_KEY env)")
}

	body, apiErr := figmaGet(ctx, "/files/"+fileKey)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

func HandleGetNodeInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey := getFileKey(args)
	if fileKey == "" {
		return err("fileKey is required")
}

	nodeID, _ :=getString(args, "nodeId")
	if nodeID == "" {
		return err("nodeId is required")
}

	depth, _ :=getInt(args, "depth")
	if depth < 1 {
		depth = 1
	}
	q := url.Values{}
	q.Set("ids", nodeID)
	q.Set("depth", strconv.Itoa(depth))
	path := "/files/" + fileKey + "/nodes?" + q.Encode()
	body, apiErr := figmaGet(ctx, path)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

func HandleGetPages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey := getFileKey(args)
	if fileKey == "" {
		return err("fileKey is required")
}

	body, apiErr := figmaGet(ctx, "/files/"+fileKey)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var doc struct {
		Name     string `json:"name"`
		Document struct {
			Children []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"children"`
		} `json:"document"`
		LastModified string `json:"lastModified"`
	}
	if parseErr := json.Unmarshal(body, &doc); parseErr != nil {
		return err("parse document: " + parseErr.Error())
}

	pages := make([]map[string]interface{}, 0)
	for _, child := range doc.Document.Children {
		if child.Type == "CANVAS" {
			pages = append(pages, map[string]interface{}{
				"id":   child.ID,
				"name": child.Name,
			})

	}
	resp := map[string]interface{}{
		"fileKey":      fileKey,
		"documentName": doc.Name,
		"lastModified": doc.LastModified,
		"pages":        pages,
	}
	respJSON, marshalErr := json.MarshalIndent(resp, "", "  ")
	if marshalErr != nil {
		return err("marshal response: " + marshalErr.Error())
}

	return ok(string(respJSON))
}

}

func HandleGetStyles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey := getFileKey(args)
	if fileKey == "" {
		return err("fileKey is required")
}

	body, apiErr := figmaGet(ctx, "/files/"+fileKey+"/styles")
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}

func HandleGetLocalComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey := getFileKey(args)
	if fileKey == "" {
		return err("fileKey is required")
}

	body, apiErr := figmaGet(ctx, "/files/"+fileKey+"/components")
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(string(body))
}