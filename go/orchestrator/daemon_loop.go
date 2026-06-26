package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	foundationorchestration "github.com/NexusSoftMDMA/TormentNexus/foundation/orchestration"
)

var daemonRunning bool

// StartKeeperDaemon natively replicates `daemon.ts` running an asynchronous evaluation sweeping TS legacy boundaries.
func StartKeeperDaemon(queue *TaskQueue, ws *TelemetrySocket) {
	log.Println("[Daemon] Starting Native Watchdog Monitoring Sequence...")
	if daemonRunning {
		return
	}
	daemonRunning = true

	go func() {
		for {
			// Pull bounds explicitly every minute dynamically defining timeouts Native to Go
			runLoop(queue, ws)

			// Extract timeout boundaries defined by operators dynamically via DB SQLite rows
			var settings KeeperSettings
			if err := DB.First(&settings, "id = ?", "default").Error; err == nil && settings.CheckIntervalSeconds > 0 {
				time.Sleep(time.Duration(settings.CheckIntervalSeconds) * time.Second)
			} else {
				time.Sleep(60 * time.Second)
			}
		}
	}()
}

func runLoop(queue *TaskQueue, ws *TelemetrySocket) {
	var settings KeeperSettings
	if err := DB.First(&settings, "id = ?", "default").Error; err != nil {
		log.Printf("[Daemon] Settings inaccessible skipping background sequence...")
		return
	}

	log.Println("[Daemon] Background Sweeper Scanning Active Fleet Bounds...")

	var activeSessions []Session
	DB.Find(&activeSessions)
	sessionIDs := make([]string, 0, len(activeSessions))
	for _, session := range activeSessions {
		sessionIDs = append(sessionIDs, session.ID)
	}

	var indexingJobs int64
	DB.Model(&QueueJob{}).Where("type = ? AND status = ?", "index_codebase", "pending").Count(&indexingJobs)
	plan := foundationorchestration.BuildDaemonSweepPlan(settings.IsEnabled, settings.JulesApiKey, sessionIDs, indexingJobs > 0)
	if plan.SkipReason != "" {
		log.Printf("[Daemon] Sweep skipped: %s", plan.SkipReason)
		return
	}

	queuedCount := 0
	for _, action := range plan.QueueActions {
		switch {
		case action == "index_codebase":
			DB.Create(&QueueJob{
				ID:      fmt.Sprintf("index-%s", uuid.New().String()[:8]),
				Type:    "index_codebase",
				Payload: "{}",
				Status:  "pending",
			})
			queuedCount++
		case len(action) > len("check_session:") && action[:len("check_session:")] == "check_session:":
			sessionID := action[len("check_session:"):]
			payloadBytes, _ := json.Marshal(map[string]interface{}{
				"session": sessionID,
			})
			DB.Create(&QueueJob{
				ID:      fmt.Sprintf("chk-%s-%s", sessionID, uuid.New().String()[:8]),
				Type:    "check_session",
				Payload: string(payloadBytes),
				Status:  "pending",
			})
			queuedCount++
		}
	}

	if queuedCount > 0 {
		log.Printf("[Daemon] %s", plan.Summary)
	}

	rawJson, _ := json.Marshal(map[string]interface{}{
		"status":       "ok",
		"summary":      plan.Summary,
		"queueActions": plan.QueueActions,
	})
	ws.Broadcast(fmt.Sprintf(`{"event": "%s", "payload": %s}`, plan.TelemetryEvent, string(rawJson)))
}
