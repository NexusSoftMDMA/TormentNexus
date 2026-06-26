package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListBrands(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	brands := []string{"Nike", "Adidas", "Puma"}
	data, _ := json.Marshal(brands)
	return ok(string(data))
}

func HandleGetBrand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	resp, e := http.DefaultClient.Get("https://example.com/brands/" + name)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch brand: %v", e))
}

	defer resp.Body.Close()
	var brand map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&brand)
	data, _ := json.Marshal(brand)
	return ok(string(data))
}