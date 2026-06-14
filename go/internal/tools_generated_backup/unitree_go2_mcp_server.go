package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	robotIP, _ :=getString(args, "robot_ip")
	if robotIP == "" {
		return err("robot_ip is required")
}

	url := fmt.Sprintf("http://%s/status", robotIP)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleSendCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	robotIP, _ :=getString(args, "robot_ip")
	cmd, _ :=getString(args, "command")
	if robotIP == "" || cmd == "" {
		return err("robot_ip and command are required")
}

	payload := fmt.Sprintf(`{"command":"%s"}`, cmd)
	url := fmt.Sprintf("http://%s/command", robotIP)
	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(payload))
	if e != nil {
		return err("failed to send command: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("command failed: %s", string(body)))
}

	return success("command sent successfully")
}