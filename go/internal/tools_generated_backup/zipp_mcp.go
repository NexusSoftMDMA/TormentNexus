package tools

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListZipContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch %s: %v", url, e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP %d from %s", resp.StatusCode, url))
}

	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	reader, e := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if e != nil {
		return err(fmt.Sprintf("failed to open zip: %v", e))
}

	var names []string
	for _, f := range reader.File {
		names = append(names, f.Name)

	return ok(strings.Join(names, "\n"))
}
}