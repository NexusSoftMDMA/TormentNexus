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

// IMAPConfig holds the configuration for IMAP email polling.
type IMAPConfig struct {
	Host		string	// e.g. "imap.gmail.com"
	Port		int	// e.g. 993 for SSL
	Username	string	// email address
	Password	string	// app password
	Folder		string	// mailbox to poll (default: "INBOX")
}

// InboundEmail represents a parsed inbound email from IMAP.
type InboundEmail struct {
	From	string	// sender email address
	Subject	string
	Body	string	// plain text body
	Date	time.Time
	UID	uint32	// IMAP UID for tracking
}

// EmailReceiver polls an IMAP inbox for new inbound emails.
type EmailReceiver struct {
	config		IMAPConfig
	manager		*Manager
	lastUID		uint32	// tracks the last processed message UID
	connected	bool
}

// NewEmailReceiver creates a new IMAP email receiver.
func NewEmailReceiver(cfg IMAPConfig, manager *Manager) *EmailReceiver {
	if cfg.Port == 0 {
		cfg.Port = 993
	}
	if cfg.Folder == "" {
		cfg.Folder = "INBOX"
	}
	return &EmailReceiver{
		config:		cfg,
		manager:	manager,
	}
}

// Run starts the IMAP polling loop. It checks for new emails at the specified interval.
func (r *EmailReceiver) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("IMAP: Email receiver started (polling %s every %v)", r.config.Host, interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("IMAP: Email receiver stopping...")
			return
		case <-ticker.C:
			if err := r.pollNewEmails(ctx); err != nil {
				slog.Info(fmt.Sprintf("IMAP: Poll error: %v", err))
			}
		}
	}
}

// pollNewEmails connects to IMAP, fetches unread messages, and processes them.
func (r *EmailReceiver) pollNewEmails(ctx context.Context) error {
	c, err := r.connect()
	if err != nil {
		return fmt.Errorf("imap: connect failed: %w", err)
	}
	defer r.disconnect(c)

	// Select the mailbox
	mbox, err := c.Select(r.config.Folder, true)	// read-only
	if err != nil {
		return fmt.Errorf("imap: select %s failed: %w", r.config.Folder, err)
	}

	if mbox.Messages == 0 {
		return nil	// empty inbox
	}

	// Fetch messages since last processed UID
	// If we haven't processed any yet, fetch the last 10 unread
	var fromUID uint32
	if r.lastUID > 0 {
		fromUID = r.lastUID + 1
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	if fromUID > 0 {
		criteria.Uid = new(imap.SeqSet)
		criteria.Uid.AddRange(fromUID, 0)	// 0 means max
	}

	uids, err := c.UidSearch(criteria)
	if err != nil {
		return fmt.Errorf("imap: search failed: %w", err)
	}

	if len(uids) == 0 {
		return nil	// no new messages
	}

	// Limit to last 10 messages to avoid overwhelming the pipeline
	if len(uids) > 10 {
		uids = uids[len(uids)-10:]
	}

	slog.Info(fmt.Sprintf("IMAP: Found %d new unread messages", len(uids)))

	// Fetch the messages
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody, imap.FetchUid}
	messages := make(chan *imap.Message, len(uids))

	go func() {
		if err := c.UidFetch(seqSet, items, messages); err != nil {
			slog.Info(fmt.Sprintf("IMAP: Fetch error: %v", err))
		}
	}()

	var maxUID uint32
	for msg := range messages {
		if msg == nil {
			continue
		}

		email := r.parseMessage(msg)
		if email == nil {
			continue
		}

		// Track the highest UID we've seen
		if msg.Uid > maxUID {
			maxUID = msg.Uid
		}

		// Process the inbound email through the sales pipeline
		r.processInboundEmail(ctx, *email)
	}

	// Update lastUID for next poll
	if maxUID > r.lastUID {
		r.lastUID = maxUID
	}

	return nil
}

// parseMessage extracts an InboundEmail from an IMAP message.
func (r *EmailReceiver) parseMessage(msg *imap.Message) *InboundEmail {
	if msg.Envelope == nil {
		return nil
	}

	from := ""
	if len(msg.Envelope.From) > 0 {
		addr := msg.Envelope.From[0]
		from = fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName)
	}

	if from == "" {
		return nil	// can't process without sender
	}

	body := ""
	if msg.Body != nil {
		for _, literal := range msg.Body {
			buf := make([]byte, 32*1024)	// 32KB max
			n, err := literal.Read(buf)
			if err == nil && n > 0 {
				body = string(buf[:n])
				break
			}
		}
	}

	return &InboundEmail{
		From:		strings.ToLower(strings.TrimSpace(from)),
		Subject:	msg.Envelope.Subject,
		Body:		body,
		Date:		msg.Envelope.Date,
		UID:		msg.Uid,
	}
}

// processInboundEmail feeds a received email into the sales pipeline.
func (r *EmailReceiver) processInboundEmail(ctx context.Context, email InboundEmail) {
	slog.Info(fmt.Sprintf("IMAP: Received email from %s (subject: %s)", email.From, email.Subject))

	// Look up the contact by email address
	contact, err := r.manager.db.GetContactByEmail(ctx, email.From)
	if err != nil || contact == nil {
		slog.Info(fmt.Sprintf("IMAP: No contact found for %s — skipping (not in pipeline)", email.From))
		return
	}

	// Feed into the communication manager's inbound pipeline
	text := email.Subject
	if email.Body != "" {
		text = fmt.Sprintf("Subject: %s\n\n%s", email.Subject, email.Body)
	}

	reply, err := r.manager.ProcessInbound(ctx, *contact, text)
	if err != nil {
		slog.Info(fmt.Sprintf("IMAP: Error processing inbound from %s: %v", email.From, err))
		return
	}

	if reply != "" {
		slog.Info(fmt.Sprintf("IMAP: Generated reply to %s: %s", email.From, truncate(reply, 100)))
	}
}

// connect establishes an IMAP connection.
func (r *EmailReceiver) connect() (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("imap: TLS dial to %s failed: %w", addr, err)
	}

	if err := c.Login(r.config.Username, r.config.Password); err != nil {
		_ = c.Logout()
		return nil, fmt.Errorf("imap: login failed: %w", err)
	}

	r.connected = true
	return c, nil
}

// disconnect closes the IMAP connection gracefully.
func (r *EmailReceiver) disconnect(c *client.Client) {
	if c != nil {
		_ = c.Logout()
		r.connected = false
	}
}

// HealthCheck verifies IMAP connectivity.
func (r *EmailReceiver) HealthCheck(ctx context.Context) error {
	c, err := r.connect()
	if err != nil {
		return err
	}
	defer r.disconnect(c)
	return nil
}

// MockEmailReceiver is a no-op email receiver for testing.
type MockEmailReceiver struct{}

func (m *MockEmailReceiver) Run(ctx context.Context, interval time.Duration) {
	slog.Info("MockEmailReceiver: Running (no-op)")
	<-ctx.Done()
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
