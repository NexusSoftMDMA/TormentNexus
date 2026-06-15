//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	memory, _ :=getString(args, "memory")
	baseUrl, _ :=getString(args, "baseUrl")
	if baseUrl == "" {
		return err("baseUrl is required")
	switch action {
	case "store":
		body := map[string]string{"memory": memory}
		jsonBody, _ := json.Marshal(body)
		resp, e := http.DefaultClient.Post(baseUrl+"/memory", "application/json", bytes.NewBuffer(jsonBody))
		if e != nil {
			return err(fmt.Sprintf("failed to store memory: %v", e))
}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return err("failed to store memory: status " + resp.Status)
		return ok("memory stored")
	case "retrieve":
		resp, e := http.DefaultClient.Get(baseUrl + "/memory")
		if e != nil {
			return err(fmt.Sprintf("failed to retrieve memory: %v", e))
}

		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		return ok(string(data))
	default:
		return err("unknown action: " + action)

}
}
}
}