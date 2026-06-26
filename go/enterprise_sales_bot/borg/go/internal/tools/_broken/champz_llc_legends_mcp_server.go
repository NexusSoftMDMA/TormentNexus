package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPlayerStats_champz_llc_legends_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	player, _ :=getString(args, "playerId")
	url := fmt.Sprintf("https://api.legendsofchampz.com/player/%s/stats", player)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleClaimReward(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	payload := map[string]string{
		"address":   getString(args, "address"),
		"signature": getString(args, "signature"),
		"rewardId":  getString(args, "rewardId"),
	}
	data, e := json.Marshal(payload)
	if e != nil {
		return err(e.Error())
}

	url := "https://api.legendsofchampz.com/rewards/claim"
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(data))
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}