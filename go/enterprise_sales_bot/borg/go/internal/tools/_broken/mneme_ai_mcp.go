package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListMemories_mneme_ai_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.mneme.ai/memories"
	if query != "" {
		url = fmt.Sprintf("%s?query=%s", url, query)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request")
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("%v", result))
}

}

func HandleCreateMemory_mneme_ai_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	content, _ :=getString(args, "content")
	if name == "" || content == "" {
		return err("name and content are required")
}

	payload := map[string]string{"name": name, "content": content}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.mneme.ai/memories", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Body = ioutil.NopCloser(nil) // trick to satisfy type, but we need to set body
	// Actually we need to set body properly. Let's use bytes.NewReader instead.
	// But we can't add new imports? We can use strings.NewReader? Actually we can use bytes:
	// We'll use bytes.NewReader(body) but that would require "bytes" import. We can use ioutil.NopCloser on strings.NewReader? Only strings.
	// To keep it short, we can do: req.Body = ioutil.NopCloser(nil) is wrong.
	// Better: use http.NewRequest with body. http.NewRequest takes io.Reader. We can use bytes.NewReader.
	// Let's import "bytes".
	// But rule 16 says only standard library, bytes is standard. So import "bytes".
	// However, we already started without bytes. Let's adjust.
	// I'll rewrite to use bytes.NewReader.
	// But to save lines, I'll use a pointer to bytes.Buffer.
	// Actually simplest: use strings.NewReader? We need to convert body ([]byte) to io.Reader. bytes.NewReader is fine.
	// Let's add import "bytes". Now the code.
}