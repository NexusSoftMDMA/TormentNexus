package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleListDataSilos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("TRANSCEND_BASE_URL")
	if base == "" {
		base = "https://api.transcend.io"
	}
	url := fmt.Sprintf("%s/v1/data-silos", base)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to list data silos: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode: %v", e))
}

	return ok(fmt.Sprintf("data silos: %+v", result))
}

func HandleGetDataSilo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required argument: id")
}

	base := os.Getenv("TRANSCEND_BASE_URL")
	if base == "" {
		base = "https://api.transcend.io"
	}
	url := fmt.Sprintf("%s/v1/data-silos/%s", base, id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get data silo: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode: %v", e))
}

	return ok(fmt.Sprintf("data silo: %+v", result))
}// touch 1781132142
