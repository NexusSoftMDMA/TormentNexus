package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListDonations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost/wp-json/give-api/v1/donations"
	}
	status, _ :=getString(args, "status")
	perPage, _ :=getInt(args, "per_page")
	url := base + "?"
	if status != "" {
		url += "status=" + status + "&"
	}
	if perPage > 0 {
		url += "per_page=" + fmt.Sprint(perPage)

	url = strings.TrimRight(url, "?&")
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status " + resp.Status + ": " + string(body))
}

	var donations []map[string]interface{}
	if e := json.Unmarshal(body, &donations); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(fmt.Sprintf("Found %d donations", len(donations)))
}

}

func HandleGetDonation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost/wp-json/give-api/v1/donations"
	}
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := base + "/" + id
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status " + resp.Status + ": " + string(body))
}

	var donation map[string]interface{}
	if e := json.Unmarshal(body, &donation); e != nil {
		return err("parse failed: " + e.Error())
}

	jsonStr, _ := json.MarshalIndent(donation, "", "  ")
	return success(string(jsonStr))
}