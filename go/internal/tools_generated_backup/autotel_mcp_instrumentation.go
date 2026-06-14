package tools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

func HandleStartTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	traceBytes := make([]byte, 16)
	spanBytes := make([]byte, 8)
	_, e := rand.Read(traceBytes)
	if e != nil {
		return err("failed to generate trace ID")
}

	_, e = rand.Read(spanBytes)
	if e != nil {
		return err("failed to generate span ID")
}

	traceID := hex.EncodeToString(traceBytes)
	spanID := hex.EncodeToString(spanBytes)
	return ok("Trace started: " + name + " traceID=" + traceID + " spanID=" + spanID)
}