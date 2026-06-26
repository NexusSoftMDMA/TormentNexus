package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientID, _ :=getString(args, "clientID")
	version, _ :=getString(args, "version")

	if clientID == "" {
		return err("clientID is required")
}

	if version == "" {
		return err("version is required")
}

	resp, e := http.DefaultClient.Get("http://example.com/launch?clientID=" + clientID + "&version=" + version)
	if e != nil {
		return err("failed to launch client")
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("failed to launch client: " + resp.Status)
}

	return success("client launched successfully")
}