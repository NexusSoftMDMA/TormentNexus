package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPatient(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	patientID, _ :=getInt(args, "patientId")
	url := fmt.Sprintf("https://api.omop.dev/patient/%d", patientID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch patient: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Patient: %v", result))
}

func HandleSearchPatients(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	url := fmt.Sprintf("https://api.omop.dev/patients?name=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(fmt.Sprintf("Search results: %s", string(body)))
}