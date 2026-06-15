package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	if title == "" || content == "" {
		return err("title and content are required")
}

	reqBody := map[string]string{"title": title, "content": content}
	body, e := json.Marshal(reqBody)
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.memvid.example/videos", "application/json", fakeReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status")
}

	return ok("video created")
}

func HandleListVideos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.memvid.example/videos?limit=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	raw, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var list []map[string]interface{}
	if e := json.Unmarshal(raw, &list); e != nil {
		return err("failed to parse response")
}

	return success(fmt.Sprintf("found %d videos", len(list)))
}

func fakeReader(data []byte) io.Reader {
	return &fakeReaderImpl{data: data}
}

type fakeReaderImpl struct {
	data []byte
	pos  int
}

func (r *fakeReaderImpl) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}