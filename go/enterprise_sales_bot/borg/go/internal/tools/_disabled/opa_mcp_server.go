package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleEvaluateOpaPolicy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing required arg: url")
}

	policyPath, _ :=getString(args, "policyPath")
	if policyPath == "" {
		return err("missing required arg: policyPath")
}

	inputStr, _ :=getString(args, "input")
	var inputData interface{}
	if inputStr != "" {
		if e := json.Unmarshal([]byte(inputStr), &inputData); e != nil {
			return err(fmt.Sprintf("invalid input JSON: %v", e))

	}
	body := map[string]interface{}{"input": inputData}
	bodyBytes, e := json.Marshal(body)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", url+"/v1/data/"+policyPath, bytes.NewBuffer(bodyBytes))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBytes, &result); e != nil {
		return err(fmt.Sprintf("invalid response JSON: %v", e))
}

	if decision, found := result["result"]; found {
		return ok(fmt.Sprintf("OPA decision: %v", decision))
}

	return ok(fmt.Sprintf("OPA response: %v", result))
}
}