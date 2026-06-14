package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleGetRecord(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceURL, _ :=getString(args, "instance_url")
	module, _ :=getString(args, "module")
	id, _ :=getString(args, "id")
	if instanceURL == "" || module == "" || id == "" {
		return err("missing required parameters: instance_url, module, id")
}

	url := strings.TrimRight(instanceURL, "/") + "/webservice.php?operation=retrieve&sessionName=&id=" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("JSON decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok("Record retrieved: " + string(data))
}

func HandleListRecords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceURL, _ :=getString(args, "instance_url")
	module, _ :=getString(args, "module")
	query, _ :=getString(args, "query")
	if instanceURL == "" || module == "" {
		return err("missing required parameters: instance_url, module")
}

	url := strings.TrimRight(instanceURL, "/") + "/webservice.php?operation=query&sessionName=&query=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("JSON decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok("Records listed: " + string(data))
}