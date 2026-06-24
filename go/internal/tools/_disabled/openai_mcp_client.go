package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	data := map[string]interface{}{"model": model, "prompt": prompt}
	body, _ := json.Marshal(data)
	resp, e := http.DefaultClient.Post("https://api.openai.com/v1/engines/"+model+"/completions", "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to call OpenAI API")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected response from OpenAI API")
}

	return success("request successful")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Additional handler can be implemented here
	return success("HandleY executed")
}