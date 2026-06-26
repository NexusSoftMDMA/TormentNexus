package mcpimpl

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleGetVestingSchedule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contractId, _ :=getString(args, "contract_id")
	accountId, _ :=getString(args, "account_id")
	if contractId == "" || accountId == "" {
		return err("contract_id and account_id are required")
}

	url := "https://api.near.org/vesting/" + contractId + "/" + accountId
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get vesting schedule: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	json.Unmarshal(body, &result)
	return ok(result)
}

func HandleClaimVesting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Vesting claimed (simulated)")
}