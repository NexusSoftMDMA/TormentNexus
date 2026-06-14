package tools

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func HandleCheckColorContrast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fgHex := strings.TrimPrefix(getString(args, "foreground"), "#")
	bgHex := strings.TrimPrefix(getString(args, "background"), "#")
	if len(fgHex) != 6 || len(bgHex) != 6 {
		return err("foreground and background must be 6-digit hex colors")
}

	rgb := func(hex string) (float64, float64, float64) {
		r, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b, _ := strconv.ParseUint(hex[4:6], 16, 8)
		return float64(r), float64(g), float64(b)
}

	luminance := func(r, g, b float64) float64 {
		sr := r / 255.0
		sg := g / 255.0
		sb := b / 255.0
		sr = math.Pow(sr, 2.2)
		sg = math.Pow(sg, 2.2)
		sb = math.Pow(sb, 2.2)
		return 0.2126*sr + 0.7152*sg + 0.0722*sb
	}
	r1, g1, b1 := rgb(fgHex)
	r2, g2, b2 := rgb(bgHex)
	l1 := luminance(r1, g1, b1)
	l2 := luminance(r2, g2, b2)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	contrast := (l1 + 0.05) / (l2 + 0.05)
	aa := contrast >= 4.5
	aaa := contrast >= 7.0
	msg := fmt.Sprintf("Contrast ratio: %.2f:1 - AA: %v, AAA: %v", contrast, aa, aaa)
	return ok(msg)
}

func HandleSuggestAltText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	desc, _ :=getString(args, "description")
	if desc == "" {
		return success("A brief description of the image content.")
}

	return success(fmt.Sprintf("Suggestion: %s (consider adding context)", desc))
}