package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandlePredictRot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	topk, _ :=getInt(args, "topk")
	url := "http://localhost:8080/predict_rot?text=" + text + "&topk=" + itoa(topk)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse response failed")
}

	return ok(string(body))
}

func HandleGetTieredRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "http://localhost:8080/tiered_read?id=" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed")
}

	return ok(string(body))
}

func itoa(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf("%d", i)
}