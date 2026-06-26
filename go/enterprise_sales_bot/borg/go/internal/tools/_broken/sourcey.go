package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		text = "text"
	}
	return ok(text)
}

func main() {
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		args := make(map[string]interface{})
		if e := json.NewDecoder(r.Body).Decode(&args); e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		response, e := HandleXxx(ctx, args)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if e := json.NewEncoder(w).Encode(response); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)

	})
	if e := http.ListenAndServe(":8080", nil); e != nil {
		_ = e
	}
}
}