package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		targetURL = "http://localhost:8081"
	}

	fmt.Printf("Starting smoke test for: %s\n", targetURL)

	// 1. Verify basic health
	err := verifyEndpoint(fmt.Sprintf("%s/health", targetURL), "OK\n")
	if err != nil {
		fmt.Printf("Smoke Test Failed: %v\n", err)
		os.Exit(1)
	}

	// 2. Verify detailed health (DB connection)
	err = verifyDetailedHealth(fmt.Sprintf("%s/health/detailed", targetURL))
	if err != nil {
		fmt.Printf("Smoke Test DB Validation Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Smoke Test Successful.")
}

func verifyEndpoint(url, expected string) error {
	resp, err := http.Get(url) // #nosec G107
	// #nosec G704
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Basic check for response body
	body := make([]byte, len(expected))
	_, err = resp.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		return err
	}

	if string(body) != expected {
		return fmt.Errorf("unexpected response body: expected %q, got %q", expected, string(body))
	}

	return nil
}

func verifyDetailedHealth(url string) error {
	resp, err := http.Get(url) // #nosec G107
	// #nosec G704
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return err
	}

	if health["database"] != "OK" {
		return fmt.Errorf("database health is not OK: %v", health["database"])
	}

	if health["workers"] != "active" {
		return fmt.Errorf("workers are not active: %v", health["workers"])
	}

	return nil
}
