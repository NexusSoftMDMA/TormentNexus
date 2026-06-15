package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSieveRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	input, _ :=getString(args, "input")
	if model == "" || input == "" {
		return err("model and input required")
}

	body, e := json.Marshal(map[string]interface{}{"model": model, "input": input})
	if e != nil {
		return err("marshal error: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.sievedata.com/v1/predict", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err("request error: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("decode error: " + e.Error())
}

	out, e := json.Marshal(result)
	if e != nil {
		return err("marshal output error: " + e.Error())
}

	return ok(string(out))
}