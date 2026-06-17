package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/sync"
)

func main() {
	cwd, _ := os.Getwd()
	// Find workspace root using filepath.Dir for robustness
	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			// Reached root without finding .git, fallback to cwd
			root = cwd
			break
		}
		root = parent
	}

	syncer := sync.NewRepositorySync(root)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	fmt.Printf("Starting Repository Synchronization for %s\n", root)
	if err := syncer.Run(ctx); err != nil {
		fmt.Printf("Sync FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Repository is SYNCED.")
}
