package db

import "time"

// LeadState represents the current state of a lead in the pipeline.
type LeadState string

const (
	StateDiscovered   LeadState = "Discovered"
	StateResearched   LeadState = "Researched"
	StateOutreachSent LeadState = "Outreach_Sent"
	StateEngaged      LeadState = "Engaged"
	StateNegotiating  LeadState = "Negotiating"
	StatePendingApproval LeadState = "Pending_Approval"  // Awaiting human review for high-value deals
	StateClosedWon    LeadState = "Closed_Won"
	StateClosedLost   LeadState = "Closed_Lost"
)

// Company represents a target organization.
type Company struct {
	ID            int64     `db:"id"`
	Name          string    `db:"name"`
	Domain        string    `db:"domain"`
	TechStack     []string  `db:"tech_stack"`
	HiringSignals []string  `db:"hiring_signals"`
	MarketCapTier string    `db:"market_cap_tier"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Channel represents a communication channel for outreach.
type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelLinkedIn Channel = "linkedin"
	ChannelGitHub   Channel = "github"
)

// DefaultChannel returns the default outreach channel.
func DefaultChannel() Channel {
	return ChannelEmail
}

// IsValid checks if the channel is a recognized channel.
func (c Channel) IsValid() bool {
	switch c {
	case ChannelEmail, ChannelLinkedIn, ChannelGitHub:
		return true
	}
	return false
}

// String returns the string representation of the channel.
func (c Channel) String() string {
	return string(c)
}

// Contact represents an individual decision-maker at a company.
type Contact struct {
	ID               int64     `db:"id"`
	CompanyID        int64     `db:"company_id"`
	Name             string    `db:"name"`
	Role             string    `db:"role"`
	Email            string    `db:"email"`
	GitHubHandle     string    `db:"github_handle"`
	LinkedInURL      string    `db:"linkedin_url"`
	PreferredChannel string    `db:"preferred_channel"` // "email", "linkedin", "github"
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// Interaction tracks communications with a contact.
type Interaction struct {
	ID        int64     `db:"id"`
	ContactID int64     `db:"contact_id"`
	Channel   string    `db:"channel"`   // e.g., Email, LinkedIn, GitHub
	Direction string    `db:"direction"` // e.g., Inbound, Outbound
	RawText   string    `db:"raw_text"`
	Summary   string    `db:"summary"`
	Sentiment string    `db:"sentiment"`
	Success   bool      `db:"success"`   // Indicates if the interaction led to a positive outcome
	TemplateID string   `db:"template_id"` // Optional: template used for outbound interactions
	ResponseID string   `db:"response_id"` // Optional: objection response used for this interaction
	CreatedAt time.Time `db:"created_at"`
}

// Deal tracks the financial and state progress of a lead.
type Deal struct {
	ID                 int64     `db:"id"`
	CompanyID          int64     `db:"company_id"`
	CurrentState       LeadState `db:"current_state"`
	QuotedPricing      float64   `db:"quoted_pricing"`
	CustomRequirements string    `db:"custom_requirements"`
	TechnicalDossier   string    `db:"technical_dossier"`
	CadenceStep        int       `db:"cadence_step"` // 0 = not started, 1+ = current step index
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// Template represents a reusable outreach message template.
type Template struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Subject   string    `db:"subject"`
	Body      string    `db:"body"`
	Channel   string    `db:"channel"`   // email, linkedin, github
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TemplateMetrics tracks usage and success of outreach templates.
type TemplateMetrics struct {
	TemplateID  string    `db:"template_id"`
	Impressions int       `db:"impressions"`
	Successes   int       `db:"successes"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// PerformanceMetrics aggregates key sales pipeline statistics.
type PerformanceMetrics struct {
	TotalLeads         int            `json:"total_leads"`
	LeadsByState       map[LeadState]int `json:"leads_by_state"`
	SuccessfulOutreach int            `json:"successful_outreach"`
	WinRate            float64        `json:"win_rate"`
}

// SocialPost represents a logged social media post.
type SocialPost struct {
	ID              int64     `db:"id"`
	Brand           string    `db:"brand"`            // "tormentnexus" or "hypernexus"
	Platform        string    `db:"platform"`         // "reddit", "bluesky", "linkedin", "twitter"
	AccountUsername string    `db:"account_username"`
	PostContent     string    `db:"post_content"`
	Status          string    `db:"status"`           // "posted", "failed", "draft"
	CreatedAt       time.Time `db:"created_at"`
}
