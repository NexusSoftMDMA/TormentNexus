package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearchPartners(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	db, _ :=getString(args, "db")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "password")
	if url == "" || db == "" || user == "" || pass == "" {
		return err("missing required arguments: url, db, user, password")
}

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "call",
		"params": map[string]interface{}{
			"service":    "object",
			"method":     "execute_kw",
			"args":       []interface{}{db, 1, pass, "res.partner", "search", []interface{}{[]interface{}{}}},
			"kwargs":     map[string]interface{}{"limit": 5},
		},
	}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post(url+"/jsonrpc", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["error"] != nil {
		return err("odoo error: " + fmt.Sprintf("%v", result["error"]))
}

	return ok("found partners: " + fmt.Sprintf("%v", result["result"]))
}

func HandleCreatePartner(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	db, _ :=getString(args, "db")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "password")
	name, _ :=getString(args, "name")
	if url == "" || db == "" || user == "" || pass == "" || name == "" {
		return err("missing required arguments: url, db, user, password, name")
}

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "call",
		"params": map[string]interface{}{
			"service":    "object",
			"method":     "execute_kw",
			"args":       []interface{}{db, 1, pass, "res.partner", "create", []interface{}{map[string]interface{}{"name": name}}},
		},
	}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post(url+"/jsonrpc", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["error"] != nil {
		return err("odoo error: " + fmt.Sprintf("%v", result["error"]))
}

	return success("created partner with id: " + fmt.Sprintf("%v", result["result"]))
}