package catalogvalidator

import (
	"fmt"
	"net/url"
	"strings"
)

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) ValidateEntry(title, urlStr, description string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if strings.TrimSpace(title) == "" {
		result.Errors = append(result.Errors, "title is required")
	}

	if strings.TrimSpace(urlStr) == "" {
		result.Errors = append(result.Errors, "url is required")
	} else {
		parsed, err := url.Parse(urlStr)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid URL: %s", urlStr))
		}
	}

	if len(title) > 500 {
		result.Warnings = append(result.Warnings, "title exceeds 500 characters")
	}

	if len(description) > 5000 {
		result.Warnings = append(result.Warnings, "description exceeds 5000 characters")
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}
	return result
}

func (s *Service) ValidateBatch(entries []struct{ Title, URL, Description string }) []*ValidationResult {
	results := make([]*ValidationResult, len(entries))
	for i, e := range entries {
		results[i] = s.ValidateEntry(e.Title, e.URL, e.Description)
	}
	return results
}

func (s *Service) ValidateConfig(config map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}
	for k, v := range config {
		if v == nil || v == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("config key '%s' is empty", k))
		}
	}
	return result
}
