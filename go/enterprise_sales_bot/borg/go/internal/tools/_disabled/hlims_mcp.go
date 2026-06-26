package tools

import (
	"context"
	"io"
	"net/http"
	"os"
)

func HandleGetSampleInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("HLIMS_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	url := base + "/sample/" + getString(args, "sample_id")
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("req: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status " + resp.Status)
}

	return ok(string(body))
}