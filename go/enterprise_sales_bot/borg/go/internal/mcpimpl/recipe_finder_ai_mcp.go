package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleFindRecipe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ingredient, _ :=getString(args, "ingredient")
	if ingredient == "" {
		return err("ingredient is required")
}

	url := fmt.Sprintf("https://api.example.com/recipes?ingredient=%s", ingredient)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch recipes: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleSubstituteIngredient(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ingredient, _ :=getString(args, "ingredient")
	if ingredient == "" {
		return err("ingredient is required")
}

	url := fmt.Sprintf("https://api.example.com/substitutes?ingredient=%s", ingredient)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch substitutes: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}