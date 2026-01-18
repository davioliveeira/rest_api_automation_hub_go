// Package tasks provides concrete implementations of TaskExecutor for various workflow operations.
// Each task type (HTTP request, transform, HTML parser, etc.) is implemented as a separate executor.
package tasks

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
)

// HTMLParserTask implements TaskExecutor for HTML parsing using CSS selectors.
// It extracts structured data from HTML content stored in ExecutionContext.
type HTMLParserTask struct{}

// SelectorConfig defines how to extract data using a CSS selector.
type SelectorConfig struct {
	Name      string // Field name in output map
	Selector  string // CSS selector
	Attribute string // Optional: attribute to extract (e.g., "href", "src")
	Multiple  bool   // Optional: extract all matches vs first match
}

// Execute implements the TaskExecutor interface for HTML parsing.
// Configuration fields:
//   - html_source (string, required): ExecutionContext key containing HTML content
//   - selectors ([]map[string]interface{}, required): Array of selector configurations
//
// Each selector configuration:
//   - name (string, required): Output field name
//   - selector (string, required): CSS selector
//   - attribute (string, optional): Attribute to extract instead of text
//   - multiple (bool, optional): Extract all matches (default: false, first match only)
//
// Returns extracted data as []map[string]any per AC1.
func (h *HTMLParserTask) Execute(ctx *engine.ExecutionContext, config map[string]interface{}) engine.TaskResult {
	// Validate html_source
	htmlSource, ok := config["html_source"].(string)
	if !ok || htmlSource == "" {
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  "missing or invalid 'html_source' in configuration",
		}
	}

	// Validate selectors
	selectorsConfig, ok := config["selectors"].([]interface{})
	if !ok || len(selectorsConfig) == 0 {
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  "missing or invalid 'selectors' in configuration",
		}
	}

	// Load HTML content from context
	htmlContent, exists := ctx.Get(htmlSource)
	if !exists {
		slog.Warn("HTML source not found in context", "source", htmlSource)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("HTML source '%s' not found in context", htmlSource),
		}
	}

	// Convert to string
	htmlString, ok := htmlContent.(string)
	if !ok {
		// Try to extract from nested structure (e.g., HTTP response body)
		if htmlMap, ok := htmlContent.(map[string]interface{}); ok {
			if body, exists := htmlMap["body"]; exists {
				htmlString, ok = body.(string)
				if !ok {
					return engine.TaskResult{
						Status: "failed",
						Output: nil,
						Error:  "HTML content is not a string",
					}
				}
			} else {
				return engine.TaskResult{
					Status: "failed",
					Output: nil,
					Error:  "HTML content is not a string",
				}
			}
		} else {
			return engine.TaskResult{
				Status: "failed",
				Output: nil,
				Error:  "HTML content is not a string",
			}
		}
	}

	// Parse HTML with Goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlString))
	if err != nil {
		slog.Error("Failed to parse HTML", "error", err)
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("failed to parse HTML: %v", err),
		}
	}

	// Parse selector configurations
	selectors, err := h.parseSelectors(selectorsConfig)
	if err != nil {
		return engine.TaskResult{
			Status: "failed",
			Output: nil,
			Error:  fmt.Sprintf("invalid selector configuration: %v", err),
		}
	}

	slog.Info("Executing HTML parser", "html_source", htmlSource, "selector_count", len(selectors))

	// Extract data using selectors
	results := h.extractData(doc, selectors)

	slog.Info("HTML parsing completed successfully", "results_count", len(results))

	return engine.TaskResult{
		Status: "success",
		Output: results,
		Error:  "",
	}
}

// parseSelectors converts raw config to SelectorConfig structs.
func (h *HTMLParserTask) parseSelectors(selectorsConfig []interface{}) ([]SelectorConfig, error) {
	selectors := make([]SelectorConfig, 0, len(selectorsConfig))

	for i, sel := range selectorsConfig {
		selMap, ok := sel.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("selector at index %d is not a map", i)
		}

		// Required: name
		name, ok := selMap["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("selector at index %d missing 'name'", i)
		}

		// Required: selector
		selector, ok := selMap["selector"].(string)
		if !ok || selector == "" {
			return nil, fmt.Errorf("selector at index %d missing 'selector'", i)
		}

		// Optional: attribute
		attribute, _ := selMap["attribute"].(string)

		// Optional: multiple
		multiple, _ := selMap["multiple"].(bool)

		selectors = append(selectors, SelectorConfig{
			Name:      name,
			Selector:  selector,
			Attribute: attribute,
			Multiple:  multiple,
		})
	}

	return selectors, nil
}

// extractData applies all selectors to the document and returns extracted data.
// Returns []map[string]any per AC1.
func (h *HTMLParserTask) extractData(doc *goquery.Document, selectors []SelectorConfig) []map[string]any {
	// Start with a single result map
	resultMap := make(map[string]any)

	for _, sel := range selectors {
		selection := doc.Find(sel.Selector)

		if sel.Multiple {
			// Extract all matches
			values := []string{}
			selection.Each(func(i int, s *goquery.Selection) {
				value := h.extractValue(s, sel.Attribute)
				if value != "" {
					values = append(values, value)
				}
			})

			if len(values) == 0 {
				slog.Warn("CSS selector returned no results", "selector", sel.Selector, "name", sel.Name)
			}

			resultMap[sel.Name] = values
		} else {
			// Extract first match only
			value := h.extractValue(selection.First(), sel.Attribute)

			if value == "" {
				slog.Warn("CSS selector returned no results", "selector", sel.Selector, "name", sel.Name)
			}

			resultMap[sel.Name] = value
		}
	}

	// Return as []map[string]any per AC1
	return []map[string]any{resultMap}
}

// extractValue extracts either text content or an attribute from a selection.
func (h *HTMLParserTask) extractValue(s *goquery.Selection, attribute string) string {
	if attribute != "" {
		// Extract attribute value
		value, exists := s.Attr(attribute)
		if !exists {
			return ""
		}
		return strings.TrimSpace(value)
	}

	// Extract text content (default)
	return strings.TrimSpace(s.Text())
}

// RegisterHTMLParserTask registers the HTML Parser task executor with the provided registry.
// The task is registered with the type name "html_parser".
func RegisterHTMLParserTask(registry *engine.Registry) {
	registry.Register("html_parser", &HTMLParserTask{})
	slog.Info("Registered HTML Parser task executor", "type", "html_parser")
}
