package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/buildinfo"
	"github.com/tormentnexushq/tormentnexus-go/internal/config"
	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
	"github.com/tormentnexushq/tormentnexus-go/internal/httpapi"
	"github.com/tormentnexushq/tormentnexus-go/internal/license"
	"github.com/tormentnexushq/tormentnexus-go/internal/lockfile"
	"github.com/tormentnexushq/tormentnexus-go/internal/sessionimport"
	"path/filepath"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	command := "serve"
	if len(args) > 0 {
		if strings.HasPrefix(args[0], "tormentnexus://") {
			return runDeepLink(args[0])
		}
		switch args[0] {
		case "serve", "version", "start", "stop", "status", "mcp":
			command = args[0]
			args = args[1:]
		}
	}

	switch command {
	case "version":
		fmt.Println(buildinfo.Version)
		return 0
	case "serve":
		return runServe(args)
	case "start":
		return cmdStart(args)
	case "stop":
		return cmdStop(args)
	case "status":
		return cmdStatus(args)
	case "mcp":
		return cmdMCP(args)
	default:
		log.Printf("unknown command %q", command)
		return 1
	}
}

func runDeepLink(deepLink string) int {
	cfg := config.Default()
	record, err := lockfile.Read(cfg.LockPath())
	if err != nil {
		log.Printf("TormentNexus sidecar server is not currently running. Please start it using 'tormentnexus serve' first.")
		return 1
	}

	targetURL := fmt.Sprintf("http://%s:%d/api/native/protocol/tormentnexus", record.Host, record.Port)

	payload, err := json.Marshal(map[string]string{"url": deepLink})
	if err != nil {
		log.Printf("Failed to marshal deep link payload: %v", err)
		return 1
	}

	resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to dispatch deep link to running TormentNexus server: %v", err)
		return 1
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned error (%d): %s", resp.StatusCode, string(body))
		return 1
	}

	log.Printf("Deep link dispatched successfully: %s", string(body))
	return 0
}

func runServe(args []string) int {
	cfg := config.Default()

	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.StringVar(&cfg.Host, "host", cfg.Host, "Host to bind the experimental Go cli-orchestrator port to.")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "Port to bind the experimental Go cli-orchestrator port to.")
	fs.StringVar(&cfg.ConfigDir, "config-dir", cfg.ConfigDir, "Config directory for the experimental Go cli-orchestrator port.")
	if err := fs.Parse(args); err != nil {
		log.Printf("failed to parse flags: %v", err)
		return 2
	}

	// Verify offline license if present in workspace root
	licPath := filepath.Join(cfg.WorkspaceRoot, "tormentnexus.lic")
	if _, err := os.Stat(licPath); err == nil {
		lic, err := license.VerifyLicense(cfg.WorkspaceRoot)
		if err != nil {
			log.Printf("License validation failed: %v", err)
			return 1
		}
		log.Printf("Licensed to: %s (Seats: %d, Expires: %s)", lic.Holder, lic.Seats, lic.ExpiresAt.Format(time.RFC822))
	} else {
		log.Printf("No tormentnexus.lic license file found. Running under free limitations.")
	}

	record := lockfile.Record{
		Host:      cfg.Host,
		Port:      cfg.Port,
		Version:   buildinfo.Version,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := lockfile.Write(cfg.LockPath(), record); err != nil {
		log.Printf("failed to write lock file: %v", err)
		return 1
	}
	defer func() {
		_ = os.Remove(cfg.LockPath())
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	detector := controlplane.NewDetector(1500*time.Millisecond, 30*time.Minute)
	server := httpapi.New(cfg, detector)

	// Pre-warm caches in the background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = detector.DetectAll(ctx)
	}()
	go func() {
		// Pre-warm import scan cache
		homeDir, _ := os.UserHomeDir()
		if homeDir == "" {
			homeDir = cfg.MainConfigDir
		}
		scanner := sessionimport.NewScanner(cfg.WorkspaceRoot, homeDir, 50)
		results, _ := scanner.ScanValidated()
		server.PreWarmImportCache(results)
	}()

	log.Printf(
		"Experimental Go cli-orchestrator port listening on %s (index: %s/api/index, runtime: %s/api/runtime/status, cli: %s/api/cli/summary, import: %s/api/import/summary, providers: %s/api/providers/routing-summary)",
		cfg.BaseURL(),
		cfg.BaseURL(),
		cfg.BaseURL(),
		cfg.BaseURL(),
		cfg.BaseURL(),
		cfg.BaseURL(),
	)
	server.PreWarmCaches()
	if err := server.ListenAndServe(ctx); err != nil {
		log.Printf("server failed: %v", err)
		return 1
	}

	return 0
}
