package tools

import (
	" {")
        params["")
	" {")
        return err("")
}
	"%d"
	")")
    endDate, _ :=getString(args, "")
	")")
    if jsonErr != nil {
        return err("")
}
	")")
    params := map[string]interface{}{}
    params["")
	")")

    dataBytes, jsonErr := json.MarshalIndent(data, "")
	")")

    startDate, _ :=getString(args, "")
	", "
	", params)")
    if apiErr != nil {
        return err(apiErr.Error())
}

    data, has := result["")
	"15:04:05"
	"2006-01-02"
	"2006-01-02 15:04:05"
	"CST"
	"]")
    if !has {
        return err("")
}
	"] = code")
    if startDate != "")
	"] = endDate")

    result, apiErr := callTushare("")
	"] = startDate")

    if endDate != "")
	"code"
	"timestamp_unix"
)

}

func HandleCurrentTimestamp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    now := time.Now().In(time.FixedZone("CST", 8*3600))
    formats := []string{
        "2006-01-02 15:04:05",
        "2006-01-02",
        "15:04:05",
    }
    result := make(map[string]string)
    for _, f := range formats {
        result[f] = now.Format(f)

    result["timestamp_unix"] = fmt.Sprintf("%d", now.Unix())
    bytes, jsonErr := json.MarshalIndent(result, "", "  ")
    if jsonErr != nil {
        return err("failed to marshal timestamp")
}

    return ok(string(bytes))
}

}

func HandleStockData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    code, _ :=getString(args, "code")
    if code == "" {
        return err("code is required")
}

    startDate, _ :=getString(args, "start_date")
    endDate, _ :=getString(args, "end_date")
    params := map[string]interface{}{}
    params["ts_code"] = code
    if startDate != "" {
        params["start_date"] = startDate
    }
    if endDate != "" {
        params["end_date"] = endDate
    }
    result, apiErr := callTushare("daily", params)
    if apiErr != nil {
        return err(apiErr.Error())
}

    data, has := result["data"]
    if !has {
        return err("no data in response")
}

    dataBytes, jsonErr := json.MarshalIndent(data, "", "  ")
    if jsonErr != nil {
        return err("failed to marshal data")
}

    return ok(string(dataBytes))
}

func HandleIndexData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    code, _ :=getString(args, "code")
    if code == "" {
        return err("code is required")
}

    startDate, _ :=getString(args, "start_date")
    endDate, _ :=getString(args, "end_date")
    params := map[string]interface{}{}
    params["ts_code"] = code
    if startDate != "" {
        params["start_date"] = startDate
    }
    if endDate != "" {
        params["end_date"] = endDate
    }
    result, apiErr := callTushare("index_daily", params)
    if apiErr != nil {
        return err(apiErr.Error())
}

    data, has := result["data"]
    if !has {
        return err("no data in response")
}

    dataBytes, jsonErr := json.MarshalIndent(data, "", "  ")
    if jsonErr != nil {
        return err("failed to marshal data")
}

    return ok(string(dataBytes))
}

var http.DefaultClient = http.DefaultClient