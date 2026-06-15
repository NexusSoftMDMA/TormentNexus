package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + fmt.Sprintf("%d", resp.StatusCode))
}

	return ok("conformance test passed")
}