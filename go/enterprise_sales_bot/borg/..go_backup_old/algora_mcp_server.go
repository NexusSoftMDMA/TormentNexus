package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAlgoraGetBounty(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	url := fmt.Sprintf("https://api.algora.io/bounties/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("bad status: " + resp.Status)
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok("got bounty: " + fmt.Sprint(data))
}

func HandleAlgoraListBounties(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.algora.io/bounties"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("bad status: " + resp.Status)
}

	var data []interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("found %d bounties", len(data)))
}