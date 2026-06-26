package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListReleases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	url := fmt.Sprintf("https://helm.example.com/api/releases?namespace=%s", namespace)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch releases: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("Releases: %v", result))
}

func HandleInstallChart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	chart, _ :=getString(args, "chart")
	namespace, _ :=getString(args, "namespace")
	payload := map[string]string{"name": name, "chart": chart, "namespace": namespace}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	resp, e := http.DefaultClient.Post("https://helm.example.com/api/releases", "application/json", io.NopCloser(bytes.NewReader(body)))
	if e != nil {
		return err(fmt.Sprintf("failed to install chart: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("install failed with status %d", resp.StatusCode))
}

	return success("Chart installed successfully")
}