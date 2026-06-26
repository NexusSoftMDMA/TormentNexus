package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	digits, _ :=getInt(args, "digits")
	if digits < 1 || digits > 1000 {
		digits = 100
	}
	url := fmt.Sprintf("https://api.pi.delivery/v1/pi?start=0&numberOfDigits=%d", digits)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch pi: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	pi, found := result["content"].(string)
	if !found {
		return err("unexpected response format")
}

	return ok(fmt.Sprintf("Pi (first %d digits): %s", digits, pi))
}

func HandleGetPiInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Piapi MCP Server: Provides Pi digits via pi.delivery API")
}