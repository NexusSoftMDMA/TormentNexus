package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListJoinedRooms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "access_token")
	url := strings.TrimRight(baseURL, "/") + "/_matrix/client/r0/joined_rooms?access_token=" + token
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get rooms: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result struct {
		JoinedRooms []string `json:"joined_rooms"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return success(fmt.Sprintf("Found %d rooms: %v", len(result.JoinedRooms), result.JoinedRooms))
}