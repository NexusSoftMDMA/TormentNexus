package tools

import (
	"context"
	"net"
	"time"
)

func HandleResolveHost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("missing host")
}

	addrs, e := net.LookupHost(host)
	if e != nil {
		return err("resolve failed: " + e.Error())
}

	return ok("resolved: " + host + " -> " + addrs[0])
}

func HandleCheckPort(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	if host == "" || port == 0 {
		return err("missing host or port")
}

	address := net.JoinHostPort(host, string(rune(port)))
	conn, e := net.DialTimeout("tcp", address, 5*time.Second)
	if e != nil {
		return ok("port " + string(rune(port)) + " closed")
}

	conn.Close()
	return ok("port " + string(rune(port)) + " open")
}