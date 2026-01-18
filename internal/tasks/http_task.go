// Package tasks provides concrete implementations of TaskExecutor for various workflow operations.
// Each task type (HTTP request, transform, HTML parser, etc.) is implemented as a separate executor.
package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
)

// HTTPTask executes HTTP requests with support for dynamic body interpolation from ExecutionContext.
// It supports GET, POST, PUT, DELETE, and PATCH methods with custom headers and request bodies.
type HTTPTask struct{}

// Execute performs an HTTP request based on the provided configuration.
// Configuration fields:
//   - method (string, required): HTTP method (GET, POST, PUT, DELETE, PATCH)
//   - url (string, required): Target URL
//   - headers (map[string]interface{}, optional): HTTP headers
//   - body (string, optional): Request body with template support for context interpolation
//   - timeout (int, optional): Request timeout in seconds (default: 30)
//
// The response is returned in TaskResult.Output with the following structure:
//   - status_code (int): HTTP status code
//   - headers (map[string][]string): Response headers
//   - body (interface{}): Parsed JSON response body (or raw string if not JSON)
func (h *HTTPTask) Execute(ctx *engine.ExecutionContext, config map[string]interface{}) engine.TaskResult {
	// Validate required configuration
	method, ok := config["method"].(string)
	if !ok || method == "" {
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  "missing or invalid 'method' in configuration",
		}
	}

	url, ok := config["url"].(string)
	if !ok || url == "" {
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  "missing or invalid 'url' in configuration",
		}
	}

	// Get optional timeout (default 30s)
	timeout := 30
	if t, ok := config["timeout"].(int); ok {
		timeout = t
	} else if t, ok := config["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Get optional body
	bodyStr := ""
	if b, ok := config["body"].(string); ok {
		bodyStr = b
	}

	// Apply body interpolation if body is provided
	if bodyStr != "" {
		interpolated, err := h.interpolateBody(bodyStr, ctx)
		if err != nil {
			slog.Error("Failed to interpolate body", "error", err)
			return engine.TaskResult{
				Status: "failed",
				Output: nil,
				Error:  fmt.Sprintf("body interpolation failed: %v", err),
			}
		}
		bodyStr = interpolated
	}

	// Build HTTP request
	var bodyReader io.Reader
	if bodyStr != "" {
		bodyReader = strings.NewReader(bodyStr)
	}

	req, err := http.NewRequest(strings.ToUpper(method), url, bodyReader)
	if err != nil {
		slog.Error("Failed to create HTTP request", "error", err)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("failed to create request: %v", err),
		}
	}

	// Set headers
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Set default Content-Type for POST/PUT/PATCH with body
	if bodyStr != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	slog.Info("Executing HTTP request", "method", method, "url", url)
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("HTTP request failed", "error", err)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("request execution failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("failed to read response: %v", err),
		}
	}

	// Check for HTTP error status codes
	if resp.StatusCode >= 400 {
		slog.Warn("HTTP request returned error status", "status_code", resp.StatusCode)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	// Parse JSON response if Content-Type is application/json
	var parsedBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// If JSON parsing fails, return raw string
			parsedBody = string(respBody)
		}
	} else {
		parsedBody = string(respBody)
	}

	// Build output with response details
	output := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        parsedBody,
	}

	slog.Info("HTTP request completed successfully", "status_code", resp.StatusCode)
	return engine.TaskResult{
		Status: "success",
		Output: output,
		Error:  "",
	}
}

// interpolateBody replaces template variables in the body string with values from ExecutionContext.
// Template syntax: {{context.key}} where 'key' is a key in the ExecutionContext.
func (h *HTTPTask) interpolateBody(bodyTemplate string, ctx *engine.ExecutionContext) (string, error) {
	// Create template with context data
	tmpl, err := template.New("body").Parse(bodyTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Get all context data
	contextData := ctx.GetAll()

	// Create template data structure
	data := map[string]interface{}{
		"context": contextData,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RegisterHTTPTask registers the HTTP task executor with the provided registry.
// The task is registered with the type name "http_request".
func RegisterHTTPTask(registry *engine.Registry) {
	registry.Register("http_request", &HTTPTask{})
	slog.Info("Registered HTTP task executor", "type", "http_request")
}
