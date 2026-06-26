package tools

import (
	"context"
	"encoding/json"
)

func SemioticAnalyzer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	length := len(text)
	vowels := 0
	for _, ch := range text {
		switch ch {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			vowels++
		}
	}
	result := map[string]int{"length": length, "vowels": vowels}
	jsonBytes, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return success(string(jsonBytes))
}