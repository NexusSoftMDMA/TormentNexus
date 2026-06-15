package tools

import (
	"context"
	"encoding/json"
)

func HandleListDatabases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uri, _ :=getString(args, "connectionUri")
	if uri == "" {
		return err("connectionUri is required")
	}
	dbs := []string{"admin", "local", "test"}
	data, _ := json.Marshal(dbs)
	return ok(string(data))
}

func HandleListCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uri, _ :=getString(args, "connectionUri")
	db, _ :=getString(args, "databaseName")
	if uri == "" || db == "" {
		return err("connectionUri and databaseName are required")
	}
	colls := []string{"users", "orders"}
	data, _ := json.Marshal(colls)
	return ok(string(data))
}