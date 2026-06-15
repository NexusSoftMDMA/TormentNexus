//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetBusiness(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	businessID, _ :=getString(args, "business_id")
	if businessID == "" {
		return err("business_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.dversum.com/v1/businesses/" + businessID)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
	return ok("Business retrieved: " +
}


-reasoner (deepseek)*
}
}