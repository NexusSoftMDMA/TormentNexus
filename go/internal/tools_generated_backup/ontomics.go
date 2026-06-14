package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchOntologyTerm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	term, _ :=getString(args, "term")
	ontology, _ :=getString(args, "ontology")
	if term == "" {
		return err("term parameter is required")
}

	url := fmt.Sprintf("https://www.ebi.ac.uk/ols/api/search?q=%s&ontology=%s", term, ontology)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleGetOntologyRelationships(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	ontology, _ :=getString(args, "ontology")
	if id == "" || ontology == "" {
		return err("id and ontology parameters are required")
}

	url := fmt.Sprintf("https://www.ebi.ac.uk/ols/api/ontologies/%s/terms/%s/relations", ontology, id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Relationships: %v", result))
}