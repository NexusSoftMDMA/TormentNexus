package tools, so we can just call getString and getInt (the pre-defined ones). So we should remove our definitions.

Also need to adjust imports: we have many imports, but we may not need all. Keep stdlib only.

Also main function may not be needed but okay.

Also need to adjust return types: functions should return (ToolResponse, error). Use ok() and err("error") helpers per rule. The rule: return ok("text"), if e != nil { return err(e.Error()) }. So we should use ok() to wrap success. So instead of returning ToolResponse{...}, we should return ok(string). But ok likely returns (ToolResponse, error). So we can just return ok(string). Similarly for errors we can return err(errMsg). So modify functions accordingly.

Thus modify HandleSearch: after getting papersJSON, return ok(string(papersJSON)). If error, return err(e.Error()).

Similarly HandleDownload: return ok("Download successful").

HandleRead: return ok(text).

Also need to adjust imports: we used fmt, net/http, etc.package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Paper struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	maxResults, _ :=getInt(args, "max_results")
	sources, _ :=getString(args, "sources")

	searchTool := NewSearchTool(query, maxResults, sources)

	papers, e := searchTool.Search()
	if e != nil {
		return err(e.Error())
}

	papersJSON, e := json.Marshal(papers)
	if e != nil {
		return err(e.Error())
}

	return ok(string(papersJSON))
}

func HandleDownload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ :=getString(args, "paper_id")
	savePath, _ :=getString(args, "save_path")

	downloadTool := NewDownloadTool()

	if e := downloadTool.Download(paperID, savePath); e != nil {
		return err(e.Error())
}

	return ok("Download successful")
}

func HandleRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ :=getString(args, "paper_id")
	savePath, _ :=getString(args, "save_path")

	readTool := NewReadTool()

	text, e := readTool.Read(paperID, savePath)
	if e != nil {
		return err(e.Error())
}

	return ok(text)
}

func NewSearchTool(query string, maxResults int, sources string) *SearchTool {
	return &SearchTool{query: query, maxResults: maxResults, sources: sources}
}

func NewDownloadTool() *DownloadTool {
	return &DownloadTool{}
}

func NewReadTool() *ReadTool {
	return &ReadTool{}
}

type SearchTool struct {
	query      string
	maxResults int
	sources    string
}

func (st *SearchTool) Search() ([]Paper, error) {
	// Placeholder implementation
	return []Paper{
}
		{ID: "1", Title: "Sample Paper"},
	}, nil
}

type DownloadTool struct{}

func (dt *DownloadTool) Download(paperID string, savePath string) error {
	// Placeholder implementation
	return nil
}

type ReadTool struct{}

func (rt *ReadTool) Read(paperID string, savePath string) (string, error) {
	// Placeholder implementation
	return "Sample text content", nil
}

func main() {
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		args := make(map[string]interface{})
		if e := json.NewDecoder(r.Body).Decode(&args); e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		response, e := HandleSearch(ctx, args)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		args := make(map[string]interface{})
		if e := json.NewDecoder(r.Body).Decode(&args); e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		response, e := HandleDownload(ctx, args)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		args := make(map[string]interface{})
		if e := json.NewDecoder(r.Body).Decode(&args); e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		response, e := HandleRead(ctx, args)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}