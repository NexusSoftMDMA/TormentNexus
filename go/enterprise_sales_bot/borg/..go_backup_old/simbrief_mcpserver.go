package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFlightPlan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userid, _ :=getString(args, "userid")
	if userid == "" {
		return err("userid is required")
	}
	url := fmt.Sprintf("https://www.simbrief.com/api/xml.fet.php?userid=%s", userid)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	return ok(string(body))
}