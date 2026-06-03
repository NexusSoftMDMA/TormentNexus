package repograph

import (
	"os"
	"path/filepath"
	"strings"
)

func (rgs *RepoGraphService) resolveGoImport(importPath string) string {
	// If it's a relative import (rare in Go, but possible in some projects)
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return "import:" + importPath
	}

	// Dynamically detect the Go module path
	modulePrefix := rgs.detectGoModule()

	// Handle standard library imports (no dot in path = stdlib)
	if !strings.Contains(importPath, ".") && !strings.HasPrefix(importPath, modulePrefix) {
		return "import:std/" + importPath
	}

	// If this is an internal import (matches our module), resolve to file
	if modulePrefix != "" && strings.HasPrefix(importPath, modulePrefix) {
		internalPath := strings.TrimPrefix(importPath, modulePrefix)
		internalPath = strings.TrimPrefix(internalPath, "/")

		if rgs.graph != nil {
			targetPkg := filepath.Base(internalPath)

			for _, node := range rgs.graph.Nodes {
				if node.Type == NodeFile && node.Package == targetPkg {
					if strings.Contains(node.Path, internalPath) {
						return "file:" + node.Path
					}
				}
			}
		}
	}

	// External dependency — categorize by host
	return rgs.categorizeExternalImport(importPath)
}

// detectGoModule reads the go.mod file to determine the module path.
// Falls back to "github.com/tormentnexushq/tormentnexus-go" if go.mod is not found.
func (rgs *RepoGraphService) detectGoModule() string {
	rgs.goModuleMu.Do(func() {
		candidates := []string{
			filepath.Join(rgs.root, "go.mod"),
			filepath.Join(rgs.root, "go", "go.mod"),
		}
		for _, candidate := range candidates {
			data, err := os.ReadFile(candidate)
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					rgs.goModule = strings.TrimSpace(strings.TrimPrefix(line, "module "))
					return
				}
			}
		}
		// Fallback
		rgs.goModule = "github.com/tormentnexushq/tormentnexus-go"
	})
	return rgs.goModule
}

// categorizeExternalImport classifies third-party dependencies by their hosting platform.
func (rgs *RepoGraphService) categorizeExternalImport(importPath string) string {
	prefixes := map[string]string{
		"github.com/":         "import:github/",
		"gitlab.com/":        "import:gitlab/",
		"bitbucket.org/":     "import:bitbucket/",
		"gopkg.in/":          "import:gopkg/",
		"go.uber.org/":       "import:uber/",
		"go.etcd.io/":        "import:etcd/",
		"honnef.co/":         "import:honnef/",
		"cloud.google.com/":  "import:gcp/",
		"google.golang.org/": "import:google/",
		"golang.org/x/":      "import:golang-x/",
		"k8s.io/":            "import:k8s/",
		"sigstore.dev/":      "import:sigstore/",
	}
	for prefix, category := range prefixes {
		if strings.HasPrefix(importPath, prefix) {
			return category + strings.TrimPrefix(importPath, prefix)
		}
	}
	return "import:external/" + importPath
}
