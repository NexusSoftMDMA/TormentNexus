package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleCheckIp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	if ip == "" {
		return err("missing ip")
}

	resp, e := http.DefaultClient.Get("https://api.abuseipdb.com/api/v2/check?ipAddress=" + ip)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleCheckDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("missing domain")
}

	resp, e := http.DefaultClient.Get("https://api.urlscan.io/v1/search/?q=domain:" + domain)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}