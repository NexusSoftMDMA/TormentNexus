package tools

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func HandleHocrExtractText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	re := regexp.MustCompile(`<span[^>]*class="ocrx_word"[^>]*>([^<]+)</span>`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	var texts []string
	for _, m := range matches {
		texts = append(texts, m[1])

	result := strings.Join(texts, " ")
	return ok(result)
}
}