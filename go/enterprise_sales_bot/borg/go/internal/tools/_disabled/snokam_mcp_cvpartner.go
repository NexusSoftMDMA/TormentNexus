package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListCVs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("CVPARTNER_API_KEY")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.cvpartner.com/v1/cvs", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(fmt.Sprintf("CVs:\n%s", string(body)))
}

func HandleGetCV(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cvID, _ :=getString(args, "cv_id")
	if cvID == "" {
		return err("cv_id is required")
}

	token := os.Getenv("CVPARTNER_API_KEY")
	url := fmt.Sprintf("https://api.cvpartner.com/v1/cvs/%s", cvID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return ok(fmt.Sprintf("CV:\n%s", string(body)))
}