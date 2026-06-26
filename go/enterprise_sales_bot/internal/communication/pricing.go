package communication

import "strings"

// CalculateQuote determines a base annual price based on company tier.
func CalculateQuote(tier string) int {
	switch strings.ToLower(tier) {
	case "enterprise":
		return 50000
	case "mid-market":
		return 15000
	case "smb":
		return 5000
	default:
		return 10000
	}
}
