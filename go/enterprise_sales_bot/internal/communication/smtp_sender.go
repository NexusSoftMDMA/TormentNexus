package communication

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// EmailSender defines the interface for sending outbound emails.
type EmailSender interface {
	Send(ctx context.Context, to EmailMessage) error
}

// EmailMessage represents an outbound email to send.
type EmailMessage struct {
	To	string	// recipient email address
	Subject	string
	Body	string	// plain text body (HTML can be added later)
	ReplyTo	string	// optional reply-to address
}

// SMTPConfig holds the configuration for SMTP email sending.
type SMTPConfig struct {
	Host		string	// e.g. "smtp.gmail.com"
	Port		int	// e.g. 587 for STARTTLS, 465 for SSL
	Username	string	// email address or username
	Password	string	// app password or SMTP password
	From		string	// sender email address (usually same as Username)
	FromName	string	// sender display name
}

// SMTPSender implements EmailSender using SMTP with STARTTLS.
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender creates a new SMTP email sender.
func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	if cfg.FromName == "" {
		cfg.FromName = "TormentNexus Sales"
	}
	return &SMTPSender{config: cfg}
}

// Send sends an email via SMTP.
func (s *SMTPSender) Send(ctx context.Context, msg EmailMessage) error {
	if msg.To == "" {
		return fmt.Errorf("smtp: recipient address is empty")
	}

	addr := net.JoinHostPort(s.config.Host, fmt.Sprintf("%d", s.config.Port))

	// Build RFC 5322 message
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	headers := map[string]string{
		"From":		from,
		"To":		msg.To,
		"Subject":	msg.Subject,
		"MIME-Version":	"1.0",
		"Content-Type":	"text/plain; charset=\"UTF-8\"",
		"Date":		time.Now().Format(time.RFC1123Z),
	}

	if msg.ReplyTo != "" {
		headers["Reply-To"] = msg.ReplyTo
	}

	var message strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&message, "%s: %s\r\n", k, v)
	}
	message.WriteString("\r\n")
	message.WriteString(msg.Body)

	body := []byte(message.String())

	// TLS config
	tlsConfig := &tls.Config{
		ServerName:	s.config.Host,
		MinVersion:	tls.VersionTLS12,
	}

	// Connect and send
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	if s.config.Port == 465 {
		// Direct SSL connection (port 465)
		return s.sendDirectSSL(addr, tlsConfig, auth, msg.To, body)
	}

	// STARTTLS connection (port 587)
	return s.sendSTARTTLS(addr, tlsConfig, auth, s.config.From, msg.To, body)
}

// sendSTARTTLS connects via plain TCP then upgrades to TLS (port 587).
func (s *SMTPSender) sendSTARTTLS(addr string, tlsConfig *tls.Config, auth smtp.Auth, from, to string, body []byte) error {
	// Use net.DialTimeout for connection with context awareness
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("smtp: connection failed: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("smtp: client creation failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	// STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp: STARTTLS failed: %w", err)
		}
	}

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp: auth failed: %w", err)
		}
	}

	// Set sender and recipient
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp: MAIL FROM failed: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp: RCPT TO failed: %w", err)
	}

	// Send body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: DATA failed: %w", err)
	}

	_, err = writer.Write(body)
	if err != nil {
		return fmt.Errorf("smtp: write body failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("smtp: close writer failed: %w", err)
	}

	slog.Info(fmt.Sprintf("SMTP: Email sent to %s via %s", to, addr))
	return client.Quit()
}

// sendDirectSSL connects via direct TLS (port 465).
func (s *SMTPSender) sendDirectSSL(addr string, tlsConfig *tls.Config, auth smtp.Auth, to string, body []byte) error {
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp: TLS connection failed: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("smtp: client creation failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp: auth failed: %w", err)
		}
	}

	if err = client.Mail(s.config.From); err != nil {
		return fmt.Errorf("smtp: MAIL FROM failed: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp: RCPT TO failed: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: DATA failed: %w", err)
	}

	_, err = writer.Write(body)
	if err != nil {
		return fmt.Errorf("smtp: write body failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("smtp: close writer failed: %w", err)
	}

	slog.Info(fmt.Sprintf("SMTP: Email sent to %s via %s (SSL)", to, addr))
	return client.Quit()
}

// HealthCheck verifies SMTP connectivity by connecting and authenticating.
func (s *SMTPSender) HealthCheck(ctx context.Context) error {
	addr := net.JoinHostPort(s.config.Host, fmt.Sprintf("%d", s.config.Port))

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp health: connection to %s failed: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("smtp health: client creation failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Try STARTTLS if available
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:	s.config.Host,
			MinVersion:	tls.VersionTLS12,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp health: STARTTLS failed: %w", err)
		}
	}

	// Try auth
	if s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp health: auth failed: %w", err)
		}
	}

	return client.Quit()
}

// MockEmailSender is a no-op email sender for testing.
type MockEmailSender struct {
	Sent []EmailMessage
}

func (m *MockEmailSender) Send(ctx context.Context, msg EmailMessage) error {
	m.Sent = append(m.Sent, msg)
	slog.Info(fmt.Sprintf("MockEmailSender: Would send email to %s (subject: %s)", msg.To, msg.Subject))
	return nil
}
