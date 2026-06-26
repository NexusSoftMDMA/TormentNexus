package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchWikidata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	search, _ :=getString(args, "search")
	if search == "" {
		return err("search parameter is required")
}

	lang, _ :=getString(args, "language")
	if lang == "" {
		lang = "en"
	}
	u := fmt.Sprintf("https://www.wikidata.org/w/api.php?action=wbsearchentities&search=%s&language=%s&format=json", url.QueryEscape(search), url.QueryEscape(lang))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleSparqlQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	u := fmt.Sprintf("https://query.wikidata.org/sparql?format=json&query=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to execute SPARQL: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}