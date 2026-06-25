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

// ClickHouseConfig holds connection settings
type ClickHouseConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// getClickHouseURL builds the ClickHouse HTTP endpoint URL
func getClickHouseURL(cfg ClickHouseConfig, query string) string {
	u := url.URL{
		Scheme: "http",
		Host:   cfg.Host + ":" + cfg.Port,
		Path:   "/",
	}
	q := u.Query()
	q.Set("query", query)
	if cfg.Database != "" {
		q.Set("database", cfg.Database)

	if cfg.Username != "" {
		q.Set("user", cfg.Username)

	if cfg.Password != "" {
		q.Set("password", cfg.Password)

	u.RawQuery = q.Encode()
	return u.String()
}

}
}
}

// HandleClickHouseQuery executes a SELECT query and returns results
func HandleClickHouseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Database: getString(args, "database"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", clickHouseURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

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
		return err(fmt.Sprintf("ClickHouse error (status %d): %s", resp.StatusCode, string(body)))
}

	// Try to parse as JSON for formatted output
	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr == nil {
		formatted, formatErr := json.MarshalIndent(result, "", "  ")
		if formatErr == nil {
			return ok(string(formatted))

	}

	return ok(string(body))
}

}

// HandleClickHouseInsert inserts data into a table
func HandleClickHouseInsert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	if table == "" {
		return err("table parameter is required")
}

	format, _ :=getString(args, "format")
	if format == "" {
		format = "JSONEachRow"
	}

	data, _ :=getString(args, "data")
	if data == "" {
		return err("data parameter is required")
}

	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Database: getString(args, "database"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	query := fmt.Sprintf("INSERT INTO %s FORMAT %s", table, format)
	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", clickHouseURL, strings.NewReader(data))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

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
		return err(fmt.Sprintf("ClickHouse insert error (status %d): %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Successfully inserted data into %s", table))
}

// HandleClickHouseListTables lists all tables in the database
func HandleClickHouseListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Database: getString(args, "database"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	query := "SHOW TABLES"
	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", clickHouseURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

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
		return err(fmt.Sprintf("ClickHouse error (status %d): %s", resp.StatusCode, string(body)))
}

	tables := strings.Split(strings.TrimSpace(string(body)), "\n")
	var result []string
	for _, t := range tables {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)

	}

	formatted, formatErr := json.MarshalIndent(result, "", "  ")
	if formatErr != nil {
		return err(formatErr.Error())
}

	return ok(string(formatted))
}

}

// HandleClickHouseDescribe returns the schema of a table
func HandleClickHouseDescribe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	if table == "" {
		return err("table parameter is required")
}

	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Database: getString(args, "database"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	query := fmt.Sprintf("DESCRIBE TABLE %s", table)
	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", clickHouseURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

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
		return err(fmt.Sprintf("ClickHouse error (status %d): %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

// HandleClickHouseCreateDatabase creates a new database
func HandleClickHouseCreateDatabase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	if database == "" {
		return err("database parameter is required")
}

	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", clickHouseURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

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
		return err(fmt.Sprintf("ClickHouse error (status %d): %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Database '%s' created successfully", database))
}

// HandleClickHouseCreateTable creates a new table
func HandleClickHouseCreateTable(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	if table == "" {
		return err("table parameter is required")
}

	schema, _ :=getString(args, "schema")
	if schema == "" {
		return err("schema parameter is required (e.g., 'id UInt32, name String')")
}

	engine, _ :=getString(args, "engine")
	if engine == "" {
		engine = "MergeTree()"
	}

	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Database: getString(args, "database"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s) ENGINE = %s", table, schema, engine)
	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", clickHouseURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

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
		return err(fmt.Sprintf("ClickHouse error (status %d): %s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Table '%s' created successfully", table))
}

// HandleClickHouseSystemMetrics returns system metrics
func HandleClickHouseSystemMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cfg := ClickHouseConfig{
		Host:     getString(args, "host"),
		Port:     getString(args, "port"),
		Database: getString(args, "database"),
		Username: getString(args, "username"),
		Password: getString(args, "password"),
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "8123"
	}

	query := "SELECT * FROM system.metrics FORMAT JSON"
	clickHouseURL := getClickHouseURL(cfg, query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", clickHouseURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

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
		return err(fmt.Sprintf("ClickHouse error (status %d): %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}