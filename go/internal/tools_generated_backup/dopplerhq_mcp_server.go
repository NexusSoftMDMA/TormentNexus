package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListSecrets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	config, _ :=getString(args, "config")
	if project == "" || config == "" {
		return err("project and config are required")
	}

	url := fmt.Sprintf("https://api.doppler.com/v3/workplace/configs/%s/secrets?project=%s", config, project)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}

	return ok("secrets retrieved")
}

func HandleGetSecret(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	config, _ :=getString(args, "config")
	name, _ :=getString(args, "name")
	if project == "" || config == "" || name == "" {
		return err("project, config, and name are required")
	}

	url := fmt.Sprintf("https://api.doppler.com/v3/workplace/configs/%s/secrets/%s?project=%s", config, name, project)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}

	return ok("secret retrieved")
}// touch 1781132125
