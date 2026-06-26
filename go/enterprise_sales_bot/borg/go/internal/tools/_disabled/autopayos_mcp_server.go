package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleCreatePayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	amount, _ :=getString(args, "amount")
	currency, _ :=getString(args, "currency")
	source, _ :=getString(args, "source")
	body := fmt.Sprintf(`{"amount":"%s","currency":"%s","source":"%s"}`, amount, currency, source)
	req, e := http.NewRequestWithContext(ctx, "POST", os.Getenv("AUTOPAYOS_URL")+"/payments", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("API error: " + string(data))
}

	return ok(string(data))
}

func HandleGetPayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paymentID, _ :=getString(args, "paymentId")
	req, e := http.NewRequestWithContext(ctx, "GET", os.Getenv("AUTOPAYOS_URL")+"/payments/"+paymentID, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode >= 400 {
		return err("API error: " + string(data))
}

	var payment map[string]interface{}
	if e := json.Unmarshal(data, &payment); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	return success(payment)
}