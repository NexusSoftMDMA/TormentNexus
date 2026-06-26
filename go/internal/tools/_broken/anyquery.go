package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	if database == "" {
		return err("database is required")
}

	encodedDB := url.QueryEscape(database)
	reqURL := fmt.Sprintf("http://localhost:3000/api/v1/tables?database=%s", encodedDB)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleDescribeTable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	table, _ :=getString(args, "table")
	if database == "" {
		return err("database is required")
}

	if table == "" {
		return err("table is required")
}

	encodedDB := url.QueryEscape(database)
	encodedTable := url.QueryEscape(table)
	reqURL := fmt.Sprintf("http://localhost:3000/api/v1/tables/%s?database=%s", encodedTable, encodedDB)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	query, _ :=getString(args, "query")
	if database == "" {
		return err("database is required")
}

	if query == "" {
		return err("query is required")
}

	payload := map[string]interface{}{
		"database": database,
		"query":    query,
	}
	jsonPayload, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	reqURL := "http://localhost:3000/api/v1/query"
	req, reqErr := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(jsonPayload)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reqURL := "http://localhost:3000/api/v1/databases"

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}