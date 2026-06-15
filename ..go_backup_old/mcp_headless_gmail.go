package tools

import (
	"context"
	"net/smtp"
	"os"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" || body == "" {
		return err("missing required fields: to, subject, body")
}

	from := os.Getenv("GMAIL_FROM")
	password := os.Getenv("GMAIL_PASSWORD")
	if from == "" || password == "" {
		return err("GMAIL_FROM and GMAIL_PASSWORD environment variables required")
}

	msg := []byte("To: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + body)
	e := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, password, "smtp.gmail.com"), from, []string{to}, msg)
	if e != nil {
		return err("failed to send email: " + e.Error())
}

	return success("Email sent successfully")
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}