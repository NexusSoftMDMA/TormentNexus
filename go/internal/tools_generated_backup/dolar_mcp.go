package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetDolarRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rateType, _ :=getString(args, "type")
	url := "https://api.bluelytics.com.ar/v2/latest"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch data")
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode JSON")
}

	var output string
	if rateType == "" {
		output = fmt.Sprintf("%v", result)
	} else {
		data, found := result[rateType].(map[string]interface{})
		if !found {
			return err("rate type not found")
}

		output = fmt.Sprintf("%v", data)

	return ok(output)
}
}