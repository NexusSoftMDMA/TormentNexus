package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListCertificates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080/api/v1/certificates"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var certs []map[string]string
	if e := json.NewDecoder(resp.Body).Decode(&certs); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(certs)
	return success(string(data))
}

func HandleGetCertificate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080/api/v1/certificates"
	}
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	resp, e := http.DefaultClient.Get(url + "/" + name)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var cert map[string]string
	if e := json.NewDecoder(resp.Body).Decode(&cert); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(cert)
	return success(string(data))
}