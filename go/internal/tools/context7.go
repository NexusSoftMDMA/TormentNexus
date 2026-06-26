package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	key, _ :=getString(args, "key")
	if key == "" {
		return err("missing key")
}

	return ok("text")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	value, _ :=getInt(args, "value")
	if value == 0 {
		return err("missing value")
}

	return ok(fmt.Sprintf("value is %d", value))
}

func HandleZ(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	flag, _ :=getBool(args, "flag")
	if !flag {
		return err("missing flag")
}

	return ok("flag is true")
}

func HandleA(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	client := http.DefaultClient
	resp, fetchErr := client.Get("http://example.com")
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

func HandleB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	re, parseErr := regexp.Compile(`\d+`)
	if parseErr != nil {
		return err(parseErr.Error())
}

	text := "abc123def456"
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return err("no matches found")
}

	return ok(fmt.Sprintf("matches: %v", matches))
}