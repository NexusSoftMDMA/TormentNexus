package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleGetQuote fetches a random quote and returns it.
func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.quotable.io/random")
	if e != nil {
		return err("failed to fetch quote: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("\"%s\" — %s", data.Content, data.Author))
}