package main

import (
	"context"
	"fmt"
	"time"
)

// App is the main Wails application struct with TormentNexus bindings.
type App struct {
	ctx      context.Context
	started  time.Time
	services *TormentNexusServices
}

// TormentNexusServices holds references to the TormentNexus Go-sidecar services.
type TormentNexusServices struct {
	Status     func() map[string]interface{}
	MemoryInfo func() map[string]interface{}
	HealerInfo func() map[string]interface{}
	PeerCount  func() int
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{
		started: time.Now(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// BindServices connects the Wails app to TormentNexus core services.
func (a *App) BindServices(svc *TormentNexusServices) {
	a.services = svc
}

// Greet returns a greeting for the given name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetSystemStatus returns the current system status for the native dashboard.
func (a *App) GetSystemStatus() map[string]interface{} {
	result := map[string]interface{}{
		"status":   "running",
		"uptime":   time.Since(a.started).String(),
		"version":  "1.0.0-alpha.62",
		"protocol": "tormentnexus://",
	}

	if a.services != nil && a.services.Status != nil {
		for k, v := range a.services.Status() {
			result[k] = v
		}
	}

	return result
}

// GetMemoryStatus returns L1/L2/L3 memory tier information.
func (a *App) GetMemoryStatus() map[string]interface{} {
	if a.services != nil && a.services.MemoryInfo != nil {
		return a.services.MemoryInfo()
	}
	return map[string]interface{}{
		"l1_status": "unavailable",
		"l2_status": "unavailable",
		"l3_status": "unavailable",
	}
}

// GetHealerStatus returns the immune system status.
func (a *App) GetHealerStatus() map[string]interface{} {
	if a.services != nil && a.services.HealerInfo != nil {
		return a.services.HealerInfo()
	}
	return map[string]interface{}{
		"active_pathogens": 0,
		"immune_status":    "unknown",
	}
}

// GetMeshStatus returns A2A mesh peer discovery status.
func (a *App) GetMeshStatus() map[string]interface{} {
	result := map[string]interface{}{
		"mesh_available": false,
		"peer_count":     0,
	}
	if a.services != nil && a.services.PeerCount != nil {
		result["mesh_available"] = true
		result["peer_count"] = a.services.PeerCount()
	}
	return result
}
