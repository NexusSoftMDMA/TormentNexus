package tools

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tormentnexushq/tormentnexus-go/internal/skillregistry"
)

func HandleEvolve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cwd, _ := os.Getwd()
	dbPath := filepath.Join(cwd, "tormentnexus.db")

	db, dbErr := sql.Open("sqlite", dbPath)
	if dbErr != nil {
		return err("failed to open database: " + dbErr.Error())
	}
	defer db.Close()

	deactivated, evolveErr := skillregistry.EvolveSkills(ctx, cwd, db)
	if evolveErr != nil {
		return err("evolution run failed: " + evolveErr.Error())
	}

	return ok(fmt.Sprintf("Evolution algorithm run complete. Deactivated %d low-performing tools (win-rate < 50%%).", deactivated))
}
