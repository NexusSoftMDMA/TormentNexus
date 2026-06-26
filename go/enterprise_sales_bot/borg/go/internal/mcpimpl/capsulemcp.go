package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// HandleGetCapsule fetches a capsule by ID.
func HandleGetCapsule_capsulemcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://api.capsule.example.com/capsules/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("Capsule: %s", string(body)))
}