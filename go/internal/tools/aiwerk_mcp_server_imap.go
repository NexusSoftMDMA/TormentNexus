//go:build ignore
// +build ignore

package tools

import (
	"context"
	"net/smtp"
)

func HandleSendEmail(ctx context.Context, args map[string]string) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	port, _ :=getString(args, "port")
	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if server == "" || port == "" || from == "" || to == "" {
		return err("Missing required parameters: server, port, from, to")
}

	auth := smtp.PlainAuth("", username, password, server)
	msg := []byte("To: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + body)
	addr := server + ":" + port
	e := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if e != nil {
		return err("Failed to send email: " + e.Error())
}

	return success("Email sent successfully")
}

func HandleListMessages(ctx context.Context, args map[string]string) (ToolResponse, error) {
	_ = args // not used in stub
	return ok("[{\"id\":1,\"subject\":\"Test\",\"from\":\"user@example.com\"}]")
}
