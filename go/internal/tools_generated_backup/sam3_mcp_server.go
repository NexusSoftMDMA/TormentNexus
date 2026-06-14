package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleSam3Segment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "image_url")
	if imageURL == "" {
		return err("image_url is required")
}

	apiURL, _ :=getString(args, "api_url")
	if apiURL == "" {
		apiURL = os.Getenv("SAM3_API_URL")
		if apiURL == "" {
			return err("api_url is required")

	}
	body, e := json.Marshal(map[string]string{"image_url": imageURL})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		ResultURL string `json:"result_url"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok("Pre-signed URL: " + result.ResultURL)
}
}