package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetOdds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sport, _ :=getString(args, "sport")
	event, _ :=getString(args, "event")
	url := fmt.Sprintf("https://api.sportsbetting.com/odds?sport=%s&event=%s", sport, event)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}