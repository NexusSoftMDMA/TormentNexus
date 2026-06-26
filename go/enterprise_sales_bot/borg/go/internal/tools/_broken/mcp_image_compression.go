package mcpimpl

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
)

func HandleCompressImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "imageData")
	quality, _ :=getInt(args, "quality")
	if quality == 0 {
		quality = 75
	}
	raw, e := base64.StdEncoding.DecodeString(data)
	if e != nil {
		return err("failed to decode base64: " + e.Error())
}

	img, format, e := image.Decode(bytes.NewReader(raw))
	if e != nil {
		return err("failed to decode image: " + e.Error())
}

	var buf bytes.Buffer
	if strings.EqualFold(format, "png") {
		e = png.Encode(&buf, img)
	} else {
		e = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})

	if e != nil {
		return err("failed to compress image: " + e.Error())
}

	result := base64.StdEncoding.EncodeToString(buf.Bytes())
	return ok(result)
}
}