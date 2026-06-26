package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type fngData struct {
	Value                string `json:"value"`
	ValueClassification string `json:"value_classification"`
}

type fngResponse struct {
	Data []fngData `json:"data"`
}

func HandleGetSentiment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	coin, _ :=getString(args, "coin")
	_ = coin

	resp, e := http.DefaultClient.Get("https://api.alternative.me/fng/?limit=1")
	if e != nil {
		return err("failed to fetch sentiment")
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var fng fngResponse
	if e := json.Unmarshal(body, &fng); e != nil {
		return err("failed to parse response")
}

	if len(fng.Data) == 0 {
		return err("no sentiment data available")
}

	msg := fmt.Sprintf("Crypto Fear & Greed Index: %s - %s", fng.Data[0].Value, fng.Data[0].ValueClassification)
	return success(msg)
}