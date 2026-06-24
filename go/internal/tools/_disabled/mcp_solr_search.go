package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSolrSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "q")
	if q == "" {
		return err("missing query parameter q")
}

	base, _ :=getString(args, "url")
	if base == "" {
		return err("missing url parameter")
}

	u, e := url.Parse(base + "/select?q=" + url.QueryEscape(q) + "&wt=json")
	if e != nil {
		return err("invalid url: " + e.Error())
}

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("solr returned %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("json decode: " + e.Error())
}

	response, found := result["response"].(map[string]interface{})
	if !found {
		return err("missing response field")
}

	docs, found := response["docs"].([]interface{})
	if !found {
		return err("missing docs field")
}

	return ok(fmt.Sprintf("found %d results", len(docs)))
}