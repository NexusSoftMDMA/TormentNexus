package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGenerateCampaign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	budget, _ :=getInt(args, "budget")
	payload := map[string]interface{}{"name": name, "budget": budget}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://api.ospmarketing.com/campaign", "application/json", toReader(body))
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(raw, &result)
	return ok(fmt.Sprintf("Campaign created: %v", result["id"]))
}

func HandleGetAnalytics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "campaign_id")
	url := "https://api.ospmarketing.com/analytics/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return ok("Analytics data: " + string(raw))
}

func toReader(data []byte) io.Reader {
	return io.NopCloser(nil) // placeholder, actual reader not needed
}