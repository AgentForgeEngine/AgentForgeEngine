package templates

import (
	"fmt"
	"strings"
)

// Qwen3Template handles the specific qwen3 template format
type Qwen3Template struct {
	template string
}

// NewQwen3Template creates a new qwen3 template processor
func NewQwen3Template(templateContent string) *Qwen3Template {
	return &Qwen3Template{template: templateContent}
}

// Render processes the qwen3 template with message data
func (qt *Qwen3Template) Render(messages []map[string]interface{}) (string, error) {
	content := qt.template

	// Extract system prompt and mode
	systemPrompt := ""
	mode := "plan"

	for _, msg := range messages {
		if role, ok := msg["role"].(string); ok && role == "system" {
			if content, ok := msg["content"].(string); ok {
				systemPrompt = content

				// Extract <mode>...</mode> if present
				if strings.Contains(content, "<mode>") && strings.Contains(content, "</mode>") {
					start := strings.Index(content, "<mode>") + 6
					end := strings.Index(content, "</mode>")
					if start > 5 && end > start {
						mode = strings.TrimSpace(content[start:end])
					}
				}
				break
			}
		}
	}

	// Replace mode placeholder
	content = strings.ReplaceAll(content, "{{ mode | upper }}", strings.ToUpper(mode))
	content = strings.ReplaceAll(content, "{{ mode }}", mode)

	// Replace system prompt placeholder
	content = strings.ReplaceAll(content, "{{ system_prompt }}", systemPrompt)

	// Process message history
	var messageHistory strings.Builder

	for _, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)

		if role == "user" {
			messageHistory.WriteString("<|im_start|>user\n")
			messageHistory.WriteString(content)
			messageHistory.WriteString("\n<|im_end|>\n\n")
		} else if role == "assistant" {
			messageHistory.WriteString("<|im_start|>assistant\n")

			// Check for function call
			if fc, ok := msg["function_call"].(map[string]interface{}); ok {
				name, _ := fc["name"].(string)
				args, _ := fc["arguments"].(string)
				messageHistory.WriteString(fmt.Sprintf("<function_call name=\"%s\">\n%s\n</function_call>", name, args))
			} else {
				messageHistory.WriteString(content)
			}

			messageHistory.WriteString("\n<|im_end|>\n\n")
		}
	}

	// Replace message loop with actual messages
	start := strings.Index(content, "{% for msg in messages %}")
	end := strings.Index(content, "{% endfor %}")

	if start != -1 && end != -1 {
		before := content[:start]
		after := content[end+13:] // Length of "{% endfor %}"
		content = before + messageHistory.String() + after
	}

	// Handle conditional blocks
	if mode == "plan" {
		// Keep plan mode rules
		content = strings.ReplaceAll(content, "{% if mode == \"plan\" %}", "")
		content = strings.ReplaceAll(content, "{% else %}", "REMOVE")
		content = strings.ReplaceAll(content, "{% endif %}", "")

		// Remove the else block content
		lines := strings.Split(content, "\n")
		var result []string
		skip := false

		for _, line := range lines {
			if strings.Contains(line, "REMOVE") {
				skip = true
				continue
			}
			if skip && strings.Contains(line, "- You may use read/write/execute Linux tools.") {
				skip = false
				continue
			}
			if !skip {
				result = append(result, line)
			}
		}
		content = strings.Join(result, "\n")
	} else {
		// Keep execute mode rules
		content = strings.ReplaceAll(content, "{% if mode == \"plan\" %}", "REMOVE")
		content = strings.ReplaceAll(content, "{% else %}", "")
		content = strings.ReplaceAll(content, "{% endif %}", "")

		// Remove the if block content
		lines := strings.Split(content, "\n")
		var result []string
		skip := false

		for _, line := range lines {
			if strings.Contains(line, "REMOVE") {
				skip = true
				continue
			}
			if skip && strings.Contains(line, "- You may use read/write/execute Linux tools.") {
				skip = false
				continue
			}
			if !skip {
				result = append(result, line)
			}
		}
		content = strings.Join(result, "\n")
	}

	return content, nil
}
