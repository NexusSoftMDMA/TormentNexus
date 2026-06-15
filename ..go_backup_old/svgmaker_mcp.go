package tools

import (
	"context"
	"fmt"
)

func HandleCreateSvg(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	width, _ :=getInt(args, "width")
	height, _ :=getInt(args, "height")
	if width == 0 {
		width = 100
	}
	if height == 0 {
		height = 100
	}
	svg := fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg"><rect width="100%%" height="100%%" fill="white"/><text x="10" y="30" font-size="12">%s</text></svg>`, width, height, prompt)
	return ok(svg)
}