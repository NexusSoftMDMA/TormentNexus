package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListRecordings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/recordings")
	if e != nil {
		return err(fmt.Sprintf("failed to list recordings: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleStartRecording(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "default_recording"
	}
	resp, e := http.DefaultClient.PostForm("http://localhost:8080/record/start", map[string][]string{
		"name": {name},
	})
	if e != nil {
		return err(fmt.Sprintf("failed to start recording: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}