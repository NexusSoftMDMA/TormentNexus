package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// CreateCompany inserts a new company into the database.
func (db *DB) CreateCompany(ctx context.Context, company *Company) error {
	query := `
		INSERT INTO companies (name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	now := time.Now()
	company.CreatedAt = now
	company.UpdatedAt = now

	err := db.Conn.QueryRowContext(ctx, query,
		company.Name,
		company.Domain,
		pq.Array(company.TechStack),
		pq.Array(company.HiringSignals),
		company.MarketCapTier,
		company.CreatedAt,
		company.UpdatedAt,
	).Scan(&company.ID)

	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}

// UpdateDealDetails updates the pricing and custom requirements of an existing deal.
func (db *DB) UpdateDealDetails(ctx context.Context, dealID int64, pricing float64, requirements string) error {
	query := `
		UPDATE deals
		SET quoted_pricing = $1, custom_requirements = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := db.Conn.ExecContext(ctx, query, pricing, requirements, time.Now(), dealID)
	if err != nil {
		return fmt.Errorf("failed to update deal details: %w", err)
	}
	return nil
}

// GetCompanyByID retrieves a company by its ID.
func (db *DB) GetCompanyByID(ctx context.Context, id int64) (*Company, error) {
	query := `
		SELECT id, name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at
		FROM companies
		WHERE id = $1
	`
	company := &Company{}
	err := db.Conn.QueryRowContext(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.Domain,
		pq.Array(&company.TechStack),
		pq.Array(&company.HiringSignals),
		&company.MarketCapTier,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company by id: %w", err)
	}
	return company, nil
}

// GetCompanyByDomain retrieves a company by its domain.
func (db *DB) GetCompanyByDomain(ctx context.Context, domain string) (*Company, error) {
	query := `
		SELECT id, name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at
		FROM companies
		WHERE domain = $1
	`
	company := &Company{}
	err := db.Conn.QueryRowContext(ctx, query, domain).Scan(
		&company.ID,
		&company.Name,
		&company.Domain,
		pq.Array(&company.TechStack),
		pq.Array(&company.HiringSignals),
		&company.MarketCapTier,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company by domain: %w", err)
	}
	return company, nil
}

// CreateDeal inserts a new deal for a company.
func (db *DB) CreateDeal(ctx context.Context, deal *Deal) error {
	query := `
		INSERT INTO deals (company_id, current_state, quoted_pricing, custom_requirements, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	now := time.Now()
	deal.CreatedAt = now
	deal.UpdatedAt = now

	err := db.Conn.QueryRowContext(ctx, query,
		deal.CompanyID,
		deal.CurrentState,
		deal.QuotedPricing,
		deal.CustomRequirements,
		deal.CreatedAt,
		deal.UpdatedAt,
	).Scan(&deal.ID)

	if err != nil {
		return fmt.Errorf("failed to create deal: %w", err)
	}
	return nil
}

// UpdateDealState updates the state of an existing deal.
func (db *DB) UpdateDealState(ctx context.Context, dealID int64, newState LeadState) error {
	query := `
		UPDATE deals
		SET current_state = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := db.Conn.ExecContext(ctx, query, newState, time.Now(), dealID)
	if err != nil {
		return fmt.Errorf("failed to update deal state: %w", err)
	}
	return nil
}

// ListDealsByState retrieves deals in a specific state.
// ListAllCompanies retrieves all companies in the database.
func (db *DB) ListAllCompanies(ctx context.Context) ([]Company, error) {
	rows, err := db.Conn.QueryContext(ctx, "SELECT id, name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at FROM companies")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.Name, &c.Domain, pq.Array(&c.TechStack), pq.Array(&c.HiringSignals), &c.MarketCapTier, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, nil
}

