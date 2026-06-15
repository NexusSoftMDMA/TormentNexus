//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queue, _ :=getString(args, "queue")
	message, _ :=getString(args, "message")
	brokerURL, _ :=getString(args, "broker_url")
	if brokerURL == "" {
		brokerURL = "http://localhost:8161/api/message"
	}
	url := brokerURL + "?queue=" + queue
	resp, e := http.DefaultClient.Post(url, "text/plain", strings.NewReader(message))
	if e != nil {
		return err("failed to send message: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("message sent: " + string(body))
}

func HandleReceiveMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	queue, _ :=getString(args, "queue")
	brokerURL, _ :=getString(args, "broker_url")
	if brokerURL == "" {
		brokerURL = "http://localhost:8161/api/message"
	}
	url := brokerURL + "?queue=" + queue
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to receive message: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	return ok("received: " + string(body))
}
