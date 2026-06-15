package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSearchCards(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.scryfall.com/cards/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch cards: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Data []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	if len(result.Data) == 0 {
		return ok("No cards found")
}

	var names []string
	for _, card := range result.Data {
		names = append(names, card.Name)

	return ok("Found cards: " + strings.Join(names, ", "))
}
}