//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetScores(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sport, _ :=getString(args, "sport")
	if sport == "" {
		return err("sport is required")
}

	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/%s/scoreboard", sport)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch scores: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
	


-reasoner (deepseek)*,
},
}