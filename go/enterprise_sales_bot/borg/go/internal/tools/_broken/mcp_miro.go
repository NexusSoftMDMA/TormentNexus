package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleCreateBoard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	desc, _ :=getString(args, "description")
	body, _ := json.Marshal(map[string]string{"name": name, "description": desc})
	req, e := http.NewRequest("POST", "https://api.miro.com/v2/boards", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("MIRO_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	boardID, found := result["id"].(string)
	if !found {
		return err("no board id in response")
}

	return success("Board created: " + boardID)
}

func HandleGetBoard_mcp_miro(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	boardID, _ :=getString(args, "boardId")
	if boardID == "" {
		return err("boardId is required")
}

	url := fmt.Sprintf("https://api.miro.com/v2/boards/%s", boardID)
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("MIRO_TOKEN"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(fmt.Sprintf("Board data: %s", string(data)))
}