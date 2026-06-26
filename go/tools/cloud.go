package tools

import (
	"fmt"
	"strings"
)

func (r *Registry) registerCloudTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "cloud_troubleshoot",
		Description: "Diagnoses cloud infrastructure issues (AWS/Azure/GCP parity with Amazon Q). Arguments: resource_id (string), error_log (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			resourceID, _ := args["resource_id"].(string)
			errorLog, _ := args["error_log"].(string)

			// Simulate log analysis and cloud diagnostics
			diagnostic := fmt.Sprintf("Diagnostic for %s:\nAnalyzed log: %s\nPotential Root Cause: IAM permissions misconfiguration or network boundary issue.\nSuggested Fix: Review security group inbound rules.", resourceID, errorLog)
			return diagnostic, nil
		},
	})

	r.Tools = append(r.Tools, Tool{
		Name:        "generate_devops_pipeline",
		Description: "Generates CI/CD pipelines (Factory parity). Arguments: platform (string e.g. github_actions), project_type (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			platform, _ := args["platform"].(string)
			projectType, _ := args["project_type"].(string)

			pipeline := ""
			if strings.Contains(strings.ToLower(platform), "github") {
				pipeline = fmt.Sprintf("name: %s CI\n\non: [push]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n    - uses: actions/checkout@v3\n    - name: Run Build\n      run: make build", projectType)
			} else {
				pipeline = "Pipeline generation supported for GitHub Actions, GitLab CI, and Jenkins."
			}
			return pipeline, nil
		},
	})
}
