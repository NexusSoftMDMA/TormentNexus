package tools

import (
    "context"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"
)

// Helper to get string parameter
func getString(args map[string]interface{}, key string) string {
    if v, found := args[key]; found {
        if s, found := v.(string); found {
            return s
        }
    }
    return ""
}

// Helper to get int parameter
func getInt(args map[string]interface{}, key string) int {
    if v, found := args[key]; found {
        switch t := v.(type) {
        case int:
            return t
}
        case float64:
            return int(t)
}
        case string:
            i, _ := strconv.Atoi(t)
            return i
        }
    }
    return 0
}

// Helper to get bool parameter
func getBool(args map[string]interface{}, key string) bool {
    if v, found := args[key]; found {
        switch t := v.(type) {
        case bool:
            return t
}
        case string:
            return strings.ToLower(t) == "true" || t == "1"
        case float64:
            return t == 1
        }
    }
    return false
}

// HandleGetStockData fetches historical stock data from Yahoo Finance
func HandleGetStockData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    symbol, _ :=getString(args, "symbol")
    startDate, _ :=getString(args, "start_date")
    endDate, _ :=getString(args, "end_date")
    
    if symbol == "" {
        return err("symbol is required")

    // Parse dates or use defaults
    var start, end time.Time
    var parseErr error
    
    if startDate != "" {
        start, parseErr = time.Parse("2006-01-02", startDate)
        if parseErr != nil {
            return err("invalid start_date format, use YYYY-MM-DD")

    } else {
        // Default to 1 year ago
        start = time.Now().AddDate(-1, 0, 0)

    if endDate != "" {
        end, parseErr = time.Parse("2006-01-02", endDate)
        if parseErr != nil {
            return err("invalid end_date format, use YYYY-MM-DD")

    } else {
        end = time.Now()

    // Convert to Unix timestamps
    period1 := start.Unix()
    period2 := end.Unix()
    
    // Build Yahoo Finance URL
    baseURL := "https://query1.finance.yahoo.com/v7/finance/download/"
    u := fmt.Sprintf("%s%s?period1=%d&period2=%d&interval=1d&events=history&includeAdjustedClose=true",
        baseURL, url.QueryEscape(symbol), period1, period2)
    
    // Make HTTP request
    client := http.DefaultClient
    req, reqErr := http.NewRequestWithContext(ctx, "GET", u, nil)
    if reqErr != nil {
        return err("failed to create request: " + reqErr.Error())
}

    resp, httpErr := client.Do(req)
    if httpErr != nil {
        return err("failed to fetch data: " + httpErr.Error())
}

    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return err(fmt.Sprintf("yahoo finance returned status %d: %s", resp.StatusCode, string(body)))
}

    // Parse CSV
    reader := csv.NewReader(resp.Body)
    records, csvErr := reader.ReadAll()
    if csvErr != nil {
        return err("failed to parse CSV: " + csvErr.Error())
}

    if len(records) < 2 {
        return err("no data returned")
}

    // Convert to JSON structure
    type DataPoint struct {
        Date   string  `json:"date"`
        Open   float64 `json:"open"`
        High   float64 `json:"high"`
        Low    float64 `json:"low"`
        Close  float64 `json:"close"`
        AdjClose float64 `json:"adj_close"`
        Volume int64   `json:"volume"`
    }
    
    headers := records[0]
    data := make([]DataPoint, 0, len(records)-1)
    
    for i := 1; i < len(records); i++ {
        row := records[i]
        if len(row) < 6 {
            continue
        }
        dp := DataPoint{
            Date: row[0],
        }
        // Parse numeric fields
        if open, e := strconv.ParseFloat(row[1], 64); e == nil {
            dp.Open = open
        }
        if high, e := strconv.ParseFloat(row[2], 64); e == nil {
            dp.High = high
        }
        if low, e := strconv.ParseFloat(row[3], 64); e == nil {
            dp.Low = low
        }
        if close, e := strconv.ParseFloat(row[4], 64); e == nil {
            dp.Close = close
        }
        if adjClose, e := strconv.ParseFloat(row[5], 64); e == nil {
            dp.AdjClose = adjClose
        }
        if len(row) > 6 {
            if vol, e := strconv.ParseInt(row[6], 10, 64); e == nil {
                dp.Volume = vol
            }
        }
        data = append(data, dp)

    result, jsonErr := json.Marshal(data)
    if jsonErr != nil {
        return err("failed to marshal result: " + jsonErr.Error())
}

    return ok(string(result))
}

}
}
}

// HandleCalculateIndicator calculates technical indicators for a stock
func HandleCalculateIndicator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    symbol, _ :=getString(args, "symbol")
    indicator, _ :=getString(args, "indicator")
    period, _ :=getInt(args, "period")
    
    if symbol == "" {
        return err("symbol is required")
}

    if indicator == "" {
        return err("indicator is required")
}

    if period <= 0 {
        period = 14 // default
    }
    
    // First get stock data (last 200 days for indicator calculation)
    dataArgs := map[string]interface{}{
        "symbol": symbol,
        "start_date": time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
        "end_date": time.Now().Format("2006-01-02"),
    }
    
    dataResp, dataErr := HandleGetStockData(ctx, dataArgs)
    if dataErr != nil {
        return err("failed to get stock data: " + dataErr.Error())
}

    // Parse the data
    var dataPoints []struct {
        Date  string  `json:"date"`
        Close float64 `json:"close"`
    }
    
    if parseErr := json.Unmarshal([]byte(dataResp.(TextContent).Text), &dataPoints); parseErr != nil {
        return err("failed to parse stock data: " + parseErr.Error())
}

    if len(dataPoints) < period {
        return err(fmt.Sprintf("insufficient data: need at least %d points, got %d", period, len(dataPoints)))
}

    var result []float64
    switch strings.ToLower(indicator) {
    case "sma":
        // Simple Moving Average
        result = make([]float64, len(dataPoints))
        for i := 0; i < len(dataPoints); i++ {
            if i < period-1 {
                result[i] = 0
                continue
            }
            sum := 0.0
            for j := i - period + 1; j <= i; j++ {
                sum += dataPoints[j].Close
            }
            result[i] = sum / float64(period)

    case "rsi":
        // Relative Strength Index
        result = make([]float64, len(dataPoints))
        changes := make([]float64, len(dataPoints))
        for i := 1; i < len(dataPoints); i++ {
            changes[i] = dataPoints[i].Close - dataPoints[i-1].Close
        }
        
        for i := period; i < len(dataPoints); i++ {
            gain := 0.0
            loss := 0.0
            for j := i - period + 1; j <= i; j++ {
                if changes[j] > 0 {
                    gain += changes[j]
                } else {
                    loss -= changes[j] // absolute value
                }
            }
            avgGain := gain / float64(period)
            avgLoss := loss / float64(period)
            
            if avgLoss == 0 {
                result[i] = 100
            } else {
                rs := avgGain / avgLoss
                result[i] = 100 - (100 / (1 + rs))

        }
        
    default:
        return err("unsupported indicator: " + indicator + " (supported: sma, rsi)")
}

    // Format output as JSON array
    type IndicatorResult struct {
        Date  string  `json:"date"`
        Value float64 `json:"value"`
    }
    
    formatted := make([]IndicatorResult, len(dataPoints))
    for i, dp := range dataPoints {
        formatted[i] = IndicatorResult{
            Date:  dp.Date,
            Value: result[i],
        }
    }
    
    out, jsonErr := json.Marshal(formatted)
    if jsonErr != nil {
        return err("failed to marshal indicator result: " + jsonErr.Error())
}

    return ok(string(out))
}

}
}

