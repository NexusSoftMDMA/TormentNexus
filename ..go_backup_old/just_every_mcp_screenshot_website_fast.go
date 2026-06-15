package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleScreenshotWebsiteFast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url is required")
}

	apiURL := fmt.Sprintf("http://localhost:3000/screenshot?url=%s", url.QueryEscape(targetURL))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("screenshot service error: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("screenshot service returned status " + resp.Status)
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(body)
	return ok(dataURI)
}// touch 1781132129
