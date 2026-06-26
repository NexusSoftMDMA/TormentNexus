package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceType, _ :=getString(args, "instance_type")
	amiID, _ :=getString(args, "ami_id")
	keyName, _ :=getString(args, "key_name")

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.aws.example.com/instances", nil)
	if e != nil {
		return err(fmt.Sprintf("create request failed: %v", e))
	}
	q := req.URL.Query()
	q.Set("instance_type", instanceType)
	q.Set("ami_id", amiID)
	q.Set("key_name", keyName)
	req.URL.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("create failed: %v", e))
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
	}
	return success(fmt.Sprintf("Instance created: %v", result["instance_id"]))
}

func HandleDestroyInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceID, _ :=getString(args, "instance_id")

	req, e := http.NewRequestWithContext(ctx, "DELETE", "https://api.aws.example.com/instances/"+instanceID, nil)
	if e != nil {
		return err(fmt.Sprintf("destroy request failed: %v", e))
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("destroy failed: %v", e))
	}
	defer resp.Body.Close()

	return ok("Instance destroyed")
}

func HandleRunCommand_cloud_command(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instanceID, _ :=getString(args, "instance_id")
	command, _ :=getString(args, "command")

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.aws.example.com/instances/"+instanceID+"/exec", nil)
	if e != nil {
		return err(fmt.Sprintf("exec request failed: %v", e))
	}
	q := req.URL.Query()
	q.Set("command", command)
	req.URL.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("exec failed: %v", e))
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
	}
	output, found := result["output"].(string)
	if !found {
		output = "command executed"
	}
	return success(output)
}// touch 1781132122
