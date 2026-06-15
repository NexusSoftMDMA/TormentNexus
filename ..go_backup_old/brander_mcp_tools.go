package tools

import "context"

func HandleBrand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	brand, _ :=getString(args, "brand")
	if brand == "" {
		brand = "BranderUX"
	}
	return success("[" + brand + "] " + text)
}