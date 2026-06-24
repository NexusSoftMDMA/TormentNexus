package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleNasdaqDataLinkGetData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataset, _ :=getString(args, "dataset_code")
	apiKey, _ :=getString(args, "api_key")
	if dataset == "" || apiKey == "" {
		return err("missing required parameters: dataset_code and api_key")
}

	url := fmt.Sprintf("https://data.nasdaq.com/api/v3/datasets/%s/data.json?api_key=%s", dataset, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status: " + resp.Status)
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body: " + e.Error())
}

	return success(string(body))
}

func HandleNasdaqDataLinkGetMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataset, _ :=getString(args, "dataset_code")
	apiKey, _ :=getString(args, "api_key")
	if dataset == "" || apiKey == "" {
		return err("missing required parameters: dataset_code and api_key")
}

	url := fmt.Sprintf("https://data.nasdaq.com/api/v3/datasets/%s.json?api_key=%s", dataset, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status: " + resp.Status)
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response body: " + e.Error())
}

	return success(string(body))
}