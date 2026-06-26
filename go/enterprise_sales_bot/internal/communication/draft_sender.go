package communication

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// DraftSender implements EmailSender by saving emails as drafts in an IMAP
// mailbox instead of sending them. This lets the user review outreach in their
// Gmail Drafts folder before actually sending.
type DraftSender struct {
	Host		string
	Port		int
	Username	string
	Password	string
	Folder		string	// target mailbox, default "Drafts"
}

// NewDraftSender creates a sender that saves drafts via IMAP.
// For Gmail, uses "[Gmail]/Drafts" as the drafts folder.
func NewDraftSender(host string, port int, username, password string) *DraftSender {
	if port == 0 {
		port = 993
	}
	return &DraftSender{
		Host:		host,
		Port:		port,
		Username:	username,
		Password:	password,
		Folder:		"[Gmail]/Drafts",
	}
}

// Send saves the email as a draft in the IMAP Drafts folder instead of sending it.
func (d *DraftSender) Send(ctx context.Context, msg EmailMessage) error {
	addr := fmt.Sprintf("%s:%d", d.Host, d.Port)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("draft: IMAP connect failed: %w", err)
	}
	defer func() { _ = c.Logout() }()

	if err := c.Login(d.Username, d.Password); err != nil {
		return fmt.Errorf("draft: IMAP login failed: %w", err)
	}

	// Build RFC 5322 message with draft headers
	now := time.Now()
	var message strings.Builder
	fmt.Fprintf(&message, "From: %s\r\n", d.Username)
	fmt.Fprintf(&message, "To: %s\r\n", msg.To)
	fmt.Fprintf(&message, "Subject: %s\r\n", msg.Subject)
	fmt.Fprintf(&message, "Date: %s\r\n", now.Format(time.RFC1123Z))
	fmt.Fprintf(&message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&message, "Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	fmt.Fprintf(&message, "\r\n")
	message.WriteString(msg.Body)

	// Append to Drafts folder with \Draft flag
	body := message.String()
	literal := newLiteral(body)

	if err := c.Append(d.Folder, []string{imap.DraftFlag}, now, literal); err != nil {
		return fmt.Errorf("draft: failed to save to %s: %w", d.Folder, err)
	}

	slog.Info(fmt.Sprintf("DraftSender: Saved draft to %s for %s (subject: %s)", d.Folder, msg.To, msg.Subject))
	return nil
}

// HealthCheck verifies IMAP connectivity.
func (d *DraftSender) HealthCheck(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", d.Host, d.Port)
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("draft health: connect failed: %w", err)
	}
	defer func() { _ = c.Logout() }()

	if err := c.Login(d.Username, d.Password); err != nil {
		return fmt.Errorf("draft health: login failed: %w", err)
	}

	return nil
}

// literalString wraps a string to implement imap.Literal (io.Reader + Len()).
type literalString struct {
	*strings.Reader
	size	int
}

func newLiteral(s string) *literalString {
	return &literalString{
		Reader:	strings.NewReader(s),
		size:	len(s),
	}
}

func (l *literalString) Len() int {
	return l.size
}
