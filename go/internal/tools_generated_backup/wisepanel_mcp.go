package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetWisepanel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	panelID, _ :=getString(args, "panel_id")
	if panelID == "" {
		return err("panel_id is required")
}

	url := fmt.Sprintf("https://api.wisepanel.example.com/panels/%s", panelID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch panel: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(fmt.Sprintf("Panel data: %s", string(body)))
}