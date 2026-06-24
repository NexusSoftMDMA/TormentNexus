package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchNcbi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	u := fmt.Sprintf("https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=%s&retmax=5&retmode=json", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
	}
	esearch, found := result["esearchresult"].(map[string]interface{})
	if !found {
		return err("unexpected response format")
	}
	idlist, found := esearch["idlist"].([]interface{})
	if !found || len(idlist) == 0 {
		return ok("no results found")
}

	ids := make([]string, len(idlist))
	for i, v := range idlist {
		ids[i] = fmt.Sprintf("%v", v)

	return ok(fmt.Sprintf("found %d IDs: %s", len(ids), ids))
}
}