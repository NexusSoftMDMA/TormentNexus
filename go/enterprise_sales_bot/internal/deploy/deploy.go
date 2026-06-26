package deploy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// Deployer handles self-service deployment operations.
type Deployer struct {
	tracker		CITracker
	dispatcher	WorkflowDispatcher
}

// NewDeployer creates a new Deployer instance.
func NewDeployer(tracker CITracker, dispatcher WorkflowDispatcher) *Deployer {
	return &Deployer{
		tracker:	tracker,
		dispatcher:	dispatcher,
	}
}

// ValidateEnvironment checks for required environment variables for production.
func (d *Deployer) ValidateEnvironment() error {
	required := []string{"DATABASE_URL", "GITHUB_TOKEN", "GITHUB_WEBHOOK_SECRET"}
	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("missing required environment variable: %s", env)
		}
	}
	return nil
}

// ExecuteSync performs a remote sync and submodule update.
func (d *Deployer) ExecuteSync() error {
	slog.Info("Deployment: Initiating repository sync...")
	if err := gitcheck.SyncRemote(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	if err := gitcheck.UpdateSubmodules(); err != nil {
		return fmt.Errorf("submodule update failed: %w", err)
	}
	slog.Info("Deployment: Sync successful.")
	return nil
}

// ExecuteBuild triggers the project build process.
// MonitorDeployment periodically checks the health of the deployment.
func (d *Deployer) MonitorDeployment(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("Deployment monitor started (interval: %v)...", interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("Deployment monitor stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			health, err := d.tracker.GetSystemHealth(ctx)
			if err != nil {
				slog.Info(fmt.Sprintf("Deployment Monitor: Health check failed: %v", err))
				continue
			}
			slog.Info(fmt.Sprintf("Deployment Monitor: System Health: %s", health))

			// Autonomous Deployment: If main is successful and we are out of sync, trigger build
			if health == "Main branch status: Success" {
				// In a real autonomous system, we'd check if local bin matches remote
				// For now, we just log the readiness.
				slog.Info("Deployment Monitor: System is healthy and ready for deployment.")
			}
		}
	}
}

func (d *Deployer) ExecuteBuild() error {
	slog.Info("Deployment: Initiating build...")

	// If a dispatcher is available, we trigger the remote deploy workflow
	if d.dispatcher != nil {
		slog.Info("Deployment: Dispatching remote deployment workflow...")
		err := d.dispatcher.Dispatch(context.Background(), "deploy.yml", "main", nil)
		if err == nil {
			slog.Info("Deployment: Remote deployment dispatched successfully.")
			return nil
		}
		slog.Info(fmt.Sprintf("Deployment Warning: Remote dispatch failed, falling back to local build: %v", err))
	}

	// Fallback to local 'go build'
	cmd := exec.Command("go", "build", "-v", "-o", "bin/sales_bot", "./cmd/sales_bot")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v, output: %s", err, string(output))
	}
	slog.Info("Deployment: Local build successful.")
	return nil
}

// Run starts the background synchronization worker.
func (d *Deployer) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("Deployment worker started (interval: %v)...", interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("Deployment worker stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			if err := d.ExecuteSync(); err != nil {
				slog.Info(fmt.Sprintf("Deployment background sync failed: %v", err))
			}
		}
	}
}
