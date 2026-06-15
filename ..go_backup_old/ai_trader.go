package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleRunBacktest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	strategy, _ :=getString(args, "strategy")
	symbol, _ :=getString(args, "symbol")
	startDate, _ :=getString(args, "start_date")
	endDate, _ :=getString(args, "end_date")
	cash, _ :=getInt(args, "initial_cash")
	if strategy == "" || symbol == "" {
		return err("strategy and symbol are required")
	}
	resp, e := http.DefaultClient.Get(fmt.Sprintf("http://localhost:8080/backtest?strategy=%s&symbol=%s&start=%s&end=%s&cash=%d", strategy, symbol, startDate, endDate, cash))
	if e != nil {
		return err("failed to run backtest: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
	}
	return ok(fmt.Sprintf("Backtest completed: %+v", result))
}

func HandleListStrategies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	strategies := []string{"SMA_Cross", "RSI_MeanReversion", "MACD", "BollingerBands", "PairTrading", "Momentum"}
	return success(fmt.Sprintf("Available strategies: %v", strategies))
}