package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
)

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "password")
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")

	if host == "" || user == "" || pass == "" || from == "" || to == "" {
		return err("missing required email configuration fields")
}

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", user, pass, host)
	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body)

	e := smtp.SendMail(addr, auth, from, strings.Split(to, ","), msg)
	if e != nil {
		return err("failed to send email: " + e.Error())
}

	return success("email sent successfully")
}

func HandleListConfigs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configs := []map[string]string{
		{"name": "gmail", "host": "smtp.gmail.com", "port": "587"},
		{"name": "outlook", "host": "smtp.office365.com", "port": "587"},
		{"name": "yahoo", "host": "smtp.mail.yahoo.com", "port": "465"},
	}
	data, e := json.Marshal(configs)
	if e != nil {
		return err("failed to marshal configs")
}

	return ok(string(data))
}