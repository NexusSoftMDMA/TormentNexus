package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSchemacrawler(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	database, _ :=getString(args, "database")
	command, _ :=getString(args, "command")
	if database == "" {
		return err("database is required")
}

	if command == "" {
		command = "schemas"
	}
	resp, e := http.DefaultClient.Get(
		fmt.Sprintf("http://localhost:1234/schemacrawler?database=%s&command=%s", database, command),
	)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	return ok(fmt.Sprintf("Schema result: %v", result))
}