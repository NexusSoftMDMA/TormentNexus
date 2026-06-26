package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
)

func HandleBuildErrorAdapter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    serviceName, _ :=getString(args, "service_name")
    errorType, _ :=getString(args, "error_type")
    statusCode, _ :=getInt(args, "status_code")
    message, _ :=getString(args, "message")

    errorResponse := map[string]interface{}{
        "service":      serviceName,
        "error_type":   errorType,
        "status_code":  statusCode,
        "message":      message,
        "adapter_slug": "_vendor",
        "timestamp":    time.Now().Unix(),
    }

    errorJSON, marshalErr := json.Marshal(errorResponse)
    if marshalErr != nil {
        return err("Failed to marshal error response")
}

    return ok(string(errorJSON))
}

Similarly for other handlers.

Need to ensure we use e function for errors; but we also have variable named e in other functions; rename to marshalErr.

Also need to import net/http, net/url, strings maybe not used; remove them.

Now ensure all handlers have correct signatures.

Also ensure we don't redeclare ok, e, getString, getInt, getBool.

Thus final codepackage tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// HandleBuildErrorAdapter demonstrates building error adapters for Arcade TDK
func HandleNetworkTransportError(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	errorKind, _ :=getString(args, "error_kind")
	canRetry, _ :=getBool(args, "can_retry")
	endpoint, _ :=getString(args, "endpoint")
	method, _ :=getString(args, "method")

	networkError := map[string]interface{}{
		"error_class": "NetworkTransportError",
		"kind":        errorKind,
		"can_retry":   canRetry,
		"endpoint":    endpoint,
		"method":      method,
		"status_code": nil,
		"service":     "arcade_mcp",
		"error_type":  "NetworkTransportError",
		"timestamp":   time.Now().Unix(),
	}

	safeMessage := fmt.Sprintf("Network request to %s failed. Error type: %s", endpoint, errorKind)

	networkJSON, marshalErr := json.Marshal(networkError)
	if marshalErr != nil {
		return err("Failed to marshal network error")
}

	return ok(fmt.Sprintf("%s\nDetails: %s", safeMessage, string(networkJSON)))
}

// HandleUpstreamError demonstrates handling upstream HTTP errors
func HandleUpstreamError(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	statusCode, _ :=getInt(args, "status_code")
	service, _ :=getString(args, "service")
	reason, _ :=getString(args, "reason")
	endpoint, _ :=getString(args, "endpoint")

	errorClass := "UpstreamError"
	if statusCode == 429 {
		errorClass = "UpstreamRateLimitError"
	}

	upstreamError := map[string]interface{}{
		"error_class":    errorClass,
		"status_code":    statusCode,
		"service":        service,
		"reason":         reason,
		"endpoint":       endpoint,
		"retry_after_ms": nil,
		"timestamp":      time.Now().Unix(),
	}

	safeMessage := fmt.Sprintf("Upstream %s request failed with status code %d.", service, statusCode)
	if reason != "" {
		safeMessage += fmt.Sprintf(" Reason: %s", reason)

	upstreamJSON, marshalErr := json.Marshal(upstreamError)
	if marshalErr != nil {
		return err("Failed to marshal upstream error")
}

	return ok(fmt.Sprintf("%s\nDetails: %s", safeMessage, string(upstreamJSON)))
}

}

// HandleFatalToolError demonstrates handling fatal tool errors
func HandleFatalToolError(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	errorType, _ :=getString(args, "error_type")
	message, _ :=getString(args, "message")
	service, _ :=getString(args, "service")

	fatalError := map[string]interface{}{
		"error_class": "FatalToolError",
		"error_type":  errorType,
		"message":     message,
		"service":     service,
		"can_retry":   false,
		"timestamp":   time.Now().Unix(),
	}

	safeMessage := fmt.Sprintf("Tool configuration error in %s: %s", service, errorType)

	fatalJSON, marshalErr := json.Marshal(fatalError)
	if marshalErr != nil {
		return err("Failed to marshal fatal error")
}

	return ok(fmt.Sprintf("%s\nDetails: %s", safeMessage, string(fatalJSON)))
}

// HandleRetryableToolError demonstrates handling retryable tool errors
func HandleRetryableToolError(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	service, _ :=getString(args, "service")
	retryCount, _ :=getInt(args, "retry_count")

	retryableError := map[string]interface{}{
		"error_class": "RetryableToolError",
		"message":     message,
		"service":     service,
		"can_retry":   true,
		"retry_count": retryCount,
		"timestamp":   time.Now().Unix(),
	}

	safeMessage := fmt.Sprintf("Transient error in %s. Please retry the operation. Attempt %d", service, retryCount)

	retryableJSON, marshalErr := json.Marshal(retryableError)
	if marshalErr != nil {
		return err("Failed to marshal retryable error")
}

	return ok(fmt.Sprintf("%s\nDetails: %s", safeMessage, string(retryableJSON)))
}

// HandleContextRequiredToolError demonstrates handling context-required tool errors
func HandleContextRequiredToolError(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	service, _ :=getString(args, "service")
	contextFields, _ :=getString(args, "context_fields")

	contextError := map[string]interface{}{
		"error_class":      "ContextRequiredToolError",
		"message":          message,
		"service":          service,
		"context_fields":   contextFields,
		"can_retry":        false,
		"requires_context": true,
		"timestamp":        time.Now().Unix(),
	}

	safeMessage := fmt.Sprintf("Additional information required from you to proceed with %s. Please provide: %s", service, contextFields)

	contextJSON, marshalErr := json.Marshal(contextError)
	if marshalErr != nil {
		return err("Failed to marshal context error")
}

	return ok(fmt.Sprintf("%s\nDetails: %s", safeMessage, string(contextJSON)))
}