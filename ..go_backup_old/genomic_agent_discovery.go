package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchGene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://rest.ensembl.org/lookup/symbol/homo_sapiens/%s?content-type=application/json", symbol))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	if id, found := data["id"]; found {
		return success("Gene ID: " + id.(string))
}

	return ok("No gene found")
}

func HandleGetVariant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rsid, _ :=getString(args, "rsid")
	if rsid == "" {
		return err("rsid is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://rest.ensembl.org/variation/homo_sapiens/%s?content-type=application/json", rsid))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	if allele, found := data["ancestral_allele"]; found {
		return success("Ancestral allele: " + allele.(string))
}

	return ok("No variant info")
}