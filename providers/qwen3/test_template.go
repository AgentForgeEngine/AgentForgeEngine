package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AgentForgeEngine/AgentForgeEngine/pkg/templates"
)

func main() {
	// Test template processing
	templateContent := `{% for msg in messages %}{{ msg.role }}: {{ msg.content }}{% endfor %}`

	tmpl, err := templates.NewJinja2Template(templateContent)
	if err != nil {
		log.Fatalf("Failed to create template: %v", err)
	}

	// Test data
	data := map[string]interface{}{
		"messages": []map[string]interface{}{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Hello!"},
		},
	}

	result, err := tmpl.Render(data)
	if err != nil {
		log.Fatalf("Failed to render template: %v", err)
	}

	fmt.Printf("Template result: %s\n", result)

	// Test qwen3 template
	qwen3Template, err := templates.FindTemplate("qwen3")
	if err != nil {
		log.Printf("Warning: Could not find qwen3 template: %v", err)
	} else {
		log.Printf("Found qwen3 template at: %s", qwen3Template)

		content, err := os.ReadFile(qwen3Template)
		if err != nil {
			log.Printf("Failed to read qwen3 template file: %v", err)
		} else {
			qwen3Tmpl, err := templates.NewJinja2Template(string(content))
			if err != nil {
				log.Printf("Failed to load qwen3 template: %v", err)
			} else {
				// Test with sample messages
				messages := []map[string]interface{}{
					{"role": "system", "content": "You are operating in <mode>plan</mode> mode."},
					{"role": "user", "content": "List files in current directory"},
				}

				result, err := qwen3Tmpl.Render(map[string]interface{}{"messages": messages})
				if err != nil {
					log.Printf("Failed to render qwen3 template: %v", err)
				} else {
					fmt.Printf("\nQwen3 template preview (first 200 chars):\n%s...\n", result[:200])
				}
			}
		}
	}
}
