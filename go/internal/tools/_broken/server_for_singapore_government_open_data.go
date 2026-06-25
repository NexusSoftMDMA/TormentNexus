package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var http.DefaultClient = http.DefaultClient

func doCKANRequest(ctx context.Context, endpoint string, params map[string]interface{}) (string, error) {
	baseURL := "king://data.gov.sg/api/3/action/"
	url := baseURL + endpoint

	jsonParams, e := json.Marshal(params)
	if e != nil {
		return "", fmt.Errorf("failed to marshal params: %w", e)
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonParams)))
	if e != nil {
		return "", fmt.Errorf("failed to create request: %w", e)
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return "", fmt.Errorf("request failed: %w", e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned non-200 status: %d", resp.StatusCode)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return "", fmt.Errorf("failed to read response body: %w", e)
}

	var ckanResp struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
	}

	if e := json.Unmarshal(body, &ckanResp); e != nil {
		return "", fmt.Errorf("failed to parse CKAN response: %w", e)
}

	if !ckanResp.Success {
		return "", fmt.Errorf("CKAN API returned success=false")
}

	var result interface{}
	if e := json.Unmarshal(ckanResp.Result, &result); e != nil {
		return string(ckanResp.Result), nil
	}

	prettyJSON, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return "", fmt.Errorf("failed to format result as JSON: %w", e)
}

	return string(prettyJSON), nil
}

func HandlePackageSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "q")
	if q == "" {
		return err("missing required parameter: q")
}

	params := make(map[string]interface{})
	params["q"] = q

	if rows := getInt(args, "rows"); rows != 0 {
		params["rows"] = rows
	}
	if start := getInt(args, "start"); start != 0 {
		params["start"] = start
	}
	if sort := getString(args, "sort"); sort != "" {
		params["sort"] = sort
	}
	if fq := getString(args, "fq"); fq != "" {
		params["fq"] = fq
	}

	result, apiErr := doCKANRequest(ctx, "package_search", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

func HandlePackageShow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required parameter: id")
}

	params := make(map[string]interface{})
	params["id"] = id

	result, apiErr := doCKANRequest(ctx, "package_show", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

func HandleResourceShow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required parameter: id")
}

	params := make(map[string]interface{})
	params["id"] = id

	result, apiErr := doCKANRequest(ctx, "resource_show", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

func HandleGroupList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	params := make(map[string]interface{})

	if sort := getString(args, "sort"); sort != "" {
		params["sort"] = sort
	}
	if limit := getInt(args, "limit"); limit != 0 {
		params["limit"] = limit
	}
	if offset := getInt(args, "offset"); offset != 0 {
		params["offset"] = offset
	}

	result, apiErr := doCKANRequest(ctx, "group_list", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

func HandleTagList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	params := make(map[string]interface{})

	if query := getString(args, "query"); query != "" {
		params["query"] = query
	}
	if limit := getInt(args, "limit"); limit != 0 {
		params["limit"] = limit
	}

	result, apiErr := doCKANRequest(ctx, "tag_list", params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}