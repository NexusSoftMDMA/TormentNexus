package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchCompounds(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://pubchem.ncbi.nlm.nih.gov/rest/pug/compound/name/%s/cids/JSON", url.PathEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		IdentifierList struct {
			CID []int `json:"CID"`
		} `json:"IdentifierList"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	if len(result.IdentifierList.CID) == 0 {
		return err("no compounds found")
}

	out, _ := json.Marshal(result.IdentifierList.CID)
	return ok(string(out))
}

func HandleGetCompoundProperties(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cid, _ :=getInt(args, "cid")
	if cid == 0 {
		return err("cid is required (integer)")
}

	u := fmt.Sprintf("https://pubchem.ncbi.nlm.nih.gov/rest/pug/compound/cid/%d/property/MolecularFormula,MolecularWeight,CanonicalSMILES/JSON", cid)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch properties: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		PropertyTable struct {
			Properties []map[string]interface{} `json:"Properties"`
		} `json:"PropertyTable"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	if len(result.PropertyTable.Properties) == 0 {
		return err("no properties found")
}

	out, _ := json.Marshal(result.PropertyTable.Properties[0])
	return ok(string(out))
}