package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleLedgerBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account")
	url := fmt.Sprintf("http://localhost:8080/balance?account=%s", account)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch balance: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}