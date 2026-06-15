package tools

import (
	"context"
	"io"
	"net/http"
)

// HandleSearchObjects searches the Smithsonian Open Access API.
func HandleSearchObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "q")
	if query == "" {
		return err("query parameter 'q' is required")
}

	url := "https://api.si.edu/openaccess/api/v1.0/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}