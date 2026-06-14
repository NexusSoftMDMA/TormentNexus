package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStrapiList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contentType, _ :=getString(args, "content_type")
	if contentType == "" {
		return err("content_type argument is required")
	}
	url := fmt.Sprintf("http://localhost:1337/api/%s", contentType)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	return success(string(body))
}

func HandleStrapiGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	contentType, _ :=getString(args, "content_type")
	id, _ :=getString(args, "id")
	if contentType == "" || id == "" {
		return err("content_type and id arguments are required")
	}
	url := fmt.Sprintf("http://localhost:1337/api/%s/%s", contentType, id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err(e.Error())
	}
	return success(string(body))
}// touch 1781132141
