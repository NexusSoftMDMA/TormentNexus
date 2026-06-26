package hsync

/**
 * @file high_value.go
 * @module go/internal/hsync
 *
 * WHAT: Go-native implementation of the HighValueIngestor.
 * Performs deep semantic extraction for high-value external resources.
 */

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
	"github.com/tormentnexushq/tormentnexus-go/internal/harnesses"
	mcp_pkg "github.com/tormentnexushq/tormentnexus-go/internal/mcp"
)

type HighValueIngestor struct {
	dbPath     string
	skillStore *harnesses.SkillStore
	mcpConfig  *mcp_pkg.ConfigManager
}

func NewHighValueIngestor(dbPath string, skillStore *harnesses.SkillStore, mcpConfig *mcp_pkg.ConfigManager) *HighValueIngestor {
	return &HighValueIngestor{
		dbPath:     dbPath,
		skillStore: skillStore,
		mcpConfig:  mcpConfig,
	}
}

func (i *HighValueIngestor) ProcessHighValueQueue(ctx context.Context, limit int) error {
	db, err := sql.Open("sqlite", i.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Select items with many stars or specific tags
	rows, err := db.QueryContext(ctx, `
		SELECT uuid, url, page_title, page_description, tags
		FROM links_backlog
		WHERE research_status = 'done'
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var uuid, url, title, desc, tagsRaw string
		if err := rows.Scan(&uuid, &url, &title, &desc, &tagsRaw); err != nil {
			continue
		}

		// Filter for "High Value" (placeholder heuristic: mcp-server tag)
		if strings.Contains(tagsRaw, "mcp-server") || strings.Contains(tagsRaw, "high-value") {
			fmt.Printf("[Go HighValue] 💎 Deep diving into: %s\n", url)
			i.deepDive(ctx, uuid, url, title, desc)
		}
	}

	return nil
}

func (i *HighValueIngestor) deepDive(ctx context.Context, uuidValue, url, title, desc string) {
	prompt := fmt.Sprintf(`
		Analyze this resource for the TormentNexus Control Plane:
		Title: %s
		Description: %s
		URL: %s

		Extract:
		1. MCP recipe (JSON) if it's a server.
		2. Skill instructions if it's a runbook.
		
		Return JSON only.
	`, title, desc, url)

	resp, err := ai.AutoRoute(ctx, []ai.Message{
		{Role: "system", Content: "You are a technical analyst."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return
	}

	// 1. If it's a Skill, save it natively
	var analysis struct {
		IsSkill      bool                `json:"isSkill"`
		SkillContent string              `json:"skillContent"`
		IsMcpServer  bool                `json:"isMcpServer"`
		McpRecipe    *mcp_pkg.McpServerConfig `json:"mcpRecipe"`
		Summary      string              `json:"summary"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &analysis); err == nil {
		if analysis.IsSkill && analysis.SkillContent != "" {
			skillID := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
			err := i.skillStore.SaveSkill(skillID, title, desc, analysis.SkillContent)
			if err == nil {
				fmt.Printf("[Go HighValue] 🧠 Saved new native skill: %s\n", skillID)
			}
		}

		if analysis.IsMcpServer && analysis.McpRecipe != nil {
			serverName := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
			err := i.mcpConfig.AddServer(serverName, *analysis.McpRecipe)
			if err == nil {
				fmt.Printf("[Go HighValue] 📦 Registered new native MCP server: %s\n", serverName)
			}
		}
	}

	fmt.Printf("[Go HighValue] 🔍 Analysis complete for %s\n", url)
}