func (db *DB) ListDealsByState(ctx context.Context, state LeadState) ([]Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, technical_dossier, created_at, updated_at
		FROM deals
		WHERE current_state = $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, state)
	if err != nil {
		return nil, fmt.Errorf("failed to list deals by state: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var deals []Deal
	for rows.Next() {
		var deal Deal
		var pricing sql.NullFloat64
		var requirements, dossier sql.NullString
		if err := rows.Scan(
			&deal.ID,
			&deal.CompanyID,
			&deal.CurrentState,
			&pricing,
			&requirements,
			&dossier,
			&deal.CreatedAt,
			&deal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal: %w", err)
		}
		deal.QuotedPricing = pricing.Float64
		deal.CustomRequirements = requirements.String
		deal.TechnicalDossier = dossier.String
		deals = append(deals, deal)
	}
	return deals, nil
}

// GetDealByCompanyID retrieves the most recent deal for a specific company.
func (db *DB) GetDealByCompanyID(ctx context.Context, companyID int64) (*Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, technical_dossier, created_at, updated_at
		FROM deals
		WHERE company_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`
	deal := &Deal{}
	var pricing sql.NullFloat64
	var requirements, dossier sql.NullString
	err := db.Conn.QueryRowContext(ctx, query, companyID).Scan(
		&deal.ID,
		&deal.CompanyID,
		&deal.CurrentState,
		&pricing,
		&requirements,
		&dossier,
		&deal.CreatedAt,
		&deal.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal by company id: %w", err)
	}
	deal.QuotedPricing = pricing.Float64
	deal.CustomRequirements = requirements.String
	deal.TechnicalDossier = dossier.String
	return deal, nil
}

// ListRecentDeals retrieves the most recently updated deals.
func (db *DB) ListRecentDeals(ctx context.Context, limit int) ([]Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, technical_dossier, created_at, updated_at
		FROM deals
		ORDER BY updated_at DESC
		LIMIT $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent deals: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var deals []Deal
	for rows.Next() {
		var deal Deal
		var pricing sql.NullFloat64
		var requirements, dossier sql.NullString
		if err := rows.Scan(
			&deal.ID,
			&deal.CompanyID,
			&deal.CurrentState,
			&pricing,
			&requirements,
			&dossier,
			&deal.CreatedAt,
			&deal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal: %w", err)
		}
		deal.QuotedPricing = pricing.Float64
		deal.CustomRequirements = requirements.String
		deal.TechnicalDossier = dossier.String
		deals = append(deals, deal)
	}
	return deals, nil
}

// UpdateTechnicalDossier updates the technical dossier for a deal.
func (db *DB) UpdateTechnicalDossier(ctx context.Context, dealID int64, dossier string) error {
	query := `
		UPDATE deals
		SET technical_dossier = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := db.Conn.ExecContext(ctx, query, dossier, time.Now(), dealID)
	if err != nil {
		return fmt.Errorf("failed to update technical dossier: %w", err)
	}
	return nil
}

// CreateContact inserts a new contact into the database.
func (db *DB) CreateContact(ctx context.Context, contact *Contact) error {
	if contact.PreferredChannel == "" {
		contact.PreferredChannel = string(DefaultChannel())
	}

	query := `
		INSERT INTO contacts (company_id, name, role, email, github_handle, linkedin_url, preferred_channel, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	now := time.Now()
	contact.CreatedAt = now
	contact.UpdatedAt = now

	err := db.Conn.QueryRowContext(ctx, query,
		contact.CompanyID,
		contact.Name,
		contact.Role,
		contact.Email,
		contact.GitHubHandle,
		contact.LinkedInURL,
		contact.PreferredChannel,
		contact.CreatedAt,
		contact.UpdatedAt,
	).Scan(&contact.ID)

	if err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}
	return nil
}

// UpdateContactPreferredChannel updates the preferred communication channel for a contact.
func (db *DB) UpdateContactPreferredChannel(ctx context.Context, contactID int64, channel string) error {
	query := `
		UPDATE contacts
		SET preferred_channel = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := db.Conn.ExecContext(ctx, query, channel, time.Now(), contactID)
	if err != nil {
		return fmt.Errorf("failed to update contact preferred channel: %w", err)
	}
	return nil
}

// ListContactsByCompany retrieves all contacts for a specific company.
func (db *DB) ListContactsByCompany(ctx context.Context, companyID int64) ([]Contact, error) {
	query := `
		SELECT id, company_id, name, role, email, github_handle, linkedin_url, preferred_channel, created_at, updated_at
		FROM contacts
		WHERE company_id = $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		if err := rows.Scan(
			&contact.ID,
			&contact.CompanyID,
			&contact.Name,
			&contact.Role,
			&contact.Email,
			&contact.GitHubHandle,
			&contact.LinkedInURL,
			&contact.PreferredChannel,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}


// GetContactByEmail retrieves a contact by their email address.
func (db *DB) GetContactByEmail(ctx context.Context, email string) (*Contact, error) {
	query := `
		SELECT id, company_id, name, role, email, github_handle, linkedin_url, preferred_channel, created_at, updated_at
		FROM contacts
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
	`
	contact := &Contact{}
	err := db.Conn.QueryRowContext(ctx, query, email).Scan(
		&contact.ID,
		&contact.CompanyID,
		&contact.Name,
		&contact.Role,
		&contact.Email,
		&contact.GitHubHandle,
		&contact.LinkedInURL,
		&contact.PreferredChannel,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get contact by email: %w", err)
	}
	return contact, nil
}
// CreateInteraction inserts a new interaction into the database.
func (db *DB) CreateInteraction(ctx context.Context, interaction *Interaction) error {
	query := `
		INSERT INTO interactions (contact_id, channel, direction, raw_text, summary, sentiment, success, template_id, response_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	if interaction.CreatedAt.IsZero() {
		interaction.CreatedAt = time.Now()
	}

	err := db.Conn.QueryRowContext(ctx, query,
		interaction.ContactID,
		interaction.Channel,
		interaction.Direction,
		interaction.RawText,
		interaction.Summary,
		interaction.Sentiment,
		interaction.Success,
		interaction.TemplateID,
		interaction.ResponseID,
		interaction.CreatedAt,
	).Scan(&interaction.ID)

	if err != nil {
		return fmt.Errorf("failed to create interaction: %w", err)
	}
	return nil
}

// UpdateInteractionSuccess updates the success status of an existing interaction.
func (db *DB) UpdateInteractionSuccess(ctx context.Context, interactionID int64, success bool) error {
	query := `
		UPDATE interactions
		SET success = $1
		WHERE id = $2
	`
	_, err := db.Conn.ExecContext(ctx, query, success, interactionID)
	if err != nil {
		return fmt.Errorf("failed to update interaction success: %w", err)
	}
	return nil
}

// ListSuccessfulInteractions retrieves recent interactions marked as successful.
func (db *DB) ListSuccessfulInteractions(ctx context.Context, limit int) ([]Interaction, error) {
	query := `
		SELECT id, contact_id, channel, direction, raw_text, summary, sentiment, success, created_at
		FROM interactions
		WHERE success = true
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list successful interactions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var interactions []Interaction
	for rows.Next() {
		var interaction Interaction
		if err := rows.Scan(
			&interaction.ID,
			&interaction.ContactID,
			&interaction.Channel,
			&interaction.Direction,
			&interaction.RawText,
			&interaction.Summary,
			&interaction.Sentiment,
			&interaction.Success,
			&interaction.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan interaction: %w", err)
		}
		interactions = append(interactions, interaction)
	}
	return interactions, nil
}

// GetPerformanceMetrics aggregates and returns current pipeline performance data.
func (db *DB) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{
		LeadsByState: make(map[LeadState]int),
	}

	// 1. Get counts by state
	query := `SELECT current_state, COUNT(*) FROM deals GROUP BY current_state`
	rows, err := db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query state counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	total := 0
	won := 0
	lost := 0
	for rows.Next() {
		var state LeadState
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("failed to scan state count: %w", err)
		}
		metrics.LeadsByState[state] = count
		total += count
		if state == StateClosedWon {
			won = count
		}
		if state == StateClosedLost {
			lost = count
		}
	}
	metrics.TotalLeads = total

	// 2. Calculate Win Rate
	if total > 0 && (won+lost) > 0 {
		metrics.WinRate = float64(won) / float64(won+lost) * 100
	}

	// 3. Get successful outreach count
	outreachQuery := `SELECT COUNT(*) FROM interactions WHERE success = true`
	err = db.Conn.QueryRowContext(ctx, outreachQuery).Scan(&metrics.SuccessfulOutreach)
	if err != nil {
		return nil, fmt.Errorf("failed to query successful outreach: %w", err)
	}

	return metrics, nil
}

// CreatePullRequest persists a new pull request record.
func (db *DB) CreatePullRequest(ctx context.Context, pr *gitcheck.PullRequest, taskDesc string) error {
	query := `
		INSERT INTO pull_requests (id, branch, title, status, task_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := db.Conn.ExecContext(ctx, query, pr.ID, pr.Branch, pr.Title, pr.Status, taskDesc)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	return nil
}

// UpdatePRStatus updates the status of an existing PR.
func (db *DB) UpdatePRStatus(ctx context.Context, prID string, status gitcheck.PRStatus) error {
	query := `
		UPDATE pull_requests
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := db.Conn.ExecContext(ctx, query, status, prID)
	if err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}
	return nil
}

// ListActivePullRequests retrieves all open pull requests.
func (db *DB) ListActivePullRequests(ctx context.Context) ([]gitcheck.PullRequest, error) {
	query := `
		SELECT id, branch, title, status
		FROM pull_requests
		WHERE status = $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, gitcheck.PRStatusOpen)
	if err != nil {
		return nil, fmt.Errorf("failed to list active PRs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var prs []gitcheck.PullRequest
	for rows.Next() {
		var pr gitcheck.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Branch, &pr.Title, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// ListInteractionsByContact retrieves all interactions for a specific contact.
func (db *DB) ListInteractionsByContact(ctx context.Context, contactID int64) ([]Interaction, error) {
	query := `
		SELECT id, contact_id, channel, direction, raw_text, summary, sentiment, success, template_id, response_id, created_at
		FROM interactions
		WHERE contact_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.Conn.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to list interactions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var interactions []Interaction
	for rows.Next() {
		var interaction Interaction
		if err := rows.Scan(
			&interaction.ID,
			&interaction.ContactID,
			&interaction.Channel,
			&interaction.Direction,
			&interaction.RawText,
			&interaction.Summary,
			&interaction.Sentiment,
			&interaction.Success,
			&interaction.TemplateID,
			&interaction.ResponseID,
			&interaction.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan interaction: %w", err)
		}
		interactions = append(interactions, interaction)
	}
	return interactions, nil
}

// GetCadenceStep retrieves the current cadence step for a deal.
func (db *DB) GetCadenceStep(ctx context.Context, dealID int64) (int, error) {
	query := `SELECT cadence_step FROM deals WHERE id = $1`
	var step int
	err := db.Conn.QueryRowContext(ctx, query, dealID).Scan(&step)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Default to 0 if deal not found
		}
		return 0, fmt.Errorf("failed to get cadence step: %w", err)
	}
	return step, nil
}

// SetCadenceStep updates the cadence step for a deal.
func (db *DB) SetCadenceStep(ctx context.Context, dealID int64, step int) error {
	query := `UPDATE deals SET cadence_step = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.Conn.ExecContext(ctx, query, step, dealID)
	if err != nil {
		return fmt.Errorf("failed to set cadence step: %w", err)
	}
	return nil
}

// GetTemplate retrieves a template by ID.
func (db *DB) GetTemplate(ctx context.Context, id string) (*Template, error) {
	query := `SELECT id, name, subject, body, channel, created_at, updated_at FROM templates WHERE id = $1`
	tmpl := &Template{}
	err := db.Conn.QueryRowContext(ctx, query, id).Scan(
		&tmpl.ID, &tmpl.Name, &tmpl.Subject, &tmpl.Body, &tmpl.Channel, &tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template %s not found", id)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return tmpl, nil
}

// ListTemplates retrieves all templates.
func (db *DB) ListTemplates(ctx context.Context) ([]Template, error) {
	query := `SELECT id, name, subject, body, channel, created_at, updated_at FROM templates ORDER BY id`
	rows, err := db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var templates []Template
	for rows.Next() {
		var tmpl Template
		if err := rows.Scan(&tmpl.ID, &tmpl.Name, &tmpl.Subject, &tmpl.Body, &tmpl.Channel, &tmpl.CreatedAt, &tmpl.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// RecordTemplateImpression increments the impression counter for a template.
func (db *DB) RecordTemplateImpression(ctx context.Context, templateID string) error {
	query := `INSERT INTO template_metrics (template_id, impressions, successes, updated_at)
		VALUES ($1, 1, 0, NOW())
		ON CONFLICT (template_id) DO UPDATE SET impressions = template_metrics.impressions + 1, updated_at = NOW()`
	_, err := db.Conn.ExecContext(ctx, query, templateID)
	return err
}

// RecordTemplateSuccess increments the success counter for a template.
func (db *DB) RecordTemplateSuccess(ctx context.Context, templateID string) error {
	query := `INSERT INTO template_metrics (template_id, impressions, successes, updated_at)
		VALUES ($1, 0, 1, NOW())
		ON CONFLICT (template_id) DO UPDATE SET successes = template_metrics.successes + 1, updated_at = NOW()`
	_, err := db.Conn.ExecContext(ctx, query, templateID)
	return err
}

// GetTemplateMetrics returns all template metrics.
func (db *DB) GetTemplateMetrics(ctx context.Context) ([]TemplateMetrics, error) {
	query := `SELECT template_id, impressions, successes, updated_at FROM template_metrics`
	rows, err := db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query template metrics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var metrics []TemplateMetrics
	for rows.Next() {
		var m TemplateMetrics
		if err := rows.Scan(&m.TemplateID, &m.Impressions, &m.Successes, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan template metric: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// GetTopTemplate returns the template with the highest conversion rate (successes/impressions).
func (db *DB) GetTopTemplate(ctx context.Context) (*Template, error) {
	metrics, err := db.GetTemplateMetrics(ctx)
	if err != nil {
		return nil, err
	}
	var bestID string
	bestScore := -1.0
	for _, m := range metrics {
		if m.Impressions == 0 {
			continue
		}
		score := float64(m.Successes) / float64(m.Impressions)
		if score > bestScore {
			bestScore = score
			bestID = m.TemplateID
		}
	}
	if bestID == "" {
		return nil, fmt.Errorf("no template with impressions found")
	}
	return db.GetTemplate(ctx, bestID)
}

// MarkTemplateSuccessForDeal marks all outbound interactions with templates for a given deal as successful.
// It updates the interaction's success flag and increments the template success counter.
func (db *DB) MarkTemplateSuccessForDeal(ctx context.Context, dealID int64) error {
	// Retrieve the company_id for the deal
	var companyID int64
	queryDeal := `SELECT company_id FROM deals WHERE id = $1`
	if err := db.Conn.QueryRowContext(ctx, queryDeal, dealID).Scan(&companyID); err != nil {
		return fmt.Errorf("failed to get company_id for deal %d: %w", dealID, err)
	}

	// List contacts for the company
	contacts, err := db.ListContactsByCompany(ctx, companyID)
	if err != nil {
		return fmt.Errorf("failed to list contacts for company %d: %w", companyID, err)
	}

	for _, contact := range contacts {
		// Find outbound interactions with a template that haven't been marked successful yet
		query := `SELECT id, template_id FROM interactions WHERE contact_id = $1 AND direction = 'Outbound' AND success = false AND template_id IS NOT NULL`
		rows, err := db.Conn.QueryContext(ctx, query, contact.ID)
		if err != nil {
			return fmt.Errorf("failed to query interactions for contact %d: %w", contact.ID, err)
		}
		for rows.Next() {
			var interactionID int64
			var tmplID string
			if err := rows.Scan(&interactionID, &tmplID); err != nil {
				_ = rows.Close()
				return fmt.Errorf("failed to scan interaction row: %w", err)
			}
			// Mark interaction as successful
			if err := db.UpdateInteractionSuccess(ctx, interactionID, true); err != nil {
				_ = rows.Close()
				return fmt.Errorf("failed to update interaction success for id %d: %w", interactionID, err)
			}
			// Record template success metric
			if err := db.RecordTemplateSuccess(ctx, tmplID); err != nil {
				_ = rows.Close()
				return fmt.Errorf("failed to record template success for %s: %w", tmplID, err)
			}
		}
		_ = rows.Close()
	}
	return nil
}

// CreateSocialPost inserts a new social post log into the database.
func (db *DB) CreateSocialPost(ctx context.Context, post *SocialPost) error {
	query := `
		INSERT INTO social_posts (brand, platform, account_username, post_content, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	if post.CreatedAt.IsZero() {
		post.CreatedAt = time.Now()
	}
	err := db.Conn.QueryRowContext(ctx, query,
		post.Brand,
		post.Platform,
		post.AccountUsername,
		post.PostContent,
		post.Status,
		post.CreatedAt,
	).Scan(&post.ID)
	if err != nil {
		return fmt.Errorf("failed to create social post: %w", err)
	}
	return nil
}
