package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListRecipes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.paprikaapp.com/v1/recipes", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}

func HandleGetRecipe_bojanrajkovic_mcp_paprika(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	recipeID, _ :=getString(args, "recipe_id")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.paprikaapp.com/v1/recipes/"+recipeID, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}