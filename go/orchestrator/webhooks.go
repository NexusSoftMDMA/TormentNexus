package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	foundationorchestration "github.com/NexusSoftMDMA/TormentNexus/foundation/orchestration"
)

type WebhookPayload struct {
	Type   string          `json:"type"`
	Source string          `json:"source"`
	Data   json.RawMessage `json:"data"`
}

// HandleTormentNexusWebhook processes real-time synchronization routes mirroring the JULES TS implementation.
func HandleTormentNexusWebhook(payload WebhookPayload, queue *TaskQueue, ws *TelemetrySocket) (map[string]interface{}, error) {
	source := payload.Source
	if source == "" {
		source = "unknown"
	}

	// 1. Log natively into GORM TormentNexusa mapped model
	actionLog := KeeperLog{
		ID:        uuid.New().String(),
		SessionId: "global",
		Type:      "info",
		Message:   fmt.Sprintf("Received TormentNexus signal: %s from %s", payload.Type, source),
		Metadata:  string(payload.Data),
	}
	if err := DB.Create(&actionLog).Error; err != nil {
		log.Printf("[Webhooks] ORM Insert skipped: %v", err)
	}

	// 2. Build a foundation-backed webhook plan and apply it.
	plan := foundationorchestration.BuildWebhookPlan(payload.Type, source)
	for _, action := range plan.QueueActions {
		log.Printf("[Webhooks] Queueing action %s from plan for signal %s", action, payload.Type)
		queue.Enqueue(action)
	}
	if plan.ClearLogs {
		DB.Where("1 = 1").Delete(&KeeperLog{})
	}
	if len(plan.QueueActions) == 0 && !plan.ClearLogs {
		log.Printf("[Webhooks] Unmapped signal subtype ignored: %s", payload.Type)
	}

	// 3. Emit live telemetry notification explicitly matching DaemonEventType logic in 'shared'
	emitPayload := map[string]interface{}{
		"type":      payload.Type,
		"source":    source,
		"summary":   plan.Summary,
		"actions":   plan.QueueActions,
		"data":      payload.Data,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	rawJson, _ := json.Marshal(emitPayload)
	ws.Broadcast(fmt.Sprintf(`{"event": "tormentnexus_signal_received", "payload": %s}`, string(rawJson)))

	return map[string]interface{}{"success": true, "processed": true, "plan": plan}, nil
}
