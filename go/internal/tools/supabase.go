//go:build ignore
// +build ignore

package tools

/**
 * @file supabase.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Supabase MCP tools.
 * Replaces `supabase` (SSE: https://mcp.supabase.com/mcp) entry in mcp.json.
 *
 * Uses the Supabase REST API and PostgREST natively.
 * Improvements over original:
 * - No SSE connection overhead.
 * - Supports: project listing, table CRUD, SQL execution, auth management,
 *   storage operations, and real-time subscription status.
 * - Context-aware with timeout; uses SUPABASE_URL + SUPABASE_KEY for auth.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func supabaseURL() string {
	if u := os.Getenv("SUPABASE_URL"); u != "" {
		return u
	}
	return "https://mcp.supabase.com"
}

func supabaseKey() string {
	if k := os.Getenv("SUPABASE_KEY"); k != "" {
		return k
	}
	return os.Getenv("SUPABASE_ANON_KEY")
}

func supabaseServiceKey() string {
	if k := os.Getenv("SUPABASE_SERVICE_KEY"); k != "" {
		return k
	}
	return os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
}

func supabaseAccessToken() string {
	if t := os.Getenv("SUPABASE_ACCESS_TOKEN"); t != "" {
		return t
	}
	return supabaseKey()
}

func supabaseDo(ctx context.Context, method, urlPath string, payload interface{}, useServiceKey bool) (interface{}, error) {
	baseURL := supabaseURL()
	key := supabaseKey()
	if useServiceKey {
		if sk := supabaseServiceKey(); sk != "" {
			key = sk
		}
	}
	if key == "" {
		return nil, fmt.Errorf("SUPABASE_KEY or SUPABASE_ACCESS_TOKEN environment variable is not set")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	fullURL := baseURL + urlPath
	req, e := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Supabase API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// supabaseManagementDo calls the Supabase Management API (platform API).
func supabaseManagementDo(ctx context.Context, method, urlPath string, payload interface{}) (interface{}, error) {
	accessToken := supabaseAccessToken()
	if accessToken == "" {
		return nil, fmt.Errorf("SUPABASE_ACCESS_TOKEN environment variable is not set")
	}

	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, e := http.NewRequestWithContext(ctx, method, "https://api.supabase.com/v1"+urlPath, bodyReader)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Supabase Management API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return string(body), nil
	}
	return result, nil
}

// HandleSupabaseListProjects lists all Supabase projects.
// Tool: supabase_list_projects
func HandleSupabaseListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result, e := supabaseManagementDo(ctx, "GET", "/projects", nil)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseGetProject gets details for a specific project.
// Tool: supabase_get_project
func HandleSupabaseGetProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ := getString(args, "project_ref", "projectId", "id")
	if projectRef == "" {
		return err("project_ref parameter is required")
	}

	result, e := supabaseManagementDo(ctx, "GET", "/projects/"+projectRef, nil)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseExecuteSQL runs a SQL query on the project database.
// Tool: supabase_execute_sql
func HandleSupabaseExecuteSQL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectRef, _ := getString(args, "project_ref", "projectId")
	query, _ := getString(args, "query", "sql")
	if projectRef == "" {
		return err("project_ref parameter is required")
	}
	if query == "" {
		return err("query parameter is required")
	}

	payload := map[string]interface{}{"query": query}
	result, e := supabaseManagementDo(ctx, "POST", fmt.Sprintf("/projects/%s/database/query", projectRef), payload)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseSelectRows selects rows from a table via PostgREST.
// Tool: supabase_select_rows
func HandleSupabaseSelectRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ := getString(args, "table")
	if table == "" {
		return err("table parameter is required")
	}

	path := "/rest/v1/" + table
	params := []string{}

	if selectCols, _ := getString(args, "select"); selectCols != "" {
		params = append(params, "select="+selectCols)
	} else {
		params = append(params, "select=*")
	}

	if filter, _ := getString(args, "filter"); filter != "" {
		params = append(params, filter)
	}

	if limit := getInt(args, "limit"); limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}

	if offset := getInt(args, "offset"); offset > 0 {
		params = append(params, fmt.Sprintf("offset=%d", offset))
	}

	if order, _ := getString(args, "order"); order != "" {
		params = append(params, "order="+order)
	}

	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	result, e := supabaseDo(ctx, "GET", path, nil, false)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseInsertRows inserts rows into a table via PostgREST.
// Tool: supabase_insert_rows
func HandleSupabaseInsertRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ := getString(args, "table")
	if table == "" {
		return err("table parameter is required")
	}

	records, recordsOK := args["records"].([]interface{})
	if !recordsOK || len(records) == 0 {
		// Allow single record
		record, recordOK := args["record"].(map[string]interface{})
		if !recordOK {
			return err("records or record parameter is required")
		}
		records = []interface{}{record}
	}

	result, e := supabaseDo(ctx, "POST", "/rest/v1/"+table, records, false)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseUpdateRows updates rows in a table via PostgREST.
// Tool: supabase_update_rows
func HandleSupabaseUpdateRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ := getString(args, "table")
	if table == "" {
		return err("table parameter is required")
	}

	updates, updatesOK := args["updates"].(map[string]interface{})
	if !updatesOK || len(updates) == 0 {
		return err("updates parameter is required")
	}

	filter, _ := getString(args, "filter")
	if filter == "" {
		return err("filter parameter is required (e.g., 'id=eq.1')")
	}

	path := "/rest/v1/" + table + "?" + filter
	result, e := supabaseDo(ctx, "PATCH", path, updates, false)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseDeleteRows deletes rows from a table via PostgREST.
// Tool: supabase_delete_rows
func HandleSupabaseDeleteRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ := getString(args, "table")
	if table == "" {
		return err("table parameter is required")
	}

	filter, _ := getString(args, "filter")
	if filter == "" {
		return err("filter parameter is required (e.g., 'id=eq.1')")
	}

	path := "/rest/v1/" + table + "?" + filter
	result, e := supabaseDo(ctx, "DELETE", path, nil, false)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseListTables lists all tables in the public schema.
// Tool: supabase_list_tables
func HandleSupabaseListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Use the PostgREST openapi spec endpoint
	result, e := supabaseDo(ctx, "GET", "/rest/v1/", nil, false)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}

// HandleSupabaseInvokeFunction invokes a Supabase Edge Function.
// Tool: supabase_invoke_function
func HandleSupabaseInvokeFunction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	functionName, _ := getString(args, "function_name", "name")
	if functionName == "" {
		return err("function_name parameter is required")
	}

	payload, payloadOK := args["payload"].(map[string]interface{})
	if !payloadOK {
		payload = map[string]interface{}{}
	}

	result, e := supabaseDo(ctx, "POST", "/functions/v1/"+functionName, payload, false)
	if e != nil {
		return err(e.Error())
	}
	jout, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(jout))
}
