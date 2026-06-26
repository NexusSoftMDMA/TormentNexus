package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleFabric(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chaincode, _ :=getString(args, "chaincode")
	function, _ :=getString(args, "function")
	arg, _ :=getString(args, "arg")
	gateway := os.Getenv("FABRIC_GATEWAY_URL")
	if gateway == "" {
		gateway = "http://localhost:7050/query"
	}
	payload := map[string]interface{}{
		"chaincode": chaincode,
		"function":  function,
		"args":      []string{arg},
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed")
}

	resp, e := http.DefaultClient.Post(gateway, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode failed")
}

	return ok(fmt.Sprintf("Result: %v", result))
}