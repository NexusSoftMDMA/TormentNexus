package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func HandleLogExpense(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    amount, found := getFloat64(args, "amount")
    if !found {
        return err("amount パラメータが必要です")
}

    category, _ :=getString(args, "category")
    if !found {
        return err("category パラメータが必要です")
}

    note, _ :=getString(args, "note")
    date, _ :=getString(args, "date")
    if !found {
        date = "today"
    }
    // ここでデータベースにログを記録するロジックを実装
    // 例: db.Exec("INSERT INTO expenses (amount, category, note, date) VALUES (?, ?, ?, ?)", amount, category, note, date)
    return ok("支出をログに記録しました")
}