//go:build ignore
// +build ignore

package tools

/**
 * @file food_data_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of FoodData Central MCP — USDA food nutrition database.
 * Replaces: github.com/jlfwong/food-data-central-mcp-server
 *
 * Provides access to the USDA FoodData Central database for food search,
 * nutrient information, and branded/foundation food data.
 * No API key required (public USDA API).
 *
 * Tools:
 *  - food_search — search foods by keyword
 *  - food_get — get food details by FDC ID
 *  - food_list — list foods by data type
 */

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const foodDataURL = "https://api.nal.usda.gov/fdc/v1"

// HandleFoodSearch searches the USDA FoodData Central database.
func HandleFoodSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "food")
	if query == "" {
		return err("query is required")
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	dataType, _ := getString(args, "data_type", "type")

	params := url.Values{}
	params.Set("query", query)
	params.Set("pageSize", fmt.Sprintf("%d", limit))
	if dataType != "" {
		params.Set("dataType", dataType)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf("%s/foods/search?%s", foodDataURL, params.Encode()))
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleFoodGet gets detailed food information by FDC ID.
func HandleFoodGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := getInt(args, "id", "fdc_id", "fdcId")
	if id <= 0 {
		return err("valid FDC ID is required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf("%s/food/%d", foodDataURL, id))
	if e != nil {
		return err(fmt.Sprintf("get food failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

// HandleFoodList lists foods by data type.
func HandleFoodList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataType, _ := getString(args, "data_type", "type")
	if dataType == "" {
		dataType = "Foundation"
	}
	limit := getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Get(fmt.Sprintf("%s/foods/list?dataType=%s&pageSize=%d",
		foodDataURL, url.QueryEscape(dataType), limit))
	if e != nil {
		return err(fmt.Sprintf("list failed: %v", e))
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}
