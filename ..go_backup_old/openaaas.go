package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleGetUsers fetches users from a public API.
func HandleGetUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	page, _ :=getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users?_limit=%d&_page=%d", limit, page)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch users: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var users []map[string]interface{}
	if e = json.Unmarshal(body, &users); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Fetched %d users", len(users)))
}