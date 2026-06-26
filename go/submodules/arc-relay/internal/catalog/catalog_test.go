package catalog

import "testing"

func TestDeriveSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips prefix", "io.github.getsentry/sentry-mcp", "sentry-mcp"},
		{"simple name", "simple-server", "simple-server"},
		{"already lowercase", "my-tool", "my-tool"},
		{"uppercased", "My-Tool", "my-tool"},
		{"special chars", "my.cool_server!", "my-cool-server"},
		{"nested prefix", "io.github.user/repo/sub-name", "sub-name"},
		{"dots become hyphens", "my.server.name", "my-server-name"},
		{"trailing hyphens trimmed", "test---", "test"},
		{"leading hyphens trimmed", "---test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveSlug(tt.input)
			if got != tt.want {
				t.Errorf("deriveSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDeriveDisplay(t *testing.T) {
	t.Run("prefers title", func(t *testing.T) {
		info := ServerInfo{Title: "My Custom Title"}
		got := deriveDisplay(info, "my-custom-title")
		if got != "My Custom Title" {
			t.Errorf("deriveDisplay() = %q, want %q", got, "My Custom Title")
		}
	})

	t.Run("falls back to slug titlecase", func(t *testing.T) {
		info := ServerInfo{}
		got := deriveDisplay(info, "sentry-mcp")
		want := "Sentry Mcp"
		if got != want {
			t.Errorf("deriveDisplay() = %q, want %q", got, want)
		}
	})

	t.Run("single word slug", func(t *testing.T) {
		info := ServerInfo{}
		got := deriveDisplay(info, "postgres")
		want := "Postgres"
		if got != want {
			t.Errorf("deriveDisplay() = %q, want %q", got, want)
		}
	})
}

func TestRelevanceScore(t *testing.T) {
	tests := []struct {
		name  string
		info  ServerInfo
		query string
		check func(t *testing.T, score int)
	}{
		{
			"exact slug match gets highest score",
			ServerInfo{Name: "io.github.user/sentry"},
			"sentry",
			func(t *testing.T, score int) {
				if score < 100 {
					t.Errorf("exact slug match score = %d, want >= 100", score)
				}
			},
		},
		{
			"common pattern sentry-mcp",
			ServerInfo{Name: "io.github.user/sentry-mcp"},
			"sentry",
			func(t *testing.T, score int) {
				if score < 90 {
					t.Errorf("common pattern score = %d, want >= 90", score)
				}
			},
		},
		{
			"common pattern mcp-github",
			ServerInfo{Name: "io.github.user/mcp-github"},
			"github",
			func(t *testing.T, score int) {
				if score < 90 {
					t.Errorf("common pattern score = %d, want >= 90", score)
				}
			},
		},
		{
			"slug contains query",
			ServerInfo{Name: "io.github.user/sentry-tools-mcp"},
			"sentry",
			func(t *testing.T, score int) {
				if score < 50 {
					t.Errorf("slug contains query score = %d, want >= 50", score)
				}
			},
		},
		{
			"title match",
			ServerInfo{Name: "io.github.user/xyz", Title: "Sentry Integration"},
			"sentry",
			func(t *testing.T, score int) {
				if score < 40 {
					t.Errorf("title match score = %d, want >= 40", score)
				}
			},
		},
		{
			"description match all words",
			ServerInfo{Name: "io.github.user/xyz", Description: "Monitor your sentry alerts"},
			"sentry alerts",
			func(t *testing.T, score int) {
				if score < 30 {
					t.Errorf("description match score = %d, want >= 30", score)
				}
			},
		},
		{
			"no match returns zero",
			ServerInfo{Name: "io.github.user/postgres-mcp", Description: "Postgres tools"},
			"sentry",
			func(t *testing.T, score int) {
				if score != 0 {
					t.Errorf("no match score = %d, want 0", score)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := relevanceScore(tt.info, tt.query)
			tt.check(t, score)
		})
	}

	// Exact match should beat prefix match
	t.Run("exact beats prefix", func(t *testing.T) {
		exact := relevanceScore(ServerInfo{Name: "io.github.user/sentry"}, "sentry")
		prefix := relevanceScore(ServerInfo{Name: "io.github.user/sentry-tools-mcp"}, "sentry")
		if exact <= prefix {
			t.Errorf("exact score (%d) should beat prefix score (%d)", exact, prefix)
		}
	})
}

func TestResolve(t *testing.T) {
	t.Run("remote entries produce remote type", func(t *testing.T) {
		entry := RegistryEntry{
			Server: ServerInfo{
				Name:        "io.github.user/test-server",
				Description: "A test server",
				Remotes: []Remote{
					{Type: "sse", URL: "https://example.com/sse"},
				},
			},
		}
		results := Resolve(entry)
		if len(results) != 1 {
			t.Fatalf("Resolve() returned %d results, want 1", len(results))
		}
		if results[0].ServerType != "remote" {
			t.Errorf("ServerType = %q, want %q", results[0].ServerType, "remote")
		}
		if results[0].RemoteURL != "https://example.com/sse" {
			t.Errorf("RemoteURL = %q, want %q", results[0].RemoteURL, "https://example.com/sse")
		}
		if results[0].RemoteType != "sse" {
			t.Errorf("RemoteType = %q, want %q", results[0].RemoteType, "sse")
		}
		if results[0].SuggestedSlug != "test-server" {
			t.Errorf("SuggestedSlug = %q, want %q", results[0].SuggestedSlug, "test-server")
		}
	})

	t.Run("OCI packages produce stdio type", func(t *testing.T) {
		entry := RegistryEntry{
			Server: ServerInfo{
				Name:        "io.github.user/docker-tool",
				Description: "Docker tool",
				Packages: []Package{
					{
						RegistryType: "oci",
						Identifier:   "ghcr.io/user/tool:latest",
						Transport:    Transport{Type: "stdio"},
					},
				},
			},
		}
		results := Resolve(entry)
		if len(results) != 1 {
			t.Fatalf("Resolve() returned %d results, want 1", len(results))
		}
		if results[0].ServerType != "stdio" {
			t.Errorf("ServerType = %q, want %q", results[0].ServerType, "stdio")
		}
		if results[0].DockerImage != "ghcr.io/user/tool:latest" {
			t.Errorf("DockerImage = %q, want %q", results[0].DockerImage, "ghcr.io/user/tool:latest")
		}
	})

	t.Run("multiple deployment options", func(t *testing.T) {
		entry := RegistryEntry{
			Server: ServerInfo{
				Name: "io.github.user/multi-server",
				Remotes: []Remote{
					{Type: "sse", URL: "https://example.com/sse"},
				},
				Packages: []Package{
					{RegistryType: "oci", Identifier: "ghcr.io/user/server:latest"},
				},
			},
		}
		results := Resolve(entry)
		if len(results) != 2 {
			t.Fatalf("Resolve() returned %d results, want 2", len(results))
		}
		// Remote comes first
		if results[0].ServerType != "remote" {
			t.Errorf("results[0].ServerType = %q, want %q", results[0].ServerType, "remote")
		}
		if results[1].ServerType != "stdio" {
			t.Errorf("results[1].ServerType = %q, want %q", results[1].ServerType, "stdio")
		}
	})

	t.Run("npm fallback when no remotes or OCI", func(t *testing.T) {
		entry := RegistryEntry{
			Server: ServerInfo{
				Name: "io.github.user/npm-tool",
				Packages: []Package{
					{
						RegistryType: "npm",
						Identifier:   "@user/tool",
						EnvironmentVariables: []EnvironmentVariable{
							{Name: "API_KEY", IsRequired: true},
						},
					},
				},
			},
		}
		results := Resolve(entry)
		if len(results) != 1 {
			t.Fatalf("Resolve() returned %d results, want 1", len(results))
		}
		if results[0].ServerType != "stdio" {
			t.Errorf("ServerType = %q, want %q", results[0].ServerType, "stdio")
		}
		if results[0].DockerImage != "" {
			t.Errorf("DockerImage = %q, want empty", results[0].DockerImage)
		}
		if len(results[0].EnvVars) != 1 {
			t.Errorf("EnvVars length = %d, want 1", len(results[0].EnvVars))
		}
	})

	t.Run("empty entry returns no results", func(t *testing.T) {
		entry := RegistryEntry{
			Server: ServerInfo{Name: "io.github.user/empty"},
		}
		results := Resolve(entry)
		if len(results) != 0 {
			t.Errorf("Resolve() returned %d results, want 0", len(results))
		}
	})
}
