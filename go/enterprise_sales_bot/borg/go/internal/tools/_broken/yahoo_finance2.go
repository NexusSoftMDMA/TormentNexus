package tools

// ... (import statements)

// ... (handler functions)

// Manifest
const manifest = `
{
  "filename": "yahoo_finance2.go",
  "server_name": "yahoo_finance2",
  "handlers": [
    {
      "tool_name": "getString",
      "handler_func": "HandleGetString",
      "description": "Get a string value from a key"
    },
    {
      "tool_name": "getInt",
      "handler_func": "HandleGetInt",
      "description": "Get an integer value from a key"
    },
    {
      "tool_name": "getBool",
      "handler_func": "HandleGetBool",
      "description": "Get a boolean value from a key"
    },
    {
      "tool_name": "getTextContent",
      "handler_func": "HandleGetTextContent",
      "description": "Get text content from a file"
    },
    {
      "tool_name": "getIntContent",
      "handler_func": "HandleGetIntContent",
      "description": "Get an integer value from a file"
    },
    {
      "tool_name": "getBoolContent",
      "handler_func": "HandleGetBoolContent",
      "description": "Get a boolean value from a file"
    },
    {
      "tool_name": "getTextContentFromURL",
      "handler_func": "HandleGetTextContentFromURL",
      "description": "Get text content from a URL"
    },
    {
      "tool_name": "getIntContentFromURL",
      "handler_func": "HandleGetIntContentFromURL",
      "description": "Get an integer value from a URL"
    },
    {
      "tool_name": "getBoolContentFromURL",
      "handler_func": "HandleGetBoolContentFromURL",
      "description": "Get a boolean value from a URL"
    }
  ]
}
`

// ... (rest of the code)

func HandleGetString(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	key := args["key"]
	val, e := parity.GetString(key)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{TextContent: val}, nil
}

func HandleGetInt(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	key := args["key"]
	val, e := parity.GetInt(key)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{IntContent: val}, nil
}

func HandleGetBool(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	key := args["key"]
	val, e := parity.GetBool(key)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{BoolContent: val}, nil
}

func HandleGetTextContent(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	filePath := args["file"]
	file, e := os.Open(filePath)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	defer file.Close()
	content, e := io.ReadAll(file)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{TextContent: string(content)}, nil
}

func HandleGetIntContent(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	filePath := args["file"]
	file, e := os.Open(filePath)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	defer file.Close()
	content, e := io.ReadAll(file)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	val, e := strconv.Atoi(string(content))
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{IntContent: val}, nil
}

func HandleGetBoolContent(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	filePath := args["file"]
	file, e := os.Open(filePath)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	defer file.Close()
	content, e := io.ReadAll(file)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	val, e := strconv.ParseBool(string(content))
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{BoolContent: val}, nil
}

func HandleGetTextContentFromURL(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	urlStr := args["url"]
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	defer resp.Body.Close()
	content, e := io.ReadAll(resp.Body)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{TextContent: string(content)}, nil
}

func HandleGetIntContentFromURL(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	urlStr := args["url"]
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	defer resp.Body.Close()
	content, e := io.ReadAll(resp.Body)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	val, e := strconv.Atoi(string(content))
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{IntContent: val}, nil
}

func HandleGetBoolContentFromURL(ctx context.Context, args map[string]interface{}) (parity.ToolResponse, error) {
	urlStr := args["url"]
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	client := http.DefaultClient
	resp, e := client.Do(req)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	defer resp.Body.Close()
	content, e := io.ReadAll(resp.Body)
	if e != nil {
		return parity.ToolResponse{}, e
	}
	val, e := strconv.ParseBool(string(content))
	if e != nil {
		return parity.ToolResponse{}, e
	}
	return parity.ToolResponse{BoolContent: val}, nil
}

func HandleQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    symbol, _ :=getString(args, "symbol")
    if symbol == "" {
        return err("missing required argument: symbol")
}

    client := http.DefaultClient
    apiURL := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s", url.QueryEscape(symbol))
    // But note: Do NOT use url.QueryEscape? The instruction says: "Do NOT use url.QueryEscape — use net/url's url.Values instead". So I should use url.Values.
    // Actually, url.QueryEscape is from net/url, but they say not to use it. They say use url.Values. So I'll do:
    // vals := url.Values{}
    // vals.Set("symbols", symbol)
    // apiURL := "https://query1.finance.yahoo.com/v7/finance/quote?" + vals.Encode()
    // That's better.
    vals := url.Values{}
    vals.Set("symbols", symbol)
    apiURL := "https://query1.finance.yahoo.com/v7/finance/quote?" + vals.Encode()
    body, fetchErr := fetchURL(ctx, client, apiURL)
    if fetchErr != nil {
        return err(fmt.Sprintf("failed to fetch quote: %v", fetchErr))
}

    // Parse JSON
    var resp struct {
        QuoteResponse struct {
            Result []struct {
                Symbol             string  `json:"symbol"`
                RegularMarketPrice float64 `json:"regularMarketPrice"`
                RegularMarketChange float64 `json:"regularMarketChange"`
                RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
                Currency           string  `json:"currency"`
                LongName           string  `json:"longName"`
            } `json:"result"`
            Error interface{} `json:"error"`
        } `json:"quoteResponse"`
    }
    parseErr := json.Unmarshal(body, &resp)
    if parseErr != nil {
        return err(fmt.Sprintf("failed to parse quote response: %v", parseErr))
}

    if len(resp.QuoteResponse.Result) == 0 {
        return err("no quote data found for symbol")
}

    q := resp.QuoteResponse.Result[0]
    // Format output
    changeStr := ""
    if q.RegularMarketChange >= 0 {
        changeStr = fmt.Sprintf("+%.2f (+%.2f%%)", q.RegularMarketChange, q.RegularMarketChangePercent)
    } else {
        changeStr = fmt.Sprintf("%.2f (%.2f%%)", q.RegularMarketChange, q.RegularMarketChangePercent)

    out := fmt.Sprintf("Symbol: %s\nName: %s\nPrice: %s %.2f\nChange: %s",
        q.Symbol, q.LongName, q.Currency, q.RegularMarketPrice, changeStr)
    return ok(out)
}
}