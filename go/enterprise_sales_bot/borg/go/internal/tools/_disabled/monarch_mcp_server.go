package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetMonarchFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://uselessfacts.jsph.pl/random.json?language=en")
	if e != nil {
		return err("failed to fetch fact: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var data struct {
		Text string `json:"text"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse: " + e.Error())
}

	return ok("Monarch fact: " + data.Text)
}