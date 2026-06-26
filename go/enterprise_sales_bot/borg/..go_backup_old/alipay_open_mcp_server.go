package tools

import (
    "context"
    "fmt"
)

func HandleGetAlipayAccountInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    userID, _ :=getString(args, "user_id")
    if userID == "" {
        return err("user_id is required")
}

    return ok(fmt.Sprintf(`{"user_id":"%s","balance":"100.00"}`, userID))
}

func HandleAlipayTransfer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    toAccount, _ :=getString(args, "to_account")
    amount, _ :=getString(args, "amount")
    if toAccount == "" || amount == "" {
        return err("to_account and amount required")
}

    return ok(fmt.Sprintf(`{"result":"success","to":"%s","amount":"%s"}`, toAccount, amount))
}