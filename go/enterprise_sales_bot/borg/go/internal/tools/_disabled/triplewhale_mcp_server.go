package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleQueryRevenue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	date, _ :=getString(args, "date")
	if date == "" {
		return err("missing required argument: date")
	}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.triplewhale.com/revenue?date=%s", date), nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}

	return success(fmt.Sprintf("Revenue for %s: %v", date, result["value"]))
}

func HandleQueryROAS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	campaign, _ :=getString(args, "campaign")
	if campaign == "" {
		return err("missing required argument: campaign")
	}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.triplewhale.com/roas?campaign=%s", campaign), nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}

	return success(fmt.Sprintf("ROAS for %s: %v", campaign, result["value"]))
}// touch 1781132142
