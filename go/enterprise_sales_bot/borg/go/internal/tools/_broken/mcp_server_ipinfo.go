package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetIpInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	url := "https://ipinfo.io/json"
	if ip != "" {
		url = fmt.Sprintf("https://ipinfo.io/%s/json", ip)

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch IP info")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response")
}

	out, _ := json.MarshalIndent(data, "", "  ")
	return success(string(out))
}
}