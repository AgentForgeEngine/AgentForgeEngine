package response

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/interfaces"
)

// Formatter handles converting AgentOutput to function_response format
type Formatter interface {
	FormatAgentOutput(agentName string, output interfaces.AgentOutput) (string, error)
}

// XMLFormatter converts AgentOutput to XML-like function_response format
type XMLFormatter struct{}

// NewXMLFormatter creates a new XML formatter
func NewXMLFormatter() *XMLFormatter {
	return &XMLFormatter{}
}

// FormatAgentOutput converts AgentOutput to function_response XML format
func (xf *XMLFormatter) FormatAgentOutput(agentName string, output interfaces.AgentOutput) (string, error) {
	if output.Success {
		// Convert data to JSON
		argsJSON, err := json.Marshal(output.Data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal arguments: %w", err)
		}

		return fmt.Sprintf(`<function_response name="%s">%s</function_response>`, agentName, string(argsJSON)), nil
	} else {
		// Return error format
		return fmt.Sprintf(`<function_response name="%s"><error>%s</error></function_response>`, agentName, output.Error), nil
	}
}

// ValidateFunctionResponse validates the format of function responses
func (xf *XMLFormatter) ValidateFunctionResponse(response string, expectedName string) error {
	// Check opening tag
	expectedOpening := fmt.Sprintf(`<function_response name="%s">`, expectedName)
	if !strings.Contains(response, expectedOpening) {
		return fmt.Errorf("invalid opening tag in response: %s", response)
	}

	// Check closing tag
	if !strings.Contains(response, `</function_response>`) {
		return fmt.Errorf("missing closing tag in response: %s", response)
	}

	// Extract and validate JSON content
	start := strings.Index(response, ">") + 1
	end := strings.Index(response, "</function_response>")
	if start == -1 || end == -1 || start >= end {
		return fmt.Errorf("malformed response content: %s", response)
	}

	jsonContent := response[start:end]
	var testJSON map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &testJSON); err != nil {
		return fmt.Errorf("invalid JSON in response: %v", err)
	}

	return nil
}

// JSONFormatter converts AgentOutput to function_response JSON format
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// FormatAgentOutput converts AgentOutput to function_response JSON format
func (jf *JSONFormatter) FormatAgentOutput(agentName string, output interfaces.AgentOutput) (string, error) {
	response := map[string]interface{}{
		"function_response": map[string]interface{}{
			"name":      agentName,
			"arguments": output.Data,
			"success":   output.Success,
		},
	}

	if !output.Success && output.Error != "" {
		response["function_response"].(map[string]interface{})["error"] = output.Error
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonBytes), nil
}

// AutoFormatter detects and uses appropriate formatter based on content
type AutoFormatter struct {
	xmlFormatter  *XMLFormatter
	jsonFormatter *JSONFormatter
}

// NewAutoFormatter creates a new auto formatter
func NewAutoFormatter() *AutoFormatter {
	return &AutoFormatter{
		xmlFormatter:  NewXMLFormatter(),
		jsonFormatter: NewJSONFormatter(),
	}
}

// FormatAgentOutput automatically detects format and converts appropriately
func (af *AutoFormatter) FormatAgentOutput(agentName string, output interfaces.AgentOutput) (string, error) {
	// Check if output data looks like it should be XML (has files array, etc.)
	if af.shouldUseXMLFormat(output.Data) {
		return af.xmlFormatter.FormatAgentOutput(agentName, output)
	} else {
		return af.jsonFormatter.FormatAgentOutput(agentName, output)
	}
}

// shouldUseXMLFormat heuristically determines if XML format is appropriate
func (af *AutoFormatter) shouldUseXMLFormat(data map[string]interface{}) bool {
	// If output contains arrays or complex nested structures, use XML
	for _, value := range data {
		switch v := value.(type) {
		case []interface{}:
			return true // Arrays suggest XML format
		case map[string]interface{}:
			// Check if map has complex nested structure
			if len(v) > 3 {
				return true
			}
		}
	}
	return false
}
