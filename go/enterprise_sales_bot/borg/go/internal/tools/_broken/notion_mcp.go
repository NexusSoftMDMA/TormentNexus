package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDatabases_notion_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.notion.com/v1/databases", nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleQueryDatabase_notion_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	dbID, _ :=getString(args, "database_id")
	if token == "" || dbID == "" {
		return err("missing token or database_id")
}

	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", dbID)
	var bodyReader io.Reader
	filter := args["filter"]
	if filter != nil {
		b, e := json.Marshal(filter)
		if e != nil {
			return err(fmt.Sprintf("marshal filter: %v", e))
}

		bodyReader = io.NopCloser(readerFromBytes(b))

	req, e := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

}

// readerFromBytes is a helper to wrap byte slice into io.Reader
func readerFromBytes(b []byte) io.Reader {
	return &byteReader{b}
}

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (n int, e error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n = copy(p, r.b[r.i:])
	r.i += n
	return
}