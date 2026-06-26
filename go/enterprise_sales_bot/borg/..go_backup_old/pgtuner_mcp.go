package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleTuneQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://pgtuner.example.com/tune?query=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to contact tuning service: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("tuning service returned status " + resp.Status)
}

	return success(string(body))
}