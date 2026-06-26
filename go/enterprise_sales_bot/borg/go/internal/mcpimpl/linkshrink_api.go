package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleShorten_linkshrink_api(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	link, _ :=getString(args, "url")
	if link == "" {
		return err("Missing required parameter: url")
}

	apiURL := "https://linkshrink.net/api/v1/shorten"
	resp, e := http.DefaultClient.PostForm(apiURL, url.Values{"url": {link}})
	if e != nil {
		return err(fmt.Sprintf("Request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Read response failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("JSON parse error: %v", e))
}

	shortURL, found := result["short_url"].(string)
	if !found {
		return err("Unexpected response format")
}

	return ok(fmt.Sprintf("Short URL: %s", shortURL))
}