package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	indexUrl, _ :=getString(args, "indexUrl")
	topK, _ :=getInt(args, "topK")
	if topK <= 0 {
		topK = 10
	}
	namespace, _ :=getString(args, "namespace")
	vectorStr, _ :=getString(args, "vector")
	parts := strings.Split(vectorStr, ",")
	floats := make([]float64, len(parts))
	for i, p := range parts {
		f, e := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if e != nil {
			return err("invalid vector element: " + e.Error())
}

		floats[i] = f
	}
	body, _ := json.Marshal(map[string]interface{}{
		"vector":    floats,
		"topK":      topK,
		"namespace": namespace,
	})
	req, e := http.NewRequestWithContext(ctx, "POST", indexUrl+"/query", strings.NewReader(string(body)))
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("bad status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode: " + e.Error())
}

	out, _ := json.Marshal(result)
	return success(string(out))
}