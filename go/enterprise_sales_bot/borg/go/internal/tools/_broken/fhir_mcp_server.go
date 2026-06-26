package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleSearchPatients_fhir_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := "https://hapi.fhir.org/baseR4"
	name, _ :=getString(args, "name")
	u, e := url.Parse(base + "/Patient")
	if e != nil {
		return err("invalid base URL")
}

	q := u.Query()
	if name != "" {
		q.Set("name", name)

	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}

}

func HandleGetPatient_fhir_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := "https://hapi.fhir.org/baseR4"
	id, _ :=getString(args, "id")
	if id == "" {
		return err("patient ID is required")
}

	u := base + "/Patient/" + url.PathEscape(id)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("FHIR server returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}