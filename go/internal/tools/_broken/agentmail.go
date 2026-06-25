package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Email struct for internal use
type emailData struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Date    string `json:"date"`
}

func getEmailDir() (string, error) {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", fmt.Errorf("cannot get home directory: %w", homeErr)
}

	return filepath.Join(homeDir, ".agentmail", "emails"), nil
}

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")

	if to == "" {
		return err("recipient (to) is required")
}

	emailDir, dirErr := getEmailDir()
	if dirErr != nil {
		return err(dirErr.Error())
}

	mkdirErr := os.MkdirAll(emailDir, 0750)
	if mkdirErr != nil {
		return err(mkdirErr.Error())
}

	// Generate a unique identifier
	ts := time.Now().UnixNano()
	safeSubject := strings.ReplaceAll(subject, " ", "_")
	if len(safeSubject) > 20 {
		safeSubject = safeSubject[:20]
	}
	filename := fmt.Sprintf("%d_%s.json", ts, safeSubject)
	filePath := filepath.Join(emailDir, filename)

	email := emailData{
		To:      to,
		Subject: subject,
		Body:    body,
		Date:    time.Now().Format(time.RFC3339),
	}

	jsonData, jsonErr := json.MarshalIndent(email, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	writeErr := os.WriteFile(filePath, jsonData, 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok(fmt.Sprintf("Email saved to %s", filePath))
}

func HandleListEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	emailDir, dirErr := getEmailDir()
	if dirErr != nil {
		return err(dirErr.Error())
}

	// If directory does not exist, return empty list
	if _, statErr := os.Stat(emailDir); os.IsNotExist(statErr) {
		return ok("[]")
}

	files, readErr := os.ReadDir(emailDir)
	if readErr != nil {
		return err(readErr.Error())
}

	type emailSummary struct {
		ID      string `json:"id"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Date    string `json:"date"`
	}

	var summaries []emailSummary
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(emailDir, f.Name())
		data, fileReadErr := os.ReadFile(filePath)
		if fileReadErr != nil {
			continue // skip unreadable
		}
		var email emailData
		if json.Unmarshal(data, &email) != nil {
			continue
		}
		summaries = append(summaries, emailSummary{
			ID:      strings.TrimSuffix(f.Name(), ".json"),
			To:      email.To,
			Subject: email.Subject,
			Date:    email.Date,
		})

	if summaries == nil {
		summaries = []emailSummary{} // ensure it's [] not null
	}

	jsonData, jsonErr := json.MarshalIndent(summaries, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonData))
}

}

func HandleReadEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	emailDir, dirErr := getEmailDir()
	if dirErr != nil {
		return err(dirErr.Error())
}

	filePath := filepath.Join(emailDir, id+".json")
	data, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return err(fmt.Sprintf("email not found: %s", id))
}

	var email emailData
	if parseErr := json.Unmarshal(data, &email); parseErr != nil {
		return err("failed to parse email: " + parseErr.Error())
}

	jsonOut, jsonErr := json.MarshalIndent(email, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonOut))
}

func HandleDeleteEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	emailDir, dirErr := getEmailDir()
	if dirErr != nil {
		return err(dirErr.Error())
}

	filePath := filepath.Join(emailDir, id+".json")
	removeErr := os.Remove(filePath)
	if removeErr != nil {
		if os.IsNotExist(removeErr) {
			return err("email not found: " + id)
}

		return err(removeErr.Error())
}

	return ok(fmt.Sprintf("Email %s deleted", id))
}