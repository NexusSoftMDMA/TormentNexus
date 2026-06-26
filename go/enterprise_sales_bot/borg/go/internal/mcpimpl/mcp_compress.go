package mcpimpl

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
)

func HandleCompress_mcp_compress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, e := gw.Write([]byte(input))
	if e != nil {
		return err(fmt.Sprintf("failed to compress: %v", e))
}

	e = gw.Close()
	if e != nil {
		return err(fmt.Sprintf("failed to close compressor: %v", e))
}

	compressed := base64.StdEncoding.EncodeToString(buf.Bytes())
	return ok(compressed)
}

func HandleDecompress(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "data")
	if data == "" {
		return err("data is required")
}

	compressed, e := base64.StdEncoding.DecodeString(data)
	if e != nil {
		return err(fmt.Sprintf("invalid base64: %v", e))
}

	gr, e := gzip.NewReader(bytes.NewReader(compressed))
	if e != nil {
		return err(fmt.Sprintf("failed to create reader: %v", e))
}

	var buf bytes.Buffer
	_, e = buf.ReadFrom(gr)
	if e != nil {
		return err(fmt.Sprintf("failed to decompress: %v", e))
}

	gr.Close()
	return ok(buf.String())
}