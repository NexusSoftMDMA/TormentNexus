package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleMakeCall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	from, _ :=getString(args, "from")
	payload, e := json.Marshal(map[string]string{"to": to, "from": from})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.telephony.example/call", "application/json", bytes.NewBuffer(payload))
	if e != nil {
		return err("failed to make call: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("call failed with status: " + resp.Status)
}

	return ok(fmt.Sprintf("Call from %s to %s initiated", from, to))
}

func HandleSendSms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	from, _ :=getString(args, "from")
	message, _ :=getString(args, "message")
	payload, e := json.Marshal(map[string]string{"to": to, "from": from, "message": message})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("https://api.telephony.example/sms", "application/json", bytes.NewBuffer(payload))
	if e != nil {
		return err("failed to send SMS: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("SMS failed with status: " + resp.Status)
}

	return ok(fmt.Sprintf("SMS sent to %s", to))
}