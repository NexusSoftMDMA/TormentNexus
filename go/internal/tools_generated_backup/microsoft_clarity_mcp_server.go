package tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleGetClarityData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "projectId")
	apiKey, _ :=getString(args, "apiKey")
	startDate, _ :=getString(args, "startDate")
	endDate, _ :=getString(args, "endDate")
	url := "https://clarity.microsoft.com/api/v1/projects/" + projectID + "/export?start=" + startDate + "&end=" + endDate
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success("Data retrieved successfully")
}