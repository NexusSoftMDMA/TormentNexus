package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleExecuteCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	language, _ :=getString(args, "language")
	code, _ :=getString(args, "code")
	version, _ :=getString(args, "version")
	if language == "" || code == "" {
		return err("language and code are required")
}

	payload := map[string]interface{}{
		"language": language,
		"files":    []map[string]string{{"content": code}},
	}
	if version != "" {
		payload["version"] = version
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post("https://emkc.org/api/v2/piston/execute", "application/json", strings.NewReader(string(body)))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(respBody)))
}

	var result map[string]interface{}
	if e = json.Unmarshal(respBody, &result); e != nil {
		return err(fmt.Sprintf("unmarshal error: %v", e))
}

	var output string
	if run, found := result["run"]; found {
		if runMap, found := run.(map[string]interface{}); found {
			if stdout, found := runMap["stdout"]; found {
				output = fmt.Sprintf("%v", stdout)

			if stderr, found := runMap["stderr"]; found {
				if output != "" {
					output += "\n"
				}
				output += fmt.Sprintf("stderr: %v", stderr)

		}
	}
	return ok(output)
}
}
}