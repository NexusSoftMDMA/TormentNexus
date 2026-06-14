package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStartLocustTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "http://localhost:8089"
	}
	users, _ :=getInt(args, "users")
	if users == 0 {
		users = 10
	}
	spawnRate, _ :=getInt(args, "spawn_rate")
	if spawnRate == 0 {
		spawnRate = 1
	}
	url := fmt.Sprintf("%s/swarm", host)
	body := fmt.Sprintf(`{"user_count":%d,"spawn_rate":%d}`, users, spawnRate)
	resp, e := http.DefaultClient.Post(url, "application/json", stringToReader(body))
	if e != nil {
		return err("failed to start Locust test: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Locust returned status " + resp.Status)
}

	return ok("Locust test started with %d users at spawn rate %d", users, spawnRate)
}

func HandleGetLocustStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "http://localhost:8089"
	}
	resp, e := http.DefaultClient.Get(host + "/stats/requests")
	if e != nil {
		return err("failed to get Locust stats: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("failed to parse stats: " + e.Error())
}

	return success(result)
}

func stringToReader(s string) io.Reader {
	return &stringReader{s}
}

type stringReader struct{ s string }

func (r *stringReader) Read(p []byte) (n int, e error) {
	if len(r.s) == 0 {
		return 0, io.EOF
	}
	n = copy(p, r.s)
	r.s = r.s[n:]
	return n, nil
}