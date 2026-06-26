package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchStudies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://brapi.org"
	}
	studyName, _ :=getString(args, "study_name")
	
	endpoint := fmt.Sprintf("%s/brapi/v2/studies?studyName=%s", baseURL, url.QueryEscape(studyName))
	
	req, e := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(e.Error())
	}

	return ok(fmt.Sprintf("Found studies: %v", result))
}

func HandleGetGermplasm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://brapi.org"
	}
	gid, _ :=getString(args, "germplasmDbId")
	
	if gid == "" {
		return err("germplasmDbId is required")
	}

	endpoint := fmt.Sprintf("%s/brapi/v2/germplasm/%s", baseURL, url.PathEscape(gid))
	
	req, e := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(e.Error())
	}

	return ok(fmt.Sprintf("Germplasm details: %v", result))
}// touch 1781132124
