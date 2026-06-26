package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"os/signal"
	"syscall"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/billing"
	"github.com/robertpelloni/enterprise_sales_bot/internal/config"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/sales"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
	"github.com/robertpelloni/enterprise_sales_bot/internal/web"
	"github.com/robertpelloni/enterprise_sales_bot/pkg/agents"

	_ "github.com/lib/pq"	// PostgreSQL driver
)

func main() {
	reconcile := flag.Bool("reconcile", false, "Run branch reconciliation and exit")
	inventory := flag.Bool("inventory", false, "Generate submodule inventory and exit")
	flag.Parse()

	if *inventory {
		slog.Info("Generating submodule inventory")
		table, err := gitcheck.GenerateSubmoduleInventory()
		if err != nil {
			slog.Error("Failed to generate inventory", "error", err)
			os.Exit(1)
		}
		fmt.Println(table)
		return
	}

	if *reconcile {
		slog.Info("Running intelligent merge engine")
		if err := gitres.ReconcileBranches(); err != nil {
			slog.Error("Reconciliation failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Reconciliation complete")
		return
	}

	slog.Info("Starting TormentNexus Autonomous Sales Bot")

	// 0. Load Configuration
	cfg := config.Load()

	// 1. Initialize Database
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not connect to database: %v", err))
	}
	defer func() { _ = database.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Setup Scraper — HN "Who is Hiring" + LinkedIn + GitHub Issues + Mock fallback
	sources := []scraper.LeadSource{
		&scraper.HNWhoIsHiringSource{Client: http.DefaultClient},
		&scraper.LinkedInSource{Client: http.DefaultClient},
		&scraper.GitHubIssueSource{Client: http.DefaultClient, Token: cfg.GitHubToken},
		&scraper.MockJobBoardSource{},
	}
	s := scraper.NewScraper(database, sources)

	keywords := []string{"AI Engineer", "LLM Orchestration", "Agentic Workflows", "AI Platform", "ML Infrastructure"}
	go s.Run(ctx, 2*time.Hour, keywords)

	// 2a. Setup CRM Integration
	var crmClient crm.CRMClient
	if cfg.CRMBaseURL != "" && cfg.CRMAPIKey != "" {
		slog.Info(fmt.Sprintf("CRM: Initializing production REST CRM client at %s", cfg.CRMBaseURL))
		crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey)
	} else {
		slog.Info("CRM: Initializing mock CRM client (missing configuration).")
		crmClient = crm.NewMockCRMClient()
	}

	// 2b. Setup Enricher — Hunter.io + Apollo.io + Mock with FallbackSource
	var enrichmentSources []enrichment.EnrichmentSource
	var sourceNames []string

	if cfg.HunterAPIKey != "" {
		slog.Info("Enrichment: Initializing Hunter.io source.")
		enrichmentSources = append(enrichmentSources, enrichment.NewHunterSource(cfg.HunterAPIKey))
		sourceNames = append(sourceNames, "Hunter.io")
	} else {
		slog.Info("Enrichment: No HUNTER_API_KEY set - skipping Hunter.io.")
	}

	if cfg.ApolloAPIKey != "" {
		slog.Info("Enrichment: Initializing Apollo.io source.")
		enrichmentSources = append(enrichmentSources, enrichment.NewApolloSource(cfg.ApolloAPIKey))
		sourceNames = append(sourceNames, "Apollo.io")
	} else {
		slog.Info("Enrichment: No APOLLO_API_KEY set - skipping Apollo.io.")
	}

	// Always add mock source as final fallback for development/testing
	if len(enrichmentSources) == 0 {
		slog.Info("Enrichment: No real sources configured - using mock source only.")
		enrichmentSources = append(enrichmentSources, &enrichment.MockApolloSource{})
		sourceNames = append(sourceNames, "Mock")
	} else {
		slog.Info("Enrichment: Mock source added as final fallback.")
		enrichmentSources = append(enrichmentSources, &enrichment.MockApolloSource{})
		sourceNames = append(sourceNames, "Mock (fallback)")
	}

	// Wrap sources in fallback chain for ordered retry with clear logging
	fallbackSource := enrichment.NewFallbackSource(enrichmentSources, sourceNames)
	slog.Info(fmt.Sprintf("Enrichment: Fallback chain configured — %s", fallbackSource.Status()))

	e := enrichment.NewEnricher(database, []enrichment.EnrichmentSource{fallbackSource}, crmClient)
	go e.Run(ctx, 1*time.Hour)

	// 2c. Setup Researcher — GitHub, Tech Blogs, and RSS Feeds
	rssFeeds := []string{
		"https://hnrss.org/frontpage?points=10",
		"https://blog.rust-lang.org/feed.xml",
		"https://go.dev/blog/feed.atom",
		"https://engineering.fb.com/feed/",
		"https://netflixtechblog.com/feed/",
		"https://github.blog/category/engineering/feed/",
	}
	crawlers := []researcher.Crawler{
		&researcher.GitHubCrawler{Client: http.DefaultClient},
		&researcher.BlogCrawler{Client: http.DefaultClient},
		&researcher.RSSFeedCrawler{
			Feeds: rssFeeds,
			Client: http.DefaultClient,
		},
	}
	processor := &researcher.DefaultDossierProcessor{}
	r := researcher.NewResearcher(database, crawlers, processor, crmClient)
	go r.Run(ctx, 1*time.Hour)

	crmWorker := crm.NewWorker(database, crmClient)
	go crmWorker.Run(ctx, 30*time.Minute)

	// 2d. Setup Target Discovery
	outreachWorker := agents.NewTargetDiscoveryWorker(database)
	go outreachWorker.Run(ctx, 2*time.Hour)

	// 2e. Setup Deployer
	var ciTracker deploy.CITracker
	var dispatcher deploy.WorkflowDispatcher

	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			slog.Info(fmt.Sprintf("CI: Initializing GitHub CI Tracker and Dispatcher for %s", cfg.GitHubRepository))
			ciTracker = deploy.NewGitHubCITracker(parts[0], parts[1])
			dispatcher = deploy.NewGitHubDispatcher(parts[0], parts[1])
		}
	}

	if ciTracker == nil {
		slog.Info("CI: Initializing Mock CI Tracker (missing GITHUB_REPOSITORY).")
		ciTracker = &deploy.MockCITracker{}
	}

	deployer := deploy.NewDeployer(ciTracker, dispatcher)
	go deployer.Run(ctx, cfg.DeploySyncInterval)
	go deployer.MonitorDeployment(ctx, cfg.DeploySyncInterval)

	// 2f. Setup LLM Provider — Hermes or Mock
	var llmProvider llm.LLMProvider
	if cfg.HermesAPIURL != "" && cfg.HermesAPIKey != "" {
		slog.Info(fmt.Sprintf("LLM: Initializing Hermes provider at %s (model: %s)", cfg.HermesAPIURL, cfg.HermesModel))
		llmProvider = llm.NewHermesLLMProvider(llm.HermesConfig{
			BaseURL:	cfg.HermesAPIURL,
			APIKey:		cfg.HermesAPIKey,
			Model:		cfg.HermesModel,
		})

		if err := llmProvider.(*llm.HermesLLMProvider).HealthCheck(ctx); err != nil {
			slog.Info(fmt.Sprintf("LLM: WARNING — Hermes health check failed: %v", err))
		} else {
			slog.Info("LLM: Hermes health check passed ✓")
		}
	} else {
		slog.Info("LLM: Initializing Mock LLM Provider (set HERMES_API_URL and HERMES_API_KEY for real LLM).")
		llmProvider = &llm.MockLLMProvider{}
	}

	// 2k. Setup Social Media Poster Agent
	socialPoster := agents.NewSocialPosterWorker(database, llmProvider)
	go socialPoster.Run(ctx, 4*time.Hour)

	// 2g. Setup Intent Classifier
	var classifier communication.IntentClassifier
	if cfg.HermesAPIURL != "" && cfg.HermesAPIKey != "" {
		slog.Info("Communication: Initializing LLM-backed Intent Classifier via Hermes.")
		classifier = communication.NewLLMIntentClassifier(llmProvider)
	} else {
		slog.Info("Communication: Initializing Mock Intent Classifier.")
		classifier = &communication.MockIntentClassifier{}
	}

	responder := communication.NewRAGResponseGenerator(database, llmProvider)
	strategy := communication.NewLearningSalesEngine(database, crmClient, llmProvider)

	// 2h. Setup Email Sender — SMTP, Draft, or Mock
	var emailSender communication.EmailSender
	if cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" && !cfg.DryRun {
		slog.Info(fmt.Sprintf("Email: Initializing SMTP sender via %s:%d as %s", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername))
		emailSender = communication.NewSMTPSender(communication.SMTPConfig{
			Host:		cfg.SMTPHost,
			Port:		cfg.SMTPPort,
			Username:	cfg.SMTPUsername,
			Password:	cfg.SMTPPassword,
			From:		cfg.SMTPFrom,
			FromName:	cfg.SMTPFromName,
		})
	} else if cfg.DryRun && cfg.IMAPHost != "" && cfg.IMAPUsername != "" && cfg.IMAPPassword != "" {
		slog.Info(fmt.Sprintf("Email: DRY RUN mode — saving drafts to %s via IMAP.", cfg.IMAPFolder))
		emailSender = communication.NewDraftSender(cfg.IMAPHost, cfg.IMAPPort, cfg.IMAPUsername, cfg.IMAPPassword)
	} else {
		if cfg.DryRun {
			slog.Info("Email: DRY RUN mode — no IMAP configured, emails will be logged only.")
		} else {
			slog.Info("Email: No SMTP configured — outbound emails will be logged but not sent.")
		}
		emailSender = &communication.MockEmailSender{}
	}

	billingClient := &billing.MockBillingClient{}
	orderProcessor := sales.NewOrderProcessor(database, billingClient, crmClient)

	commManager := communication.NewManager(database, classifier, responder, strategy, orderProcessor, emailSender)

	// Initialize Objection Library and attach to manager
	objectionLib := communication.NewObjectionLibrary()
	commManager.SetObjectionLibrary(objectionLib)
	go commManager.Run(ctx, 30*time.Minute)

	// 2j. Setup Cadence-aware outreach scheduling
	cadenceManager := communication.NewCadenceAwareManager(commManager, database)
	go cadenceManager.RunCadence(ctx, 12*time.Hour)	// checks every 12 h for next touch

	// 2i. Setup IMAP Email Receiver
	if cfg.IMAPHost != "" && cfg.IMAPUsername != "" && cfg.IMAPPassword != "" {
		slog.Info(fmt.Sprintf("Email: Initializing IMAP receiver from %s:%d (polling every %v)", cfg.IMAPHost, cfg.IMAPPort, cfg.IMAPPollInterval))
		imapReceiver := communication.NewEmailReceiver(communication.IMAPConfig{
			Host:		cfg.IMAPHost,
			Port:		cfg.IMAPPort,
			Username:	cfg.IMAPUsername,
			Password:	cfg.IMAPPassword,
			Folder:		cfg.IMAPFolder,
		}, commManager)
		go imapReceiver.Run(ctx, cfg.IMAPPollInterval)
	} else {
		slog.Info("Email: No IMAP configured — inbound emails will not be received.")
	}

	// 3. Initialize Autonomous Development
	taskManager := autodev.NewTaskManager("TODO.md")
	agent := autodev.NewLocalAgent(llmProvider)

	var prManager gitcheck.PRManager
	if cfg.GitHubRepository != "" {
		parts := strings.Split(cfg.GitHubRepository, "/")
		if len(parts) == 2 {
			slog.Info(fmt.Sprintf("Autodev: Initializing GitHub PR Manager for %s", cfg.GitHubRepository))
			prManager = gitcheck.NewGitHubPRManager(parts[0], parts[1])
		}
	}
	if prManager == nil {
		slog.Info("Autodev: Initializing Mock PR Manager (missing GITHUB_REPOSITORY).")
		prManager = &gitcheck.MockPRManager{}
	}

	orchestrator := autodev.NewOrchestrator(database, taskManager, agent, prManager, ciTracker)
	go orchestrator.Run(ctx, 1*time.Hour)

	// 4. Start Web Server
	webServer := web.NewServer(database, deployer, ciTracker, taskManager, llmProvider)

	srv := &http.Server{
		Addr:		":" + cfg.Port,
		Handler:	webServer,
	}

	go func() {
		slog.Info(fmt.Sprintf("Web Dashboard: Listening on :%s", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Info(fmt.Sprintf("Web server error: %v", err))
		}
	}()

	// 5. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	slog.Info("Shutting down: Signal received, initiating graceful drain...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Info(fmt.Sprintf("Web server shutdown error: %v", err))
	}

	time.Sleep(2 * time.Second)
	slog.Info("Shutting down: Done.")
}
