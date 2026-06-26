package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const notionBaseURL = "https://api.notion.com/v1"
const notionVersion = "2022-06-28"

// getNotionAPIKey retrieves the Notion API key from environment variable NOTION_API_KEY.
func getNotionAPIKey() (string, error) {
	key := os.Getenv("NOTION_API_KEY")
	if key == "" {
		return "", fmt.Errorf("NOTION_API_KEY environment variable not set")
}

	return key, nil
}

// notionRequest performs an HTTP request to the Notion API and returns the response body.
// It sets appropriate headers for authentication and API version.
func notionRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	apiKey, apiErr := getNotionAPIKey()
	if apiErr != nil {
		return nil, apiErr
	}

	urlStr := notionBaseURL + path
	var reqBody io.Reader
	if body != nil {
		jsonBytes, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", marshalErr)
}

		reqBody = strings.NewReader(string(jsonBytes))

	req, reqErr := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionVersion)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return nil, fmt.Errorf("request failed: %w", fetchErr)
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response: %w", readErr)
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Notion API returned status %d: %s", resp.StatusCode, string(respBody))
}

	return respBody, nil
}

}

// HandleNotionSearch searches all pages and databases that have been shared with your integration.
// Arguments:
//   - query: (optional) A string to search for.
//   - filter: (optional) JSON string representing a filter object.
//   - sort: (optional) JSON string representing a sort object.
func HandleNotionSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query := strings.TrimSpace(getString(args, "query"))
	filterStr := strings.TrimSpace(getString(args, "filter"))
	sortStr := strings.TrimSpace(getString(args, "sort"))

	body := map[string]interface{}{}
	if query != "" {
		body["query"] = query
	}
	if filterStr != "" {
		var filterMap map[string]interface{}
		if parseErr := json.Unmarshal([]byte(filterStr), &filterMap); parseErr != nil {
			return err("Invalid filter JSON: " + parseErr.Error())
}

		body["filter"] = filterMap
	}
	if sortStr != "" {
		var sortObj map[string]interface{}
		if parseErr := json.Unmarshal([]byte(sortStr), &sortObj); parseErr != nil {
			return err("Invalid sort JSON: " + parseErr.Error())
}

		body["sort"] = sortObj
	}

	respBody, apiErr := notionRequest(ctx, http.MethodPost, "/search", body)
	if apiErr != nil {
		return err("Search failed: " + apiErr.Error())
}

	return ok(string(respBody))
}

// HandleNotionRetrievePage retrieves a page by its ID.
// Arguments:
//   - page_id: The UUID of the page (required).
func HandleNotionRetrievePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID := strings.TrimSpace(getString(args, "page_id"))
	if pageID == "" {
		return err("page_id is required")
}

	path := "/pages/" + pageID
	respBody, apiErr := notionRequest(ctx, http.MethodGet, path, nil)
	if apiErr != nil {
		return err("Retrieve page failed: " + apiErr.Error())
}

	return ok(string(respBody))
}

// HandleNotionCreatePage creates a new page in a database or parent page.
// Arguments:
//   - parent_type: "database_id" or "page_id" (required).
//   - parent_id: The ID of the parent (required).
//   - properties: JSON string of page properties (required).
//   - children: (optional) JSON string of block children to add.
func HandleNotionCreatePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	parentType := strings.TrimSpace(getString(args, "parent_type"))
	parentID := strings.TrimSpace(getString(args, "parent_id"))
	propertiesStr := strings.TrimSpace(getString(args, "properties"))
	childrenStr := strings.TrimSpace(getString(args, "children"))

	if parentType == "" || parentID == "" || propertiesStr == "" {
		return err("parent_type, parent_id, and properties are required")
}

	var properties map[string]interface{}
	if parseErr := json.Unmarshal([]byte(propertiesStr), &properties); parseErr != nil {
		return err("Invalid properties JSON: " + parseErr.Error())
}

	body := map[string]interface{}{
		"parent": map[string]interface{}{
			parentType: parentID,
		},
		"properties": properties,
	}

	if childrenStr != "" {
		var children []interface{}
		if parseErr := json.Unmarshal([]byte(childrenStr), &children); parseErr != nil {
			return err("Invalid children JSON: " + parseErr.Error())
}

		body["children"] = children
	}

	respBody, apiErr := notionRequest(ctx, http.MethodPost, "/pages", body)
	if apiErr != nil {
		return err("Create page failed: " + apiErr.Error())
}

	return ok(string(respBody))
}

// HandleNotionUpdatePage updates properties of a page.
// Arguments:
//   - page_id: The UUID of the page (required).
//   - properties: JSON string of page properties to update (required).
func HandleNotionUpdatePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID := strings.TrimSpace(getString(args, "page_id"))
	propertiesStr := strings.TrimSpace(getString(args, "properties"))

	if pageID == "" || propertiesStr == "" {
		return err("page_id and properties are required")
}

	var properties map[string]interface{}
	if parseErr := json.Unmarshal([]byte(propertiesStr), &properties); parseErr != nil {
		return err("Invalid properties JSON: " + parseErr.Error())
}

	body := map[string]interface{}{
		"properties": properties,
	}

	path := "/pages/" + pageID
	respBody, apiErr := notionRequest(ctx, http.MethodPatch, path, body)
	if apiErr != nil {
		return err("Update page failed: " + apiErr.Error())
}

	return ok(string(respBody))
}

// HandleNotionAppendBlockChildren appends blocks to a parent block.
// Arguments:
//   - block_id: The UUID of the parent block (required).
//   - children: JSON string of block children to append (required).
func HandleNotionAppendBlockChildren(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	blockID := strings.TrimSpace(getString(args, "block_id"))
	childrenStr := strings.TrimSpace(getString(args, "children"))

	if blockID == "" || childrenStr == "" {
		return err("block_id and children are required")
}

	var children []interface{}
	if parseErr := json.Unmarshal([]byte(childrenStr), &children); parseErr != nil {
		return err("Invalid children JSON: " + parseErr.Error())
}

	body := map[string]interface{}{
		"children": children,
	}

	path := "/blocks/" + blockID + "/children"
	respBody, apiErr := notionRequest(ctx, http.MethodPatch, path, body)
	if apiErr != nil {
		return err("Append block children failed: " + apiErr.Error())
}

	return ok(string(respBody))
}