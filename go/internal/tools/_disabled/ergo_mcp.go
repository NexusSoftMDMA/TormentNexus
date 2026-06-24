package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	blockID, _ :=getString(args, "blockId")
	if blockID == "" {
		return err("blockId is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.ergoplatform.com/api/v1/blocks/%s", blockID))
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err("json error: " + e.Error())
}

	result, _ := json.Marshal(data)
	return success(string(result))
}

func HandleGetTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txID, _ :=getString(args, "txId")
	if txID == "" {
		return err("txId is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.ergoplatform.com/api/v1/transactions/%s", txID))
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data map[string]interface{}
	e = json.Unmarshal(body, &data)
	if e != nil {
		return err("json error: " + e.Error())
}

	result, _ := json.Marshal(data)
	return success(string(result))
}