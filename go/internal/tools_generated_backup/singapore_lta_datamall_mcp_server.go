package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetBusArrivals(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	busStopCode, _ :=getString(args, "busStopCode")
	serviceNo, _ :=getString(args, "serviceNo")

	req, e := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://datamall2.mytransport.sg/ltaodataservice/BusArrivalv2?BusStopCode=%s&ServiceNo=%s", busStopCode, serviceNo), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("AccountKey", os.Getenv("LTA_ACCOUNT_KEY"))

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}

	return ok(string(body))
}

func HandleGetTrafficIncidents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://datamall2.mytransport.sg/ltaodataservice/TrafficIncidents", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("AccountKey", os.Getenv("LTA_ACCOUNT_KEY"))

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json decode error: " + e.Error())
	}

	return success(string(body))
}