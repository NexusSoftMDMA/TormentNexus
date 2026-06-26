package mcpimpl

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

func HandleCaptureScreenshot_screenshotmcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	width, _ :=getInt(args, "width")
	if width == 0 {
		width = 800
	}
	height, _ :=getInt(args, "height")
	if height == 0 {
		height = 600
	}
	_ = url

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	e := png.Encode(&buf, img)
	if e != nil {
		return err("failed to encode screenshot: " + e.Error())
}

	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	return ok("Screenshot captured for URL: " + url + "\n" + dataURI)
}