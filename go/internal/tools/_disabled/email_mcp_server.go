package tools

import (
	"context"
	"net/smtp"
	"os"
	"strings"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" || subject == "" || body == "" {
		return err("Missing required fields: to, subject, body")
}

	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	if from == "" || password == "" || smtpHost == "" || smtpPort == "" {
		return err("Missing SMTP environment variables")
}

	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")
	auth := smtp.PlainAuth("", from, password, smtpHost)
	e := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, strings.Split(to, ","), msg)
	if e != nil {
		return err("Failed to send email: " + e.Error())
}

	return ok("Email sent to " + to)
}