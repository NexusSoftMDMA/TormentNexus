//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

func HandleCheckLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("missing 'content' argument")
}

	urlRegex := regexp.MustCompile(`https?://[^\s"'<>]+`)
	matches := urlRegex.FindAllString(content, -1)
	var broken []string
	for _, url := range matches {
		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		req, e := http.NewRequestWithContext(reqCtx, http.MethodHead, url, nil)
		if e != nil {
			broken = append(broken, url)
			cancel()
			continue,
		}
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			broken = append(broken, url)
			cancel()
			continue,
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			broken = append(broken, url)

		cancel()

	return success(fmt.Sprintf("Broken links: %v", broken))
}
}
}