package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetConvictions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.conviction.fm/convictions")
	if e != nil {
		return err("failed to fetch convictions")
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	return ok("convictions fetched")
}

func HandleSubmitConviction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	proposal, _ :=getString(args, "proposal")
	amount, _ :=getInt(args, "amount")
	if proposal == "" || amount <= 0 {
		return err("proposal and amount required")
}

	body, _ := json.Marshal(map[string]interface{}{"proposal": proposal, "amount": amount})
	resp, e := http.DefaultClient.Post("https://api.conviction.fm/convictions", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to submit")
}

	resp.Body.Close()
	return success("conviction submitted")
}