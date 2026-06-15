package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleEnhanceVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	output, _ :=getString(args, "output")
	if input == "" || output == "" {
		return err("missing input or output parameter")
}

	resp, e := http.DefaultClient.Post("https://api.example.com/enhance", "application/json", strings.NewReader(`{"input":"`+input+`","output":"`+output+`"}`))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(result)
}

func HandleSegmentImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	image, _ :=getString(args, "image")
	if image == "" {
		return err("missing image parameter")
}

	resp, e := http.DefaultClient.Post("https://api.example.com/segment", "application/json", strings.NewReader(`{"image":"`+image+`"}`))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(result)
}