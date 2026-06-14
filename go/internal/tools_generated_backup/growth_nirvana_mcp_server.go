package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleReadAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accountID, _ :=getInt(args, "account_id")
	url := fmt.Sprintf("https://api.growthnirvana.com/api/v1/accounts/%d", accountID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch account: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}