package plugins

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadGoPlugin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go plugins (-buildmode=plugin) are not supported on Windows")
	}

	// Build the plugin first
	pluginDir := "testdata"
	_ = os.MkdirAll(pluginDir, 0755)
	defer func() { _ = os.RemoveAll(pluginDir) }()

	pluginPath := filepath.Join(pluginDir, "example.so")

	// We'll reuse the example plugin we just built since building it from within tests can be finicky with caching
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginPath, "./example/plugin.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build plugin: %v", err)
	}

	reg := NewRegistry()
	err := reg.LoadGoPlugin("example", pluginPath)
	if err != nil {
		t.Fatalf("LoadGoPlugin failed: %v", err)
	}

	if len(reg.Sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(reg.Sources))
	}

	source := reg.Sources["example"]


	comps, err := source.Discover(context.Background(), nil)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(comps) != 1 || comps[0].Name != "PluginCompany" {
		t.Errorf("unexpected discovery result: %+v", comps)
	}
}
