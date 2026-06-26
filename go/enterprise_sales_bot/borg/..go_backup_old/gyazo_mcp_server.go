package tools

import (
	"context"
	"net/url"
	"strings"
)

func HandleGyazoGetDirect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("url is required")
}

	parsed, e := url.Parse(imageURL)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	path := strings.TrimPrefix(parsed.Path, "/")
	id := path
	if strings.Contains(id, "/") {
		parts := strings.Split(path, "/")
		id = parts[len(parts)-1]
	}
	if id == "" {
		return err("could not extract image ID from URL")
}

	direct := "https://i.gyazo.com/" + id + ".png"
	return ok("Direct image URL: " + direct)
}