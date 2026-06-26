package agents

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// SocialPosterWorker handles scheduled posting to social media platforms.
type SocialPosterWorker struct {
	db  *db.DB
	llm llm.LLMProvider
}

// NewSocialPosterWorker creates a new social poster worker.
func NewSocialPosterWorker(database *db.DB, llmProvider llm.LLMProvider) *SocialPosterWorker {
	return &SocialPosterWorker{
		db:  database,
		llm: llmProvider,
	}
}

// Run starts the background social media posting loop.
func (w *SocialPosterWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("SocialPoster: Dual-branded social media posting worker started (interval: %v)...", interval))

	// Run first post immediately
	w.postAll(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("SocialPoster: Worker stopping...")
			return
		case <-ticker.C:
			w.postAll(ctx)
		}
	}
}

func (w *SocialPosterWorker) postAll(ctx context.Context) {
	platforms := []string{"reddit", "bluesky", "linkedin", "twitter"}
	
	// Separate accounts for both brands
	usernames := map[string]map[string]string{
		"tormentnexus": {
			"reddit":   "tormentnexus-reddit",
			"bluesky":  "tormentnexus.bsky.social",
			"linkedin": "TormentNexus Community",
			"twitter":  "@tormentnexus",
		},
		"hypernexus": {
			"reddit":   "hypernexus-reddit",
			"bluesky":  "hypernexus.bsky.social",
			"linkedin": "HyperNexus Enterprise",
			"twitter":  "@hypernexus_site",
		},
	}

	for _, platform := range platforms {
		// Post for TormentNexus
		w.generateAndPost(ctx, "tormentnexus", platform, usernames["tormentnexus"][platform])
		// Post for HyperNexus
		w.generateAndPost(ctx, "hypernexus", platform, usernames["hypernexus"][platform])
	}
}

func (w *SocialPosterWorker) generateAndPost(ctx context.Context, brand, platform, username string) {
	var systemPrompt string
	var fallbackContent string

	if brand == "tormentnexus" {
		systemPrompt = fmt.Sprintf("You are an expert developer marketing agent for TormentNexus, a local-first cognitive control plane for multi-agent LLM workflows (Operating System for AI models). Draft a short, engaging, and professional post for the platform %s highlighting local-first memory, resilient LLM waterfalls, and universal MCP tool parity. Do not use hashtags.", platform)
		fallbackContent = fmt.Sprintf("Struggling with multi-agent coordination? TormentNexus offers a local-first cognitive control plane with progressive MCP tool routing, 14K+ persisted memories, and zero-downtime provider waterfalls. Get started today!")
	} else {
		systemPrompt = fmt.Sprintf("You are an expert enterprise devrel agent for HyperNexus (hypernexus.site), the secure cloud-hosted version of TormentNexus. Draft a short, engaging, and professional post for the platform %s targeting enterprise AI teams. Highlight SSO/OIDC, RBAC, audit trails, and our stable fork at github.com/HyperNexusSoft/HyperNexus. Do not use hashtags.", platform)
		fallbackContent = fmt.Sprintf("Scale your enterprise agentic workflows with HyperNexus (hypernexus.site). Cloud-hosted, SOC 2 compliant, featuring SSO, RBAC, and audit logs. Backed by our stable open-source fork at github.com/HyperNexusSoft/HyperNexus.")
	}

	prompt := llm.Prompt{
		System: systemPrompt,
		User:   fmt.Sprintf("Draft a post for %s. Keep it concise, under 280 characters.", platform),
	}

	content, err := w.llm.Generate(ctx, prompt)
	if err != nil || content == "" {
		content = fallbackContent
	}
	content = strings.TrimSpace(content)

	// Log simulation
	slog.Info(fmt.Sprintf("SocialPoster [SIMULATION] — Brand: %s | Platform: %s | Account: %s\nContent: %s",
		brand, platform, username, content))

	if w.db != nil {
		post := &db.SocialPost{
			Brand:           brand,
			Platform:        platform,
			AccountUsername: username,
			PostContent:     content,
			Status:          "posted",
			CreatedAt:       time.Now(),
		}
		if err := w.db.CreateSocialPost(ctx, post); err != nil {
			slog.Error("SocialPoster DB Error: failed to save post log", "error", err)
		}
	}
}
