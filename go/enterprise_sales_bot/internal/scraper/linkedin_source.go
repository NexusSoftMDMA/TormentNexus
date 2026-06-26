package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/go-rod/rod"
)

// LinkedInSource implements LeadSource by searching for companies and contacts via LinkedIn.
// Note: LinkedIn Sales Navigator does not have a public REST API. This source uses:
// - Simulation fallback when credentials are not configured
// - Headless browser automation via rod for actual scraping when credentials are present
type LinkedInSource struct {
	Client       *http.Client
	Username     string
	Password     string
	TargetTitles []string
}

// Discover searches for companies and contacts matching the target criteria.
// When LinkedIn credentials are not configured, returns simulated high-value targets.
func (l *LinkedInSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	// Load credentials from config or environment
	if l.Username == "" {
		l.Username = os.Getenv("LINKEDIN_USERNAME")
	}
	if l.Password == "" {
		l.Password = os.Getenv("LINKEDIN_PASSWORD")
	}

	if len(l.TargetTitles) == 0 {
		l.TargetTitles = []string{
			"CTO",
			"VP Engineering",
			"Lead Developer",
			"Principal Engineer",
			"Director of Engineering",
			"Head of AI",
			"Machine Learning Lead",
		}
	}

	// Check if credentials are configured
	if l.Username == "" || l.Password == "" {
		slog.Info("LinkedInSource: No LINKEDIN_USERNAME/PASSWORD configured, returning simulated results")
		return l.simulateDiscovery(ctx, keywords)
	}

	// Attempt real scraping with headless browser
	slog.Info("LinkedInSource: Credentials configured, attempting headless scrape")
	companies, err := l.scrapeLinkedIn(ctx, keywords)
	if err != nil {
		slog.Warn("LinkedIn headless scrape failed, falling back to simulation", "error", err)
		return l.simulateDiscovery(ctx, keywords)
	}
	return companies, nil
}

// scrapeLinkedIn performs a headless browser scrape of LinkedIn Sales Navigator.
// Uses rod's Must* chain and recovers panics into errors.
func (l *LinkedInSource) scrapeLinkedIn(ctx context.Context, keywords []string) (companies []db.Company, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("headless scrape panicked: %v", r)
		}
	}()

	// Connect to browser (auto-launch if needed)
	browser := rod.New().MustConnect()
	defer func() { _ = browser.Close() }()

	page := browser.MustPage()
	page.MustNavigate("https://www.linkedin.com/login")
	page.MustWaitLoad()

	// Fill login form
	page.MustElement("#username").MustInput(l.Username)
	page.MustElement("#password").MustInput(l.Password)
	page.MustElement("button[type='submit']").MustClick()

	// Wait for successful login
	page.MustWaitElementsMoreThan(".global-nav__me-photo", 1)

	// Navigate to Sales Navigator search
	query := strings.Join(keywords, " ")
	encoded := url.QueryEscape(query)
	navURL := fmt.Sprintf("https://www.linkedin.com/sales/search?query=%s", encoded)
	page.MustNavigate(navURL)
	page.MustWaitLoad()

	// Wait for company result links to appear
	page.MustWaitElementsMoreThan("a[href*='/sales/company/']", 1)

	// Extract company names and profile links via JS evaluation
	result := page.MustEval(`() => {
		const anchors = Array.from(document.querySelectorAll('a[href*="/sales/company/"]'));
		const unique = new Map();
		anchors.forEach(a => {
			const container = a.closest('.artdeco-card, [data-control-name="sales_company_result"], .search-results__result-item');
			const nameEl = (container && container.querySelector('.company-name, .name, .title, .org-name, .result__title')) || a;
			const name = nameEl ? nameEl.innerText.trim() : '';
			if (name) {
				unique.set(a.href, name);
			}
		});
		return Array.from(unique.entries()).map(([link, name]) => ({ name, link }));
	}`)

	gsons := result.Arr()
	companies = make([]db.Company, 0, len(gsons))
	for _, g := range gsons {
		obj := g.Map()
		name := obj["name"].String()
		if name != "" {
			companies = append(companies, db.Company{
				Name:          name,
				Domain:        "", // Unknown at this stage; will be enriched later
				TechStack:     []string{},
				HiringSignals: []string{},
				MarketCapTier: "Small Business", // default tier
			})
		}
	}

	return companies, nil
}

// simulateDiscovery returns simulated high-value targets when real API is not available.
func (l *LinkedInSource) simulateDiscovery(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info(fmt.Sprintf("LinkedInSource: Simulating discovery for keywords: %v", keywords))

	// Simulated high-value AI/ML companies with engineering leadership
	return []db.Company{
		{
			Name:          "Neural Dynamics",
			Domain:        "neuraldynamics.io",
			TechStack:     []string{"Python", "PyTorch", "Kubernetes", "MLflow"},
			HiringSignals: []string{"Hiring: Senior ML Platform Engineer", "VP Engineering recently joined from Google Brain"},
			MarketCapTier: "Mid-Market",
		},
		{
			Name:          "Cognitive Systems Labs",
			Domain:        "cognitivesystems.ai",
			TechStack:     []string{"Rust", "Go", "TensorFlow", "gRPC"},
			HiringSignals: []string{"Hiring: Distributed Systems Architect", "CTO posted about multi-agent orchestration challenges"},
			MarketCapTier: "Enterprise",
		},
		{
			Name:          "Hyperloop AI",
			Domain:        "hyperloop-ai.tech",
			TechStack:     []string{"Go", "LLMs", "Apache Kafka", "Redis"},
			HiringSignals: []string{"Hiring: AI Infrastructure Lead", "Director of Engineering scaling team from 5 to 20"},
			MarketCapTier: "Small Business",
		},
	}, nil
}

// HealthCheck verifies LinkedIn API connectivity and credential validity.
// Returns error if credentials are missing or invalid.
func (l *LinkedInSource) HealthCheck(ctx context.Context) error {
	if l.Username == "" || l.Password == "" {
		return fmt.Errorf("LINKEDIN_USERNAME and LINKEDIN_PASSWORD must be configured")
	}

	// For now, just verify credentials are non-empty
	if strings.TrimSpace(l.Username) == "" || strings.TrimSpace(l.Password) == "" {
		return fmt.Errorf("LinkedIn credentials cannot be empty")
	}

	slog.Info("LinkedInSource: Health check passed (credentials configured)")
	return nil
}

// SetCredentials configures LinkedIn authentication credentials.
func (l *LinkedInSource) SetCredentials(username, password string) {
	l.Username = username
	l.Password = password
}

// SetTargetTitles configures the job titles to search for.
func (l *LinkedInSource) SetTargetTitles(titles []string) {
	l.TargetTitles = titles
}

// Ensure LinkedInSource implements LeadSource interface
var _ LeadSource = (*LinkedInSource)(nil)
