//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func HandleListObjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bucketName, _ :=getString(args, "bucketName")
	if bucketName == "" {
		return err("bucket name is required")
}

	baseURL := os.Getenv("BUCKET_BASE_URL")
	if baseURL == "" {
		return err("BUCKET_BASE_URL environment variable not set")
}

	url := fmt.Sprintf("%s/%s/objects", baseURL, bucketName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	return ok(string(body))


-reasoner (deepseek)*
}
}