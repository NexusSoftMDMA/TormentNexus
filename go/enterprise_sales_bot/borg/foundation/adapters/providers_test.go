package adapters

import (
	"os"
	"strings"
	"testing"
)

func TestBuildProviderStatusAndContext(t *testing.T) {
	setenv(t, "SUPERCLI_PROVIDER", "openai")
	setenv(t, "SUPERCLI_MODEL", "gpt-4o")
	setenv(t, "OPENAI_API_KEY", "test-key")
	status := BuildProviderStatus()
	if status.CurrentProvider != "openai" {
		t.Fatalf("unexpected provider: %#v", status)
	}
	if !status.HasAPIKey {
		t.Fatal("expected api key detection")
	}
	if len(status.Available) == 0 {
		t.Fatal("expected available providers")
	}
	context := BuildProviderContext()
	if !strings.Contains(context, "Current provider: openai") {
		t.Fatalf("unexpected provider context: %s", context)
	}
}

func setenv(t *testing.T, key, value string) {
	t.Helper()
	old, had := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}
