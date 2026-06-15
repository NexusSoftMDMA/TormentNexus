package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListRegions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://psgc.gitlab.io/api/regions/")
	if e != nil {
		return err("failed to fetch regions: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var regions []struct {
		Name string `json:"name"`
	}
	if e := json.Unmarshal(body, &regions); e != nil {
		return err("failed to parse regions: " + e.Error())
}

	var names []string
	for _, r := range regions {
		names = append(names, r.Name)

	return ok(fmt.Sprintf("Regions:\n%s", joinStrings(names, "\n")))
}

}

func HandleListProvinces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "region_code")
	if code == "" {
		return err("missing region_code argument")
}

	url := "https://psgc.gitlab.io/api/provinces/?regionCode=" + code
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch provinces: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var provinces []struct {
		Name string `json:"name"`
	}
	if e := json.Unmarshal(body, &provinces); e != nil {
		return err("failed to parse provinces: " + e.Error())
}

	var names []string
	for _, p := range provinces {
		names = append(names, p.Name)

	return ok(fmt.Sprintf("Provinces:\n%s", joinStrings(names, "\n")))
}

}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}