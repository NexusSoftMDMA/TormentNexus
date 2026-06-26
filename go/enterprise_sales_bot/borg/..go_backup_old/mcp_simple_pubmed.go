package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type pubmedSearchResult struct {
	EsarchResult struct {
		IdList []string `json:"IdList"`
	} `json:"esearchresult"`
}

func HandlePubmedSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "maxResults")
	if maxResults <= 0 {
		maxResults = 10
	}
	resp, e := http.DefaultClient.Get("https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&term=" + url.QueryEscape(query) + "&retmax=" + fmt.Sprint(maxResults) + "&retmode=json")
	if e != nil {
		return err("failed to search PubMed: " + e.Error())
}

	defer resp.Body.Close()
	var result pubmedSearchResult
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(result.EsarchResult.IdList) == 0 {
		return ok("No results found")
}

	return ok(fmt.Sprintf("Found %d PMIDs: %s", len(result.EsarchResult.IdList), result.EsarchResult.IdList))
}