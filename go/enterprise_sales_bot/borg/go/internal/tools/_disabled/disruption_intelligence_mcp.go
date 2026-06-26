package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleGetDisruptions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := "https://api.disruption-intelligence.com/v1/disruptions"
	u, e := url.Parse(baseURL)
	if e != nil {
		return err("failed to parse base URL")
}

	q := u.Query()
	if region := getString(args, "region"); region != "" {
		q.Set("region", region)

	if disruptionType := getString(args, "type"); disruptionType != "" {
		q.Set("type", disruptionType)

	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body")
}

	return ok(string(body))
}
}
}