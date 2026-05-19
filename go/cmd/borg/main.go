package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/borghq/borg-go/internal/buildinfo"
	"github.com/borghq/borg-go/internal/config"
	"github.com/borghq/borg-go/internal/controlplane"
	"github.com/borghq/borg-go/internal/httpapi"
	"github.com/borghq/borg-go/internal/lockfile"
	"github.com/borghq/borg-go/internal/sessionimport"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	command := "serve"
	if len(args) > 0 {
		switch args[0] {
		case "serve", "version":
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
	default:
		log.Printf("unknown command %q", command)
		return 1
	}
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
	if err := server.ListenAndServe(ctx); err != nil {
		log.Printf("server failed: %v", err)
		return 1
	}

	return 0
}
