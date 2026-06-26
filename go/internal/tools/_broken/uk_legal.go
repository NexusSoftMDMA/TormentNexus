package tools

import (
	"context"
	"fmt"
	"regexp"
)

func HandleGetLegalInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	legalID, _ :=getString(args, "legal_id")
	if legalID == "" {
		return err("legal_id is required")
}

	// Simulate fetching legal information
	legalInfo := fmt.Sprintf("Legal information for ID: %s", legalID)
	return ok(legalInfo)
}

func HandleSearchLegalTerms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	term, _ :=getString(args, "term")
	if term == "" {
		return err("term is required")
}

	// Simulate searching legal terms
	searchResult := fmt.Sprintf("Search results for term: %s", term)
	return ok(searchResult)
}

func HandleGetLegalDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	documentID, _ :=getString(args, "document_id")
	if documentID == "" {
		return err("document_id is required")
}

	// Simulate fetching a legal document
	documentContent := fmt.Sprintf("Content of legal document ID: %s", documentID)
	return ok(documentContent)
}

func HandleValidateLegalID(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	legalID, _ :=getString(args, "legal_id")
	if legalID == "" {
		return err("legal_id is required")
}

	// Simple validation using regex
	validID := regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(legalID)
	if !validID {
		return err("invalid legal_id format")
}

	return ok("legal_id is valid")
}