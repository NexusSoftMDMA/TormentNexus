package tools

import (
	"context"
	"math"
	"strconv"
	"strings"
)

func HandleContrastRatio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fg, _ :=getString(args, "foreground")
	bg, _ :=getString(args, "background")
	if fg == "" || bg == "" {
		return err("foreground and background colors are required")
}

	l1, e := luminance(fg)
	if e != nil {
		return err("invalid foreground color: " + e.Error())
}

	l2, e := luminance(bg)
	if e != nil {
		return err("invalid background color: " + e.Error())
}

	if l1 < l2 {
		l1, l2 = l2, l1
	}
	ratio := (l1 + 0.05) / (l2 + 0.05)
	return ok("Contrast ratio: " + strconv.FormatFloat(ratio, 'f', 2, 64) + ":1")
}

func luminance(hex string) (float64, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, strconv.ErrSyntax
	}
	r, e := strconv.ParseUint(hex[0:2], 16, 8)
	if e != nil {
		return 0, e
	}
	g, e := strconv.ParseUint(hex[2:4], 16, 8)
	if e != nil {
		return 0, e
	}
	b, e := strconv.ParseUint(hex[4:6], 16, 8)
	if e != nil {
		return 0, e
	}
	linear := func(c uint8) float64 {
		v := float64(c) / 255.0
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
}

	return 0.2126*linear(uint8(r)) + 0.7152*linear(uint8(g)) + 0.0722*linear(uint8(b)), nil
}