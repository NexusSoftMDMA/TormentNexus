package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetRecipe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	url := fmt.Sprintf("https://www.themealdb.com/api/json/v1/1/search.php?s=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch recipe: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	meals, found := data["meals"].([]interface{})
	if !found || len(meals) == 0 {
		return err("no recipe found")
}

	meal := meals[0].(map[string]interface{})
	recipeName, _ := meal["strMeal"].(string)
	instructions, _ := meal["strInstructions"].(string)
	result := fmt.Sprintf("Recipe: %s\nInstructions: %s", recipeName, instructions)
	return ok(result)
}