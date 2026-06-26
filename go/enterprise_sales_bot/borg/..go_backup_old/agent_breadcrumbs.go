package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAddBreadcrumb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	user, _ :=getString(args, "user")
	body, _ := json.Marshal(map[string]string{"message": message, "user": user})
	resp, e := http.Post("https://breadcrumbs.example.com/add", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("failed to add breadcrumb: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("breadcrumb add returned non-200")
}

	return ok("breadcrumb added")
}

func HandleGetBreadcrumbs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	user, _ :=getString(args, "user")
	url := fmt.Sprintf("https://breadcrumbs.example.com/get?user=%s", user)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get breadcrumbs: %v", e))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(data))
}