// HandleScreenStocks screens stocks based on criteria
func HandleScreenStocks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    // This is a simplified implementation that would normally query a database
    // For now, we'll return a mock response with example logic
    
    minPrice := getFloat(args, "min_price")
    maxPrice := getFloat(args, "max_price")
    minVolume, _ :=getInt(args, "min_volume")
    sector, _ :=getString(args, "sector")
    
    // In a real implementation, this would query a database of stock metadata
    // For this example, we'll return a static list with filtering logic
    
    // Mock database of stocks
    type StockInfo struct {
        Symbol string  `json:"symbol"`
        Name   string  `json:"name"`
        Price  float64 `json:"price"`
        Volume int64   `json:"volume"`
        Sector string  `json:"sector"`
    }
    
    mockStocks := []StockInfo{
        {"AAPL", "Apple Inc.", 175.50, 52000000, "Technology"},
        {"MSFT", "Microsoft Corp.", 378.20, 28000000, "Technology"},
        {"GOOGL", "Alphabet Inc.", 140.30, 25000000, "Technology"},
        {"AMZN", "Amazon.com Inc.", 155.80, 35000000, "Consumer Cyclical"},
        {"TSLA", "Tesla Inc.", 242.70, 110000000, "Consumer Cyclical"},
        {"JPM", "JPMorgan Chase", 185.60, 12000000, "Financial Services"},
        {"JNJ", "Johnson & Johnson", 156.40, 8500000, "Healthcare"},
        {"V", "Visa Inc.", 275.30, 6500000, "Financial Services"},
        {"PG", "Procter & Gamble", 150.20, 7800000, "Consumer Defensive"},
        {"NVDA", "NVIDIA Corp.", 495.80, 45000000, "Technology"},
    }
    
    // Apply filters
    filtered := make([]StockInfo, 0, len(mockStocks))
    for _, stock := range mockStocks {
        if minPrice > 0 && stock.Price < minPrice {
            continue
        }
        if maxPrice > 0 && stock.Price > maxPrice {
            continue
        }
        if minVolume > 0 && stock.Volume < int64(minVolume) {
            continue
        }
        if sector != "" && !strings.EqualFold(stock.Sector, sector) {
            continue
        }
        filtered = append(filtered, stock)

    result, jsonErr := json.Marshal(filtered)
    if jsonErr != nil {
        return err("failed to marshal screening results: " + jsonErr.Error())
}

    return ok(string(result))
}
}