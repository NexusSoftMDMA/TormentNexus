package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

type fearGreedResponse struct {
	Data []struct {
		Value               string `json:"value"`
		ValueClassification string `json:"value_classification"`
		Timestamp           int64  `json:"timestamp"`
	} `json:"data"`
}

func HandleGetFearGreed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.alternative.me/fng/")
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var fg fearGreedResponse
	e = json.NewDecoder(resp.Body).Decode(&fg)
	if e != nil {
		return err(e.Error())
	}
	if len(fg.Data) == 0 {
		return err("no data")
	}
	value := fg.Data[0].Value
	classification := fg.Data[0].ValueClassification
	msg := "Fear & Greed Index: " + value + " (" + classification + ")"
	return success(msg)
}

func HandleCryptoSentiment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sym, _ :=getString(args, "symbol")
	msg := "Sentiment for " + sym + " is not implemented"
	return ok(msg)
}// touch 1781132123
