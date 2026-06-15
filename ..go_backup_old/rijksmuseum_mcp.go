package tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleSearchArtworks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey is required")
}

	url := "https://www.rijksmuseum.nl/api/en/collection?key=" + apiKey + "&q=" + query + "&format=json"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("parse failed: " + e.Error())
}

	_ = result
	return ok("Search results: " + string(body))
}