// Package tasks provides concrete implementations of TaskExecutor for various workflow operations.
// Each task type (HTTP request, transform, HTML parser, etc.) is implemented as a separate executor.
package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
)

// TransformTask implements TaskExecutor for data transformation using Go templates.
// It supports reshaping data structures, extracting specific fields, and combining data
// from multiple sources in the ExecutionContext.
type TransformTask struct {
	funcMap template.FuncMap
}

// NewTransformTask creates a new Transform task executor with custom template functions.
func NewTransformTask() *TransformTask {
	return &TransformTask{
		funcMap: createTemplateFuncMap(),
	}
}

// Execute implements the TaskExecutor interface for data transformation.
// Configuration fields:
//   - template (string, required): Go template string defining transformation
//   - data_source (string, optional): Specific ExecutionContext key to use as input. If omitted, entire context available.
//   - output_format (string, optional): "json" or "string" (default: "json")
//
// The transformed data is returned in TaskResult.Output.
func (t *TransformTask) Execute(ctx *engine.ExecutionContext, config map[string]interface{}) engine.TaskResult {
	// Extract and validate configuration
	templateStr, ok := config["template"].(string)
	if !ok || templateStr == "" {
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  "missing or invalid 'template' in configuration",
		}
	}

	// Get data source (optional - defaults to all context)
	var inputData interface{}
	if dataSource, ok := config["data_source"].(string); ok && dataSource != "" {
		// Load specific key from context
		data, exists := ctx.Get(dataSource)
		if !exists {
			slog.Warn("Data source not found in context", "source", dataSource)
			inputData = map[string]interface{}{} // Empty map if not found
		} else {
			inputData = data
		}
	} else {
		// Use entire context
		inputData = ctx.GetAll()
	}

	// Get output format (optional - defaults to "json")
	outputFormat := "json"
	if format, ok := config["output_format"].(string); ok {
		outputFormat = format
	}

	slog.Info("Executing transform", "has_data_source", config["data_source"] != nil, "output_format", outputFormat)

	// Parse and execute template
	tmpl, err := template.New("transform").Funcs(t.funcMap).Parse(templateStr)
	if err != nil {
		slog.Error("Template parsing failed", "error", err)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("failed to parse template: %v", err),
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, inputData); err != nil {
		slog.Error("Template execution failed", "error", err)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("failed to execute template: %v", err),
		}
	}

	result := buf.String()

	// Format output
	var output interface{}
	if outputFormat == "json" {
		// Try to parse result as JSON
		var jsonOutput interface{}
		if err := json.Unmarshal([]byte(result), &jsonOutput); err != nil {
			// If not valid JSON, return as string
			slog.Warn("Template output is not valid JSON, returning as string", "error", err)
			output = result
		} else {
			output = jsonOutput
		}
	} else {
		// Return as string
		output = result
	}

	slog.Info("Transform completed successfully")

	return engine.TaskResult{
		Status: "success",
		Output: output,
		Error:  "",
	}
}

// createTemplateFuncMap creates custom template functions for data transformation.
// Available functions:
//   - toUpper: Convert string to uppercase
//   - toLower: Convert string to lowercase
//   - trim: Trim whitespace from string
//   - join: Join array elements with separator
//   - toJSON: Convert value to JSON string
//   - default: Provide default value if nil/empty
func createTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"toUpper": strings.ToUpper,
		"toLower": strings.ToLower,
		"trim":    strings.TrimSpace,
		"join": func(sep string, items []interface{}) string {
			strItems := make([]string, len(items))
			for i, item := range items {
				strItems[i] = fmt.Sprint(item)
			}
			return strings.Join(strItems, sep)
		},

		// JSON functions
		"toJSON": func(v interface{}) (string, error) {
			data, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(data), nil
		},

		// Utility functions
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}
}

// RegisterTransformTask registers the Transform task executor with the provided registry.
// The task is registered with the type name "transform".
func RegisterTransformTask(registry *engine.Registry) {
	registry.Register("transform", NewTransformTask())
	slog.Info("Registered Transform task executor", "type", "transform")
}
