package adapters

import "testing"

func TestSelectProviderRouteHonorsPreferences(t *testing.T) {
	setenv(t, "SUPERCLI_PROVIDER", "openai")
	setenv(t, "SUPERCLI_MODEL", "gpt-4o")
	setenv(t, "OPENAI_API_KEY", "test-key")
	setenv(t, "GOOGLE_API_KEY", "google-key")

	route := SelectProviderRoute(ProviderRouteRequest{CostPreference: "budget", TaskType: "analysis"})
	if route.Provider == "" || route.Model == "" {
		t.Fatalf("unexpected route: %#v", route)
	}
	if len(route.Reasons) == 0 {
		t.Fatalf("expected route reasons: %#v", route)
	}
}
