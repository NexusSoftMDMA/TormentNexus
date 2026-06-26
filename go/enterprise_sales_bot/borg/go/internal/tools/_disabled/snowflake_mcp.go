package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
)

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	account := os.Getenv("SNOWFLAKE_ACCOUNT")
	user := os.Getenv("SNOWFLAKE_USER")
	password := os.Getenv("SNOWFLAKE_PASSWORD")
	if account == "" || user == "" || password == "" {
		return err("SNOWFLAKE_ACCOUNT, SNOWFLAKE_USER, SNOWFLAKE_PASSWORD required")
	}
	url := "https://" + account + ".snowflakecomputing.com/api/v2/statements"
	body := map[string]interface{}{
		"statement": query,
		"warehouse": os.Getenv("SNOWFLAKE_WAREHOUSE"),
		"database":  os.Getenv("SNOWFLAKE_DATABASE"),
		"schema":    os.Getenv("SNOWFLAKE_SCHEMA"),
	}
	jsonBytes, e := json.Marshal(body)
	if e != nil {
		return err("json marshal: " + e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
	if e != nil {
		return err("new request: " + e.Error())
	}
	auth := base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode response: " + e.Error())
	}
	if resp.StatusCode != 200 {
		return err("snowflake error: " + toString(result))
	}
	return success(toString(result))
}

func toString(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}