package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleGetVaultBalance retrieves the current balance of a Factor vault.
func HandleGetVaultBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vaultID, _ :=getString(args, "vaultId")
	if vaultID == "" {
		return err("vaultId is required")
}

	url := fmt.Sprintf("https://api.factorprotocol.com/v1/vault/%s/balance", vaultID)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Balance string `json:"balance"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok("Vault balance: " + result.Balance)
}

// HandleDepositToVault initiates a deposit to a Factor vault.
func HandleDepositToVault(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vaultID, _ :=getString(args, "vaultId")
	amount, _ :=getString(args, "amount")
	if vaultID == "" || amount == "" {
		return err("vaultId and amount are required")
}

	url := fmt.Sprintf("https://api.factorprotocol.com/v1/vault/%s/deposit", vaultID)
	payload := map[string]string{"amount": amount}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("deposit failed with status: " + resp.Status)
}

	return success("Deposit initiated")
}