package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var upstreams = []string{"http://localhost:8081", "http://localhost:8082"}

func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var all []map[string]interface{}
	for _, u := range upstreams {
		resp, e := http.DefaultClient.Get(u + "/resources")
		if e != nil {
			continue
		}
		body, e := io.ReadAll(resp.Body)
		resp.Body.Close()
		if e != nil {
			continue
		}
		var list []map[string]interface{}
		if e := json.Unmarshal(body, &list); e != nil {
			continue
		}
		all = append(all, list...)

	return ok(fmt.Sprintf("Found %d resources", len(all)))
}
}