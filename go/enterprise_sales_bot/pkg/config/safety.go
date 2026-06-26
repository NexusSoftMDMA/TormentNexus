package config

// SafetyConfig defines the guardrails for autonomous bot actions.
type SafetyConfig struct {
	MaxDailyPRs int
	ToneConstraint string
	OptOutDisclaimer string
}

// DefaultSafetyConfig returns the recommended safety parameters.
func DefaultSafetyConfig() SafetyConfig {
	return SafetyConfig{
		MaxDailyPRs: 5,
		ToneConstraint: "Helpful Peer (Developer-to-Developer)",
		OptOutDisclaimer: "Automated optimization discovery. Reply 'opt-out' to blacklist.",
	}
}
