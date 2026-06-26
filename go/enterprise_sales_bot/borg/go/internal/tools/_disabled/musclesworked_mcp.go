package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleGetMusclesWorked(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	exercise, _ :=getString(args, "exercise")
	if exercise == "" {
		return err("exercise parameter is required")
}

	apiURL := "https://musclesworked.com/api/exercise?name=" + url.QueryEscape(exercise)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}