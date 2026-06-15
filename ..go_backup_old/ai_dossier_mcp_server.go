package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListDossiers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url required")
	}
	resp, e := http.DefaultClient.Get(strings.TrimRight(base, "/") + "/dossiers")
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	return ok(string(body))
}

func HandleExecuteDossier(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	id, _ :=getString(args, "dossier_id")
	input, _ :=getString(args, "input")
	if base == "" || id == "" {
		return err("base_url and dossier_id required")
	}
	payload, _ := json.Marshal(map[string]string{"input": input})
	url := strings.TrimRight(base, "/") + "/dossiers/" + id + "/execute"
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
	}
	return ok(string(body))
}