package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetSettlement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "settlement_id")
	if id == "" {
		return err("settlement_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.agentvault.io/settlements/" + id)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(result)
}

func HandleSubmitSettlement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getInt(args, "amount")
	party, _ :=getString(args, "party")
	if party == "" || amount <= 0 {
		return err("party and positive amount are required")
}

	body, _ := json.Marshal(map[string]interface{}{"amount": amount, "party": party})
	resp, e := http.DefaultClient.Post("https://api.agentvault.io/settlements", "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("settlement submitted")
}