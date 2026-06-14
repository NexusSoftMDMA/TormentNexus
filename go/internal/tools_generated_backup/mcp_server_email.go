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
	if len(to) == 0 || len(subject) == 0 || len(body) == 0 {
		return err("to, subject, and body are required")
}

	from := os.Getenv("SMTP_FROM")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	if from == "" || host == "" || port == "" {
		return err("SMTP environment variables are not set")
}

	msg := "From: " + from + "\nTo: " + to + "\nSubject: " + subject + "\n\n" + body
	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)
	e := smtp.SendMail(addr, auth, from, strings.Split(to, ","), []byte(msg))
	if e != nil {
		return err("failed to send email: " + e.Error())
}

	return ok("email sent to " + to)
}

func HandleGetEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return err("not implemented")
}