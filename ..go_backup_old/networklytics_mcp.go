package tools

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

func HandleGetNetworkInterfaces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	interfaces, e := net.Interfaces()
	if e != nil {
		return err("Failed to get network interfaces: " + e.Error())
}

	var result string
	for _, iface := range interfaces {
		result += fmt.Sprintf("Name: %s, MTU: %d, Flags: %s\n", iface.Name, iface.MTU, iface.Flags.String())
		addrs, e := iface.Addrs()
		if e != nil {
			continue
		}
		for _, addr := range addrs {
			result += fmt.Sprintf("  %s\n", addr.String())

	}
	return success(result)
}

}

func HandleCheckConnectivity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to connect: " + e.Error())
}

	defer resp.Body.Close()
	return success(fmt.Sprintf("Status: %s", resp.Status))
}