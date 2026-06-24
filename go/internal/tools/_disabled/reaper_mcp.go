package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandlePlay(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	url := fmt.Sprintf("http://localhost:8080/command?action=PLAY&project=%s", project)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("playback started")
}

func HandleStop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	url := fmt.Sprintf("http://localhost:8080/command?action=STOP&project=%s", project)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("playback stopped")
}