package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleDeploy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	resourceType, _ :=getString(args, "type")
	if name == "" || resourceType == "" {
		return err("name and type are required")
}

	body, e := json.Marshal(map[string]string{"name": name, "type": resourceType})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.exaltbyte.com/deploy", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("deploy request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("deploy returned status " + resp.Status)
}

	return ok(resourceType + " '" + name + "' deployed successfully")
}