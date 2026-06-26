package web

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/auth"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/logging"
)

// HermesHealthChecker is an optional interface for checking LLM provider health.
type HermesHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Server handles web dashboard requests.
type Server struct {
	db          *db.DB
	deploy      *deploy.Deployer
	tracker     deploy.CITracker
	tasks       *autodev.TaskManager
	auth        *auth.Authenticator
	llmProvider llm.LLMProvider
	mux         *http.ServeMux
}

// NewServer creates a new Server instance.
func NewServer(database *db.DB, deployer *deploy.Deployer, tracker deploy.CITracker, taskManager *autodev.TaskManager, llmProvider llm.LLMProvider) *Server {
	s := &Server{
		db:          database,
		deploy:      deployer,
		tracker:     tracker,
		tasks:       taskManager,
		auth:        auth.NewAuthenticator(),
		llmProvider: llmProvider,
		mux:         http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Protected routes
	s.mux.Handle("/", s.auth.Middleware(http.HandlerFunc(s.handleDashboard)))

	// Public routes
	s.mux.HandleFunc("/login", s.auth.HandleLogin)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/health/detailed", s.handleDetailedHealth)
	s.mux.HandleFunc("/api/v1/webhook/github", s.handleGitHubWebhook)
		s.mux.HandleFunc("/api/v1/quote", s.handleGenerateQuote)
	s.mux.Handle("/api/v1/leads", s.auth.Middleware(http.HandlerFunc(s.handleLeadsAPI)))
	s.mux.Handle("/api/v1/deals", s.auth.Middleware(http.HandlerFunc(s.handleDealsAPI)))
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	logging.Init("json", "info")
	slog.Info("Web dashboard starting", "addr", addr)
	// #nosec G114 -- Simple ListenAndServe is used for internal dashboard; timeout configuration handled at higher level if needed
	return http.ListenAndServe(addr, logging.Middleware(s))
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		switch action {
		case "enrich":
			// #nosec G706 -- deal_id is used for context in manual action logs
			slog.InfoContext(r.Context(), "Manual enrichment triggered", "deal_id", r.FormValue("deal_id"))
		case "sync":
			if err := s.deploy.ExecuteSync(); err != nil {
				slog.WarnContext(r.Context(), "Sync error", "error", err)
			}
		case "flag_success":
			interactionID := r.FormValue("interaction_id")
			success := r.FormValue("success") == "true"
			var id int64
			if _, err := fmt.Sscanf(interactionID, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid interaction ID", "error", err, "interaction_id", interactionID)
			} else {
				if err := s.db.UpdateInteractionSuccess(r.Context(), id, success); err != nil {
					slog.WarnContext(r.Context(), "Error flagging interaction", "error", err, "interaction_id", id)
				}
			}
		case "update_channel":
			contactID := r.FormValue("contact_id")
			channel := r.FormValue("channel")
			var id int64
			if _, err := fmt.Sscanf(contactID, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid contact ID", "error", err, "contact_id", contactID)
			} else {
				if err := s.db.UpdateContactPreferredChannel(r.Context(), id, channel); err != nil {
					slog.WarnContext(r.Context(), "Error updating channel", "error", err, "contact_id", id, "channel", channel)
				} else {
					slog.InfoContext(r.Context(), "Contact channel updated", "contact_id", id, "channel", channel)
				}
			}
case "build":
			if err := s.deploy.ExecuteBuild(); err != nil {
				slog.WarnContext(r.Context(), "Build error", "error", err)
			}
		case "approve_deal":
			dealIDStr := r.FormValue("deal_id")
			var id int64
			if _, err := fmt.Sscanf(dealIDStr, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid deal ID for approval", "error", err)
			} else {
				if err := s.db.UpdateDealState(r.Context(), id, db.StateNegotiating); err != nil {
					slog.WarnContext(r.Context(), "Failed to approve deal", "deal_id", id, "error", err)
				} else {
					slog.InfoContext(r.Context(), "Deal approved by human", "deal_id", id)
				}
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	deals, err := s.db.ListRecentDeals(r.Context(), 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve deals: %v", err), http.StatusInternalServerError)
		return
	}

	health, _ := s.tracker.GetSystemHealth(r.Context())

	metrics, err := s.db.GetPerformanceMetrics(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "Error retrieving metrics", "error", err)
		metrics = &db.PerformanceMetrics{
			LeadsByState: make(map[db.LeadState]int),
		}
	}

	prs, err := s.db.ListActivePullRequests(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "Error listing PRs", "error", err)
	}

	taskList, _ := s.tasks.ListAllTasks(r.Context())

	// Check LLM/Hermes health for the dashboard
	llmStatus := "Mock"
	llmColor := "#6c757d"
	if checker, ok := s.llmProvider.(HermesHealthChecker); ok {
		if err := checker.HealthCheck(r.Context()); err != nil {
			llmStatus = fmt.Sprintf("Hermes: ERROR (%v)", err)
			llmColor = "#dc3545"
		} else {
			llmStatus = "Hermes: Connected"
			llmColor = "#28a745"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
<title>Enterprise Sales Bot Dashboard</title>
<style>
body { font-family: sans-serif; margin: 40px; background-color: #f4f4f9; }
table { width: 100%%; border-collapse: collapse; margin-top: 20px; background: white; }
th, td { padding: 12px; border: 1px solid #ddd; text-align: left; }
th { background-color: #007bff; color: white; }
tr:nth-child(even) { background-color: #f2f2f2; }
h1 { color: #333; }
.status { font-weight: bold; padding: 4px 8px; border-radius: 4px; cursor: help; }
.status-Discovered { background-color: #e2e3e5; color: #383d41; }
.status-Researched { background-color: #cce5ff; color: #004085; }
.status-PR { background-color: #fff3cd; color: #856404; }
.action-btn { background-color: #28a745; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; }
.action-btn:hover { background-color: #218838; }
.deploy-section { margin-top: 30px; padding: 20px; border: 1px solid #ccc; border-radius: 8px; background: #fff; }
.deploy-btn { background-color: #007bff; margin-right: 10px; }
</style>
</head>
<body>
<h1>Sales Bot Lead Dashboard</h1>
<p>Total Recent Leads: %d</p>
<table>
<tr>
<th>Deal ID</th>
<th>Company ID</th>
<th>State</th>
<th>Last Updated</th>
<th>Actions</th>
</tr>`, len(deals))

	for _, d := range deals {
		statusTitle := ""
		switch d.CurrentState {
		case db.StateDiscovered:
			statusTitle = "Company identified, awaiting technical research."
		case db.StateResearched:
			statusTitle = "Key engineering contacts found and technical dossier compiled."
		case db.StatePendingApproval:
			statusTitle = "Deal flagged for human approval before progression."
		}

		// Retrieve contacts and latest interaction ID for each deal
		contacts, _ := s.db.ListContactsByCompany(r.Context(), d.CompanyID)
		latestInteractionID := int64(0)
		var contactHTML string
		if len(contacts) > 0 {
			interactions, _ := s.db.ListInteractionsByContact(r.Context(), contacts[0].ID)
			if len(interactions) > 0 {
				latestInteractionID = interactions[0].ID
			}

			// Build contacts HTML with channel preference dropdown
			contactHTML = `<div style="margin-top: 8px; font-size: 0.9em;">`
			for _, c := range contacts {
				channel := c.PreferredChannel
				if channel == "" {
					channel = "email"
				}
				contactHTML += fmt.Sprintf(`
				<div style="margin: 4px 0;">
					<strong>%s</strong> (%s) — 
					<span style="color: %s;">%s</span>
					<form method="POST" style="display:inline; margin-left: 8px;">
						<input type="hidden" name="action" value="update_channel">
						<input type="hidden" name="contact_id" value="%d">
						<select name="channel" onchange="this.form.submit()" style="font-size: 0.85em; padding: 2px 4px;">
							<option value="email"%s>Email</option>
							<option value="linkedin"%s>LinkedIn</option>
							<option value="github"%s>GitHub</option>
						</select>
					</form>
				</div>`,
					html.EscapeString(c.Name),
					html.EscapeString(c.Role),
					"#17a2b8", html.EscapeString(channel),
					c.ID,
					map[bool]string{true: " selected", false: ""}[channel == "email"],
					map[bool]string{true: " selected", false: ""}[channel == "linkedin"],
					map[bool]string{true: " selected", false: ""}[channel == "github"])
			}
			contactHTML += `</div>`
		}

		_, _ = fmt.Fprintf(w, `
<tr>
<td>%d</td>
<td>%d</td>
<td><span class="status status-%s" title="%s">%s</span></td>
<td>%s</td>
<td>
<form method="POST" style="display:inline;">
<input type="hidden" name="action" value="enrich">
<input type="hidden" name="deal_id" value="%d">
<button type="submit" class="action-btn">Trigger Enrichment</button>
</form>
<form method="POST" style="display:inline;">
<input type="hidden" name="action" value="flag_success">
<input type="hidden" name="interaction_id" value="%d">
<input type="hidden" name="success" value="true">
<button type="submit" class="action-btn" style="background-color: #6f42c1;">Flag Success</button>
</form>
</td>
</tr>%s`, d.ID, d.CompanyID, d.CurrentState, statusTitle, d.CurrentState, d.UpdatedAt.Format("2006-01-02 15:04:05"), d.ID, latestInteractionID, contactHTML)
	}

	_, _ = fmt.Fprintf(w, `
</table>
<div class="deploy-section" style="border-top: 5px solid #17a2b8;">
<h2>Performance Metrics</h2>
<p>Real-time pipeline statistics and conversion rates.</p>
<div style="display: flex; gap: 20px; flex-wrap: wrap;">
<div style="background: #e9ecef; padding: 15px; border-radius: 8px; min-width: 150px;">
<strong>Total Leads:</strong> %d
</div>
<div style="background: #d4edda; padding: 15px; border-radius: 8px; min-width: 150px;">
<strong>Won Deals:</strong> %d
</div>
<div style="background: #f8d7da; padding: 15px; border-radius: 8px; min-width: 150px;">
<strong>Win Rate:</strong> %.1f%%
</div>
<div style="background: #fff3cd; padding: 15px; border-radius: 8px; min-width: 150px;">
<strong>Successful Outreach:</strong> %d
</div>
</div>
<h3>Leads by State</h3>
<ul>
<li><strong>Discovered:</strong> %d</li>
<li><strong>Researched:</strong> %d</li>
<li><strong>Outreach Sent:</strong> %d</li>
<li><strong>Engaged:</strong> %d</li>
<li><strong>Negotiating:</strong> %d</li>
</ul>
</div>`, metrics.TotalLeads, metrics.LeadsByState[db.StateClosedWon], metrics.WinRate, metrics.SuccessfulOutreach, metrics.LeadsByState[db.StateDiscovered], metrics.LeadsByState[db.StateResearched], metrics.LeadsByState[db.StateOutreachSent], metrics.LeadsByState[db.StateEngaged], metrics.LeadsByState[db.StateNegotiating])

	_, _ = fmt.Fprintf(w, `
<div class="deploy-section">
<h2>Autonomous Task Board</h2>
<p>Prioritized development roadmap and execution status.</p>
<table>
<tr>
<th>Description</th>
<th>Status</th>
</tr>`)

	for _, t := range taskList {
		status := "Pending"
		if t.Completed {
			status = "Completed"
		}
		_, _ = fmt.Fprintf(w, `
<tr>
<td>%s</td>
<td><span class="status status-%s">%s</span></td>
</tr>`, html.EscapeString(t.Description), status, status)
	}

	_, _ = fmt.Fprintf(w, `
</table>
</div>
<div class="deploy-section">
<h2>Autonomous Pull Requests</h2>
<p>Active feature branches and automated merge status.</p>
<table>
<tr>
<th>PR ID</th>
<th>Branch</th>
<th>Title</th>
<th>Status</th>
</tr>`)

	for _, pr := range prs {
		_, _ = fmt.Fprintf(w, `
<tr>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td><span class="status status-PR">%s</span></td>
</tr>`, html.EscapeString(pr.ID), html.EscapeString(pr.Branch), html.EscapeString(pr.Title), html.EscapeString(string(pr.Status)))
	}

	_, _ = fmt.Fprintf(w, `
</table>
</div>
<div class="deploy-section">
<h2>Self-Service Deployment</h2>
<p>Manage repository state and trigger system builds autonomously.</p>
<form method="POST" style="display:inline;">
<input type="hidden" name="action" value="sync">
<button type="submit" class="action-btn deploy-btn">Sync Repository</button>
</form>
<form method="POST" style="display:inline;">
<input type="hidden" name="action" value="build">
<button type="submit" class="action-btn deploy-btn" style="background-color: #6c757d;">Trigger Build</button>
</form>
</div>
<div class="deploy-section" style="border-left: 5px solid #28a745;">
<h2>System Health &amp; CI Status</h2>
<p>Real-time monitoring of the autonomous deployment pipeline.</p>
<ul>
<li><strong>Global Health:</strong> <span style="color: #28a745;">%s</span></li>
<li><strong>LLM Provider:</strong> <span style="color: %s;">%s</span></li>
</ul>
</div>
</body>
</html>`, health, llmColor, llmStatus)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "OK")
}

func (s *Server) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := make(map[string]interface{})

	// 1. Check DB
	if err := s.db.Conn.PingContext(r.Context()); err != nil {
		health["database"] = "ERROR: " + err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["database"] = "OK"
	}

	// 2. Check CI/Sync status
	healthStatus, _ := s.tracker.GetSystemHealth(r.Context())
	health["system_health"] = healthStatus

	// 3. Worker liveness
	health["workers"] = "active"

	// 4. Check LLM/Hermes connectivity
	if checker, ok := s.llmProvider.(HermesHealthChecker); ok {
		if err := checker.HealthCheck(r.Context()); err != nil {
			health["llm_provider"] = "ERROR: " + err.Error()
		} else {
			health["llm_provider"] = "Hermes: Connected"
		}
	} else {
		health["llm_provider"] = "Mock"
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		slog.WarnContext(r.Context(), "Error encoding health JSON", "error", err)
	}
}

func verifySignature(payload []byte, secret string, signatureHeader string) bool {
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}
	actualSignature := signatureHeader[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// SECURITY: Verify GitHub Webhook Signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		slog.WarnContext(r.Context(), "Webhook: Missing X-Hub-Signature-256 header")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Real implementation of signature verification logic
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret != "" {
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}
		// Reset body for later use if needed (though not used here yet)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		if !verifySignature(body, secret, signature) {
			slog.WarnContext(r.Context(), "Webhook: Invalid signature")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	} else {
		slog.WarnContext(r.Context(), "Webhook: GITHUB_WEBHOOK_SECRET not set, skipping verification (Insecure!)")
	}

	slog.InfoContext(r.Context(), "Webhook: Received GitHub push event, triggering deployment...")

	// Trigger sync and build in a goroutine to avoid blocking the webhook response
	go func() {
		if err := s.deploy.ExecuteSync(); err != nil {
			slog.Warn("Webhook: Sync failed", "error", err)
			return
		}
		if err := s.deploy.ExecuteBuild(); err != nil {
			slog.Warn("Webhook: Build failed", "error", err)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintln(w, "Deployment triggered")
}

func (s *Server) handleGenerateQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tier := r.URL.Query().Get("company_size")
	if tier == "" {
		tier = r.URL.Query().Get("market_cap_tier")
	}

	quote := communication.CalculateQuote(tier)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"tier":  tier,
		"quote": quote,
	}); err != nil {
		slog.ErrorContext(r.Context(), "Error encoding quote JSON", "error", err)
	}
}

// REST API for external pipeline management
func (s *Server) handleLeadsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Example: List all leads
		companies, err := s.db.ListAllCompanies(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve leads", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(companies)
	case http.MethodPost:
		// Example: Create a new lead
		var lead db.Company
		if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := s.db.CreateCompany(r.Context(), &lead); err != nil {
			http.Error(w, "Failed to create lead", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(lead)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDealsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Example: List all deals, could support filtering by state via query params
		stateFilter := r.URL.Query().Get("state")
		var deals []db.Deal
		var err error
		if stateFilter != "" {
			deals, err = s.db.ListDealsByState(r.Context(), db.LeadState(stateFilter))
		} else {
			// Add a repository method to list all deals if needed,
			// or just fall back to a specific state for now.
			deals, err = s.db.ListDealsByState(r.Context(), db.StateDiscovered) // Placeholder
		}

		if err != nil {
			http.Error(w, "Failed to retrieve deals", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deals)
	case http.MethodPost:
		// Example: Create a new deal
		var deal db.Deal
		if err := json.NewDecoder(r.Body).Decode(&deal); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := s.db.CreateDeal(r.Context(), &deal); err != nil {
			http.Error(w, "Failed to create deal", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(deal)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
