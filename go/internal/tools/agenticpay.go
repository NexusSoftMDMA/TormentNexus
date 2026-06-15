//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// HandleGetBalance retrieves the balance for a given user.
func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "userId")
	if userID == "" {
		return err("userId is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.agenticpay.com/v1/balance?userId="+userID, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Balance float64 `json:"balance"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok("balance: " + strings.TrimRight(strings.TrimRight(strings.TrimRight(strings.TrimRight("0", "0"), "."), "0"), "."))
}

// HandleCreatePayment creates a payment from one user to another.
func HandleCreatePayment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	amount, _ :=getInt(args, "amount")
	if from == "" || to == "" || amount <= 0 {
		return err("from, to, and positive amount are required")
}

	body := strings.NewReader(`{"from":"` + from + `","to":"` + to + `","amount":` + string(rune(amount)) + `}`)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.agenticpay.com/v1/payment", body)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		ID string `json:"id"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success("payment created with id: " + result.ID)
}